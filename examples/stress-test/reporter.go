package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"sort"
	"time"
)

type ReportGenerator struct {
	results *TestResults
}

func NewReportGenerator(results *TestResults) *ReportGenerator {
	return &ReportGenerator{results: results}
}

func (rg *ReportGenerator) GenerateHTMLReport(outputPath string) error {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MarkGo Stress Test Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 12px; margin-bottom: 30px; box-shadow: 0 10px 30px rgba(0,0,0,0.1); }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { font-size: 1.1em; opacity: 0.9; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: white; padding: 25px; border-radius: 12px; box-shadow: 0 5px 15px rgba(0,0,0,0.08); border-left: 4px solid #667eea; }
        .stat-card h3 { color: #667eea; font-size: 0.9em; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 10px; }
        .stat-card .value { font-size: 2.2em; font-weight: bold; color: #2d3748; }
        .stat-card .unit { font-size: 0.9em; color: #718096; margin-left: 5px; }
        .section { background: white; margin-bottom: 30px; border-radius: 12px; overflow: hidden; box-shadow: 0 5px 15px rgba(0,0,0,0.08); }
        .section-header { background: #f8f9fa; padding: 20px; border-bottom: 1px solid #e2e8f0; }
        .section-header h2 { color: #2d3748; font-size: 1.5em; }
        .section-content { padding: 20px; }
        .table { width: 100%; border-collapse: collapse; }
        .table th, .table td { padding: 12px; text-align: left; border-bottom: 1px solid #e2e8f0; }
        .table th { background: #f8f9fa; font-weight: 600; color: #4a5568; }
        .table tr:hover { background: #f8f9fa; }
        .status-200 { color: #48bb78; font-weight: bold; }
        .status-300 { color: #ed8936; font-weight: bold; }
        .status-400, .status-500 { color: #f56565; font-weight: bold; }
        .chart-container { margin: 20px 0; }
        .progress-bar { width: 100%; height: 8px; background: #e2e8f0; border-radius: 4px; overflow: hidden; }
        .progress-fill { height: 100%; background: linear-gradient(90deg, #48bb78, #38a169); transition: width 0.3s ease; }
        .error-item { background: #fed7d7; border: 1px solid #feb2b2; border-radius: 6px; padding: 15px; margin-bottom: 10px; }
        .error-url { font-weight: bold; color: #c53030; }
        .error-message { color: #742a2a; font-size: 0.9em; margin-top: 5px; }
        .timestamp { color: #718096; font-size: 0.8em; }
        .tags { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 10px; }
        .tag { background: #edf2f7; color: #4a5568; padding: 4px 12px; border-radius: 20px; font-size: 0.8em; }
        .success-rate { font-size: 1.2em; font-weight: bold; }
        .success-high { color: #48bb78; }
        .success-medium { color: #ed8936; }
        .success-low { color: #f56565; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Stress Test Report</h1>
            <p>Generated on {{.Timestamp}} | Duration: {{.Duration}}</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <h3>Total Requests</h3>
                <div class="value">{{.TotalRequests}}</div>
            </div>
            <div class="stat-card">
                <h3>Success Rate</h3>
                <div class="value success-rate {{.SuccessRateClass}}">{{printf "%.1f" .SuccessRate}}<span class="unit">%</span></div>
            </div>
            <div class="stat-card">
                <h3>URLs Discovered</h3>
                <div class="value">{{.URLsDiscovered}}</div>
            </div>
            <div class="stat-card">
                <h3>Requests/Second</h3>
                <div class="value">{{printf "%.1f" .RequestsPerSecond}}</div>
            </div>
            <div class="stat-card">
                <h3>Avg Response Time</h3>
                <div class="value">{{.AverageResponseTime}}</div>
            </div>
            <div class="stat-card">
                <h3>Failed Requests</h3>
                <div class="value">{{.FailedRequests}}</div>
            </div>
        </div>

        <div class="section">
            <div class="section-header">
                <h2>📊 Response Status Distribution</h2>
            </div>
            <div class="section-content">
                <div class="chart-container">
                    <canvas id="statusChart" width="400" height="200"></canvas>
                </div>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Status Code</th>
                            <th>Count</th>
                            <th>Percentage</th>
                            <th>Progress</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .StatusDistribution}}
                        <tr>
                            <td class="status-{{.StatusGroup}}">{{.StatusCode}}</td>
                            <td>{{.Count}}</td>
                            <td>{{printf "%.1f%%" .Percentage}}</td>
                            <td>
                                <div class="progress-bar">
                                    <div class="progress-fill" style="width: {{.Percentage}}%;"></div>
                                </div>
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        <div class="section">
            <div class="section-header">
                <h2>🔗 URL Validation Results</h2>
            </div>
            <div class="section-content">
                <table class="table">
                    <thead>
                        <tr>
                            <th>URL</th>
                            <th>Status</th>
                            <th>Response Time</th>
                            <th>Content Length</th>
                            <th>Links Found</th>
                            <th>Depth</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .URLValidations}}
                        <tr>
                            <td><a href="{{.URL}}" target="_blank">{{.URL}}</a></td>
                            <td class="status-{{.StatusGroup}}">{{.StatusCode}}</td>
                            <td>{{.ResponseTime}}</td>
                            <td>{{.ContentLength}}</td>
                            <td>{{.LinksFound}}</td>
                            <td>{{.Depth}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        {{if .SlowRequests}}
        <div class="section">
            <div class="section-header">
                <h2>🐌 Slowest Requests (>2s)</h2>
            </div>
            <div class="section-content">
                <table class="table">
                    <thead>
                        <tr>
                            <th>URL</th>
                            <th>Response Time</th>
                            <th>Status Code</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .SlowRequests}}
                        <tr>
                            <td><a href="{{.URL}}" target="_blank">{{.URL}}</a></td>
                            <td>{{.ResponseTime}}</td>
                            <td class="status-{{.StatusGroup}}">{{.StatusCode}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>
        {{end}}

        {{if .Errors}}
        <div class="section">
            <div class="section-header">
                <h2>❌ Errors Encountered</h2>
            </div>
            <div class="section-content">
                {{range .Errors}}
                <div class="error-item">
                    <div class="error-url">{{.URL}}</div>
                    <div class="error-message">{{.Error}}</div>
                    <div class="timestamp">{{.Timestamp.Format "2006-01-02 15:04:05"}}</div>
                </div>
                {{end}}
            </div>
        </div>
        {{end}}

        <div class="section">
            <div class="section-header">
                <h2>📈 Response Time Distribution</h2>
            </div>
            <div class="section-content">
                <div class="chart-container">
                    <canvas id="responseTimeChart" width="400" height="200"></canvas>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Status Code Distribution Chart
        const statusCtx = document.getElementById('statusChart').getContext('2d');
        new Chart(statusCtx, {
            type: 'doughnut',
            data: {
                labels: [{{range .StatusDistribution}}'HTTP {{.StatusCode}}',{{end}}],
                datasets: [{
                    data: [{{range .StatusDistribution}}{{.Count}},{{end}}],
                    backgroundColor: [
                        '#48bb78', '#38a169', '#68d391',
                        '#ed8936', '#f6ad55', '#fbd38d',
                        '#f56565', '#fc8181', '#feb2b2'
                    ]
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: { position: 'bottom' }
                }
            }
        });

        // Response Time Chart
        const timeCtx = document.getElementById('responseTimeChart').getContext('2d');
        const responseTimes = [{{range .ResponseTimesMs}}{{.}},{{end}}];
        
        // Create histogram data
        const buckets = 20;
        const min = Math.min(...responseTimes);
        const max = Math.max(...responseTimes);
        const bucketSize = (max - min) / buckets;
        const histogram = new Array(buckets).fill(0);
        const labels = [];
        
        responseTimes.forEach(time => {
            const bucketIndex = Math.min(Math.floor((time - min) / bucketSize), buckets - 1);
            histogram[bucketIndex]++;
        });
        
        for (let i = 0; i < buckets; i++) {
            const start = min + (i * bucketSize);
            const end = min + ((i + 1) * bucketSize);
            labels.push(start.toFixed(0) + '-' + end.toFixed(0) + 'ms');
        }
        
        new Chart(timeCtx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Request Count',
                    data: histogram,
                    backgroundColor: '#667eea',
                    borderColor: '#764ba2',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: { display: true, text: 'Number of Requests' }
                    },
                    x: {
                        title: { display: true, text: 'Response Time (ms)' }
                    }
                }
            }
        });
    </script>
</body>
</html>`

	// Prepare template data
	data := rg.prepareTemplateData()

	// Parse and execute template
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	file, err := os.Create(outputPath) // #nosec G304 - Safe: output file path is user-specified
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() {
		_ = file.Close() // Best effort cleanup
	}()

	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func (rg *ReportGenerator) GenerateJSONReport(outputPath string) error {
	data, err := json.MarshalIndent(rg.results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("writing JSON file: %w", err)
	}

	return nil
}

func (rg *ReportGenerator) prepareTemplateData() map[string]interface{} {
	// Calculate status distribution
	statusCounts := make(map[int]int)
	for _, validation := range rg.results.URLValidations {
		statusCounts[validation.StatusCode]++
	}

	statusDistribution := make([]map[string]interface{}, 0, len(statusCounts))
	totalValidations := len(rg.results.URLValidations)
	for status, count := range statusCounts {
		percentage := float64(count) / float64(totalValidations) * 100
		statusGroup := "200"
		if status >= 300 && status < 400 {
			statusGroup = "300"
		} else if status >= 400 {
			statusGroup = "400"
		}

		statusDistribution = append(statusDistribution, map[string]interface{}{
			"StatusCode":  status,
			"Count":       count,
			"Percentage":  percentage,
			"StatusGroup": statusGroup,
		})
	}

	// Sort by status code
	sort.Slice(statusDistribution, func(i, j int) bool {
		return statusDistribution[i]["StatusCode"].(int) < statusDistribution[j]["StatusCode"].(int)
	})

	// Prepare URL validations with status groups
	urlValidations := make([]map[string]interface{}, 0, len(rg.results.URLValidations))
	for _, validation := range rg.results.URLValidations {
		statusGroup := "200"
		if validation.StatusCode >= 300 && validation.StatusCode < 400 {
			statusGroup = "300"
		} else if validation.StatusCode >= 400 {
			statusGroup = "400"
		}

		urlValidations = append(urlValidations, map[string]interface{}{
			"URL":           validation.URL,
			"StatusCode":    validation.StatusCode,
			"StatusGroup":   statusGroup,
			"ResponseTime":  validation.ResponseTime.String(),
			"ContentLength": validation.ContentLength,
			"LinksFound":    validation.LinksFound,
			"Depth":         validation.Depth,
		})
	}

	// Prepare slow requests
	slowRequests := make([]map[string]interface{}, 0, len(rg.results.SlowRequests))
	for _, req := range rg.results.SlowRequests {
		statusGroup := "200"
		if req.StatusCode >= 300 && req.StatusCode < 400 {
			statusGroup = "300"
		} else if req.StatusCode >= 400 {
			statusGroup = "400"
		}

		slowRequests = append(slowRequests, map[string]interface{}{
			"URL":          req.URL,
			"ResponseTime": req.ResponseTime.String(),
			"StatusCode":   req.StatusCode,
			"StatusGroup":  statusGroup,
		})
	}

	// Convert response times to milliseconds for charting
	responseTimesMs := make([]float64, 0, len(rg.results.ResponseTimes))
	for _, entry := range rg.results.ResponseTimes {
		responseTimesMs = append(responseTimesMs, float64(entry.ResponseTime.Nanoseconds())/1000000)
	}

	// Determine success rate class
	successRateClass := "success-high"
	if rg.results.SuccessRate < 90 {
		successRateClass = "success-medium"
	}
	if rg.results.SuccessRate < 70 {
		successRateClass = "success-low"
	}

	return map[string]interface{}{
		"Timestamp":           time.Now().Format("2006-01-02 15:04:05 MST"),
		"Duration":            rg.results.Duration,
		"TotalRequests":       rg.results.TotalRequests,
		"SuccessfulRequests":  rg.results.SuccessfulRequests,
		"FailedRequests":      rg.results.FailedRequests,
		"URLsDiscovered":      rg.results.URLsDiscovered,
		"SuccessRate":         rg.results.SuccessRate,
		"SuccessRateClass":    successRateClass,
		"RequestsPerSecond":   rg.results.RequestsPerSecond,
		"AverageResponseTime": rg.results.AverageResponseTime,
		"StatusDistribution":  statusDistribution,
		"URLValidations":      urlValidations,
		"SlowRequests":        slowRequests,
		"Errors":              rg.results.Errors,
		"ResponseTimesMs":     responseTimesMs,
	}
}
