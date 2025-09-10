package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type StressTesterConfig struct {
	BaseURL        string
	Concurrency    int
	RequestTimeout time.Duration
	UserAgent      string
	FollowLinks    bool
	MaxDepth       int
	Logger         *slog.Logger
}

type StressTester struct {
	config       StressTesterConfig
	client       *http.Client
	urlQueue     chan URLTask
	results      *TestResults
	discoveredURLs sync.Map
	urlPattern   *regexp.Regexp
	baseURLParsed *url.URL
}

type URLTask struct {
	URL   string
	Depth int
}

type TestResults struct {
	Duration            string                `json:"duration"`
	URLsDiscovered      int                   `json:"urls_discovered"`
	TotalRequests       int64                 `json:"total_requests"`
	SuccessfulRequests  int64                 `json:"successful_requests"`
	FailedRequests      int64                 `json:"failed_requests"`
	AverageResponseTime string                `json:"average_response_time"`
	MinResponseTime     string                `json:"min_response_time"`
	MaxResponseTime     string                `json:"max_response_time"`
	RequestsPerSecond   float64               `json:"requests_per_second"`
	SuccessRate         float64               `json:"success_rate"`
	URLValidations      []URLValidation       `json:"url_validations"`
	Errors              []ErrorInfo           `json:"errors"`
	SlowRequests        []SlowRequest         `json:"slow_requests"`
	ResponseTimes       []ResponseTimeEntry   `json:"response_times"`
}

type URLValidation struct {
	URL            string        `json:"url"`
	StatusCode     int           `json:"status_code"`
	ResponseTime   time.Duration `json:"response_time"`
	ContentLength  int64         `json:"content_length"`
	ContentType    string        `json:"content_type"`
	LinksFound     int           `json:"links_found"`
	Depth          int           `json:"depth"`
	Error          string        `json:"error,omitempty"`
	IsValid        bool          `json:"is_valid"`
}

type ErrorInfo struct {
	URL       string    `json:"url"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Depth     int       `json:"depth"`
}

type SlowRequest struct {
	URL          string        `json:"url"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
}

type ResponseTimeEntry struct {
	URL          string        `json:"url"`
	ResponseTime time.Duration `json:"response_time"`
	Timestamp    time.Time     `json:"timestamp"`
}

func NewStressTester(config StressTesterConfig) *StressTester {
	// Parse base URL
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		panic(fmt.Sprintf("Invalid base URL: %v", err))
	}

	return &StressTester{
		config:        config,
		client:        &http.Client{Timeout: config.RequestTimeout},
		urlQueue:      make(chan URLTask, 10000),
		results:       &TestResults{URLValidations: make([]URLValidation, 0)},
		urlPattern:    regexp.MustCompile(`href=["']([^"']+)["']`),
		baseURLParsed: baseURL,
	}
}

func (st *StressTester) Run(ctx context.Context) (*TestResults, error) {
	startTime := time.Now()
	
	// Initialize results
	st.results.ResponseTimes = make([]ResponseTimeEntry, 0)
	st.results.Errors = make([]ErrorInfo, 0)
	st.results.SlowRequests = make([]SlowRequest, 0)
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < st.config.Concurrency; i++ {
		wg.Add(1)
		go st.worker(ctx, &wg)
	}
	
	// Start URL discovery with the base URL
	st.addURL(st.config.BaseURL, 0)
	
	// Start monitoring
	go st.monitor(ctx)
	
	// Wait for context cancellation or completion
	<-ctx.Done()
	
	// Close URL queue and wait for workers to finish
	close(st.urlQueue)
	wg.Wait()
	
	// Calculate final results
	st.calculateResults(time.Since(startTime))
	
	return st.results, nil
}

func (st *StressTester) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for {
		select {
		case task, ok := <-st.urlQueue:
			if !ok {
				return
			}
			st.processURL(ctx, task)
		case <-ctx.Done():
			return
		}
	}
}

func (st *StressTester) processURL(ctx context.Context, task URLTask) {
	startTime := time.Now()
	atomic.AddInt64(&st.results.TotalRequests, 1)
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", task.URL, nil)
	if err != nil {
		st.recordError(task.URL, fmt.Sprintf("creating request: %v", err), task.Depth)
		atomic.AddInt64(&st.results.FailedRequests, 1)
		return
	}
	
	// Set user agent
	req.Header.Set("User-Agent", st.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	
	// Make request
	resp, err := st.client.Do(req)
	if err != nil {
		st.recordError(task.URL, fmt.Sprintf("making request: %v", err), task.Depth)
		atomic.AddInt64(&st.results.FailedRequests, 1)
		return
	}
	defer resp.Body.Close()
	
	responseTime := time.Since(startTime)
	atomic.AddInt64(&st.results.SuccessfulRequests, 1)
	
	// Record response time
	st.recordResponseTime(task.URL, responseTime)
	
	// Create validation record
	validation := URLValidation{
		URL:           task.URL,
		StatusCode:    resp.StatusCode,
		ResponseTime:  responseTime,
		ContentLength: resp.ContentLength,
		ContentType:   resp.Header.Get("Content-Type"),
		Depth:         task.Depth,
		IsValid:       resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
	
	// Process response body for link discovery
	if st.config.FollowLinks && task.Depth < st.config.MaxDepth && 
		strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		
		body := make([]byte, 0, 64*1024) // Limit to 64KB for link extraction
		buffer := make([]byte, 4096)
		totalRead := 0
		
		for totalRead < 64*1024 {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				if totalRead+n > 64*1024 {
					body = append(body, buffer[:64*1024-totalRead]...)
					break
				}
				body = append(body, buffer[:n]...)
				totalRead += n
			}
			if err != nil {
				break
			}
		}
		
		links := st.extractLinks(string(body), task.URL)
		validation.LinksFound = len(links)
		
		// Add new URLs to queue
		for _, link := range links {
			st.addURL(link, task.Depth+1)
		}
	}
	
	// Record slow requests (>2 seconds)
	if responseTime > 2*time.Second {
		st.recordSlowRequest(task.URL, responseTime, resp.StatusCode)
	}
	
	// Add validation to results (thread-safe)
	st.addValidation(validation)
	
	st.config.Logger.Debug("URL processed",
		"url", task.URL,
		"status", resp.StatusCode,
		"response_time", responseTime,
		"depth", task.Depth,
		"links_found", validation.LinksFound)
}

func (st *StressTester) addURL(rawURL string, depth int) {
	// Parse and validate URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	
	// Make relative URLs absolute
	if !parsedURL.IsAbs() {
		parsedURL = st.baseURLParsed.ResolveReference(parsedURL)
	}
	
	// Only process URLs from the same host
	if parsedURL.Host != st.baseURLParsed.Host {
		return
	}
	
	// Clean URL (remove fragment, normalize)
	parsedURL.Fragment = ""
	cleanURL := parsedURL.String()
	
	// Check if already discovered
	if _, exists := st.discoveredURLs.LoadOrStore(cleanURL, true); exists {
		return
	}
	
	// Add to queue
	select {
	case st.urlQueue <- URLTask{URL: cleanURL, Depth: depth}:
		st.results.URLsDiscovered++
	default:
		// Queue full, skip
	}
}

func (st *StressTester) extractLinks(body, baseURL string) []string {
	matches := st.urlPattern.FindAllStringSubmatch(body, -1)
	links := make([]string, 0, len(matches))
	
	for _, match := range matches {
		if len(match) > 1 {
			link := strings.TrimSpace(match[1])
			if link != "" && !strings.HasPrefix(link, "javascript:") && 
				!strings.HasPrefix(link, "mailto:") && !strings.HasPrefix(link, "#") {
				links = append(links, link)
			}
		}
	}
	
	return links
}

func (st *StressTester) recordError(url, error string, depth int) {
	errorInfo := ErrorInfo{
		URL:       url,
		Error:     error,
		Timestamp: time.Now(),
		Depth:     depth,
	}
	
	// Thread-safe append
	st.addError(errorInfo)
}

func (st *StressTester) recordResponseTime(url string, responseTime time.Duration) {
	entry := ResponseTimeEntry{
		URL:          url,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
	}
	
	// Thread-safe append
	st.addResponseTime(entry)
}

func (st *StressTester) recordSlowRequest(url string, responseTime time.Duration, statusCode int) {
	slowReq := SlowRequest{
		URL:          url,
		ResponseTime: responseTime,
		StatusCode:   statusCode,
	}
	
	// Thread-safe append
	st.addSlowRequest(slowReq)
}

// Thread-safe methods for adding to slices
var (
	validationsMutex    sync.Mutex
	errorsMutex         sync.Mutex
	responseTimesMutex  sync.Mutex
	slowRequestsMutex   sync.Mutex
)

func (st *StressTester) addValidation(validation URLValidation) {
	validationsMutex.Lock()
	defer validationsMutex.Unlock()
	st.results.URLValidations = append(st.results.URLValidations, validation)
}

func (st *StressTester) addError(error ErrorInfo) {
	errorsMutex.Lock()
	defer errorsMutex.Unlock()
	st.results.Errors = append(st.results.Errors, error)
}

func (st *StressTester) addResponseTime(entry ResponseTimeEntry) {
	responseTimesMutex.Lock()
	defer responseTimesMutex.Unlock()
	st.results.ResponseTimes = append(st.results.ResponseTimes, entry)
}

func (st *StressTester) addSlowRequest(req SlowRequest) {
	slowRequestsMutex.Lock()
	defer slowRequestsMutex.Unlock()
	st.results.SlowRequests = append(st.results.SlowRequests, req)
}

func (st *StressTester) monitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			total := atomic.LoadInt64(&st.results.TotalRequests)
			successful := atomic.LoadInt64(&st.results.SuccessfulRequests)
			failed := atomic.LoadInt64(&st.results.FailedRequests)
			discovered := st.results.URLsDiscovered
			
			st.config.Logger.Info("Progress update",
				"total_requests", total,
				"successful_requests", successful,
				"failed_requests", failed,
				"urls_discovered", discovered,
				"queue_size", len(st.urlQueue))
		case <-ctx.Done():
			return
		}
	}
}

func (st *StressTester) calculateResults(duration time.Duration) {
	st.results.Duration = duration.String()
	
	// Calculate response time statistics
	responseTimesMutex.Lock()
	responseTimes := make([]time.Duration, len(st.results.ResponseTimes))
	for i, entry := range st.results.ResponseTimes {
		responseTimes[i] = entry.ResponseTime
	}
	responseTimesMutex.Unlock()
	
	if len(responseTimes) > 0 {
		sort.Slice(responseTimes, func(i, j int) bool {
			return responseTimes[i] < responseTimes[j]
		})
		
		st.results.MinResponseTime = responseTimes[0].String()
		st.results.MaxResponseTime = responseTimes[len(responseTimes)-1].String()
		
		// Calculate average
		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		st.results.AverageResponseTime = (total / time.Duration(len(responseTimes))).String()
	}
	
	// Calculate rates
	if duration.Seconds() > 0 {
		st.results.RequestsPerSecond = float64(st.results.TotalRequests) / duration.Seconds()
	}
	
	if st.results.TotalRequests > 0 {
		st.results.SuccessRate = (float64(st.results.SuccessfulRequests) / float64(st.results.TotalRequests)) * 100
	}
	
	// Sort slow requests by response time
	slowRequestsMutex.Lock()
	sort.Slice(st.results.SlowRequests, func(i, j int) bool {
		return st.results.SlowRequests[i].ResponseTime > st.results.SlowRequests[j].ResponseTime
	})
	slowRequestsMutex.Unlock()
}