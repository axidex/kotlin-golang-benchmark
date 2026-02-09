package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// printResults prints the benchmark results in a formatted table and JSON
func printResults(r *Result, verbose bool) {
	// For mixed operations, calculate total HTTP requests
	// Success cycles = 4 HTTP requests each
	// Failed cycles counted as cycles (partial completion)
	totalHTTPRequests := r.TotalRequests
	actualRPS := float64(r.TotalRequests) / r.TotalDuration.Seconds()

	if r.BenchmarkType == MixedOperations {
		// Successful cycles: 4 HTTP requests each
		// Failed cycles: counted as partial, not multiplied
		totalHTTPRequests = r.SuccessRequests*4 + r.FailedRequests
		actualRPS = float64(totalHTTPRequests) / r.TotalDuration.Seconds()
	}

	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("                      BENCHMARK RESULTS")
	fmt.Println("════════════════════════════════════════════════════════════════")

	if r.BenchmarkType == MixedOperations {
		fmt.Printf("CRUD Cycles:      %d\n", r.TotalRequests)
		fmt.Printf("  Success:        %d cycles (%.2f%%)\n", r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100)
		fmt.Printf("  Failed:         %d cycles (%.2f%%)\n", r.FailedRequests, float64(r.FailedRequests)/float64(r.TotalRequests)*100)
		fmt.Printf("Total HTTP Reqs:  ~%d\n", totalHTTPRequests)
		fmt.Printf("Duration:         %s\n", r.TotalDuration)
		fmt.Printf("Actual RPS:       %.2f req/s\n", actualRPS)
		fmt.Printf("Cycles/sec:       %.2f cycles/s\n", float64(r.TotalRequests)/r.TotalDuration.Seconds())
	} else {
		fmt.Printf("Total Requests:   %d\n", r.TotalRequests)
		fmt.Printf("Success:          %d (%.2f%%)\n", r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100)
		fmt.Printf("Failed:           %d (%.2f%%)\n", r.FailedRequests, float64(r.FailedRequests)/float64(r.TotalRequests)*100)
		fmt.Printf("Duration:         %s\n", r.TotalDuration)
		fmt.Printf("Actual RPS:       %.2f req/s\n", actualRPS)
	}

	fmt.Println("")
	fmt.Println("Latency:")
	if r.BenchmarkType == MixedOperations {
		fmt.Println("  (Full CRUD cycle: CREATE->GET->UPDATE->DELETE)")
	}
	fmt.Printf("  Min:            %s\n", r.MinLatency)
	fmt.Printf("  Avg:            %s\n", r.AvgLatency)
	fmt.Printf("  Max:            %s\n", r.MaxLatency)
	fmt.Printf("  P50:            %s\n", r.P50Latency)
	fmt.Printf("  P95:            %s\n", r.P95Latency)
	fmt.Printf("  P99:            %s\n", r.P99Latency)

	// Print error statistics
	if r.Errors != nil && r.Errors.GetTotalCount() > 0 {
		fmt.Println("")
		fmt.Println("════════════════════════════════════════════════════════════════")
		fmt.Println("                      ERROR STATISTICS")
		fmt.Println("════════════════════════════════════════════════════════════════")
		fmt.Printf("Total Errors:     %d\n", r.Errors.GetTotalCount())
		fmt.Printf("Unique Errors:    %d\n", r.Errors.GetUniqueCount())
		fmt.Println("")

		sortedErrors := r.Errors.GetSortedErrors()

		fmt.Println("┌─────────┬────────────────────────────────┬──────────────┬────────────────────────────────────────┐")
		fmt.Println("│  COUNT  │           OPERATION            │     TYPE     │                MESSAGE                 │")
		fmt.Println("├─────────┼────────────────────────────────┼──────────────┼────────────────────────────────────────┤")

		for _, err := range sortedErrors {
			operation := truncateString(err.Operation, 30)
			errType := truncateString(err.ErrorType, 12)
			message := truncateString(err.ErrorMessage, 38)

			fmt.Printf("│ %7d │ %-30s │ %-12s │ %-38s │\n",
				err.Count, operation, errType, message)
		}
		fmt.Println("└─────────┴────────────────────────────────┴──────────────┴────────────────────────────────────────┘")

		// Detailed output with response bodies (if verbose)
		if verbose {
			fmt.Println("")
			fmt.Println("Detailed Error Samples (with response bodies):")
			fmt.Println("───────────────────────────────────────────────")
			for i, err := range sortedErrors {
				if i >= 10 { // Limit to 10 errors in detailed output
					fmt.Printf("\n... and %d more unique error types\n", len(sortedErrors)-10)
					break
				}
				fmt.Printf("\n[%d] %s | %s\n", err.Count, err.Operation, err.ErrorType)
				fmt.Printf("    Message: %s\n", err.ErrorMessage)
				if err.StatusCode > 0 {
					fmt.Printf("    Status:  %d\n", err.StatusCode)
				}
				fmt.Printf("    First:   %s\n", err.FirstSeen.Format("15:04:05.000"))
				fmt.Printf("    Last:    %s\n", err.LastSeen.Format("15:04:05.000"))
				if err.SampleBody != "" {
					fmt.Printf("    Body:    %s\n", err.SampleBody)
				}
			}
		}
	}

	fmt.Println("")
	fmt.Println("════════════════════════════════════════════════════════════════")

	// JSON results
	errorList := make([]map[string]interface{}, 0)
	if r.Errors != nil {
		for _, err := range r.Errors.GetSortedErrors() {
			errorList = append(errorList, map[string]interface{}{
				"count":      err.Count,
				"operation":  err.Operation,
				"type":       err.ErrorType,
				"message":    err.ErrorMessage,
				"statusCode": err.StatusCode,
				"firstSeen":  err.FirstSeen.Format(time.RFC3339),
				"lastSeen":   err.LastSeen.Format(time.RFC3339),
			})
		}
	}

	jsonData := map[string]interface{}{
		"duration_seconds": r.TotalDuration.Seconds(),
		"latency": map[string]string{
			"min": r.MinLatency.String(),
			"avg": r.AvgLatency.String(),
			"max": r.MaxLatency.String(),
			"p50": r.P50Latency.String(),
			"p95": r.P95Latency.String(),
			"p99": r.P99Latency.String(),
		},
		"errors": map[string]interface{}{
			"total":  r.Errors.GetTotalCount(),
			"unique": r.Errors.GetUniqueCount(),
			"list":   errorList,
		},
	}

	if r.BenchmarkType == MixedOperations {
		jsonData["crud_cycles"] = r.TotalRequests
		jsonData["success_cycles"] = r.SuccessRequests
		jsonData["failed_cycles"] = r.FailedRequests
		jsonData["total_http_requests"] = totalHTTPRequests
		jsonData["rps"] = actualRPS
		jsonData["cycles_per_second"] = float64(r.TotalRequests) / r.TotalDuration.Seconds()
	} else {
		jsonData["total_requests"] = r.TotalRequests
		jsonData["success_requests"] = r.SuccessRequests
		jsonData["failed_requests"] = r.FailedRequests
		jsonData["rps"] = actualRPS
	}

	jsonResult, _ := json.MarshalIndent(jsonData, "", "  ")
	fmt.Println("")
	fmt.Println("JSON Results:")
	fmt.Println(string(jsonResult))
}