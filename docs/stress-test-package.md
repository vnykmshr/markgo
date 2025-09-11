

TYPES

type ErrorInfo struct {
	URL       string    `json:"url"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Depth     int       `json:"depth"`
}

type PerformanceTarget struct {
	Name        string
	Target      string
	Actual      string
	Description string
	Passed      bool
}
    PerformanceTarget represents a performance validation target

type PerformanceValidator struct {
	// Has unexported fields.
}
    PerformanceValidator validates test results against performance targets

func NewPerformanceValidator() *PerformanceValidator
    NewPerformanceValidator creates a new performance validator

func (pv *PerformanceValidator) GetValidationSummary() map[string]interface{}
    GetValidationSummary returns a summary of the validation results

func (pv *PerformanceValidator) PrintValidationReport()
    PrintValidationReport prints a detailed performance validation report

func (pv *PerformanceValidator) ValidateResults(results *TestResults)
    ValidateResults validates the test results against performance targets

type ReportGenerator struct {
	// Has unexported fields.
}

func NewReportGenerator(results *TestResults) *ReportGenerator

func (rg *ReportGenerator) GenerateHTMLReport(outputPath string) error

func (rg *ReportGenerator) GenerateJSONReport(outputPath string) error

type ResponseTimeEntry struct {
	URL          string        `json:"url"`
	ResponseTime time.Duration `json:"response_time"`
	Timestamp    time.Time     `json:"timestamp"`
}

type SlowRequest struct {
	URL          string        `json:"url"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
}

type StressTestConfig struct {
	BaseURL        string  `json:"base_url"`
	Concurrency    int     `json:"concurrency"`
	Duration       string  `json:"duration"`
	Timeout        string  `json:"timeout"`
	UserAgent      string  `json:"user_agent"`
	FollowLinks    bool    `json:"follow_links"`
	MaxDepth       int     `json:"max_depth"`
	OutputFile     string  `json:"output_file"`
	Verbose        bool    `json:"verbose"`
	RequestsPerSec float64 `json:"requests_per_sec"`
}

type StressTester struct {
	// Has unexported fields.
}

func NewStressTester(config StressTesterConfig) *StressTester

func (st *StressTester) Run(ctx context.Context) (*TestResults, error)

type StressTesterConfig struct {
	BaseURL        string
	Concurrency    int
	RequestTimeout time.Duration
	UserAgent      string
	FollowLinks    bool
	MaxDepth       int
	RequestsPerSec float64
	Logger         *slog.Logger
}

type TestResults struct {
	Duration              string                 `json:"duration"`
	URLsDiscovered        int                    `json:"urls_discovered"`
	TotalRequests         int64                  `json:"total_requests"`
	SuccessfulRequests    int64                  `json:"successful_requests"`
	FailedRequests        int64                  `json:"failed_requests"`
	AverageResponseTime   string                 `json:"average_response_time"`
	MinResponseTime       string                 `json:"min_response_time"`
	MaxResponseTime       string                 `json:"max_response_time"`
	RequestsPerSecond     float64                `json:"requests_per_second"`
	SuccessRate           float64                `json:"success_rate"`
	URLValidations        []URLValidation        `json:"url_validations"`
	Errors                []ErrorInfo            `json:"errors"`
	SlowRequests          []SlowRequest          `json:"slow_requests"`
	ResponseTimes         []ResponseTimeEntry    `json:"response_times"`
	PerformanceValidation map[string]interface{} `json:"performance_validation,omitempty"`
}

type URLTask struct {
	URL   string
	Depth int
}

type URLValidation struct {
	URL           string        `json:"url"`
	StatusCode    int           `json:"status_code"`
	ResponseTime  time.Duration `json:"response_time"`
	ContentLength int64         `json:"content_length"`
	ContentType   string        `json:"content_type"`
	LinksFound    int           `json:"links_found"`
	Depth         int           `json:"depth"`
	Error         string        `json:"error,omitempty"`
	IsValid       bool          `json:"is_valid"`
}

