package main

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// UniqueError represents a unique error with counter
type UniqueError struct {
	Operation    string    // Operation (GET /api/products, POST /api/products, etc.)
	ErrorType    string    // Error type (network_error, http_error, etc.)
	ErrorMessage string    // Error message (normalized)
	StatusCode   int       // HTTP status code (if applicable)
	Count        int64     // Count of such errors
	FirstSeen    time.Time // First occurrence
	LastSeen     time.Time // Last occurrence
	SampleBody   string    // Sample response body (for first error)
}

// ErrorKey is a key for grouping unique errors
type ErrorKey struct {
	Operation    string
	ErrorType    string
	ErrorMessage string
	StatusCode   int
}

// ErrorStats stores statistics for unique errors
type ErrorStats struct {
	mu           sync.Mutex
	UniqueErrors map[ErrorKey]*UniqueError // Unique errors
	TotalCount   int64                     // Total error count
}

func NewErrorStats() *ErrorStats {
	return &ErrorStats{
		UniqueErrors: make(map[ErrorKey]*UniqueError),
	}
}

// RecordError records an error, grouping by unique key
func (es *ErrorStats) RecordError(operation, errType, errMsg string, statusCode int, responseBody string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Normalize error message (remove dynamic parts)
	normalizedMsg := normalizeErrorMessage(errMsg)

	key := ErrorKey{
		Operation:    operation,
		ErrorType:    errType,
		ErrorMessage: normalizedMsg,
		StatusCode:   statusCode,
	}

	now := time.Now()
	es.TotalCount++

	if existing, ok := es.UniqueErrors[key]; ok {
		existing.Count++
		existing.LastSeen = now
	} else {
		es.UniqueErrors[key] = &UniqueError{
			Operation:    operation,
			ErrorType:    errType,
			ErrorMessage: normalizedMsg,
			StatusCode:   statusCode,
			Count:        1,
			FirstSeen:    now,
			LastSeen:     now,
			SampleBody:   truncateString(responseBody, 500),
		}
	}
}

// GetSortedErrors returns errors sorted by count (descending)
func (es *ErrorStats) GetSortedErrors() []*UniqueError {
	es.mu.Lock()
	defer es.mu.Unlock()

	errors := make([]*UniqueError, 0, len(es.UniqueErrors))
	for _, err := range es.UniqueErrors {
		errors = append(errors, err)
	}

	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Count > errors[j].Count
	})

	return errors
}

// GetTotalCount returns total error count
func (es *ErrorStats) GetTotalCount() int64 {
	es.mu.Lock()
	defer es.mu.Unlock()
	return es.TotalCount
}

// GetUniqueCount returns unique error count
func (es *ErrorStats) GetUniqueCount() int {
	es.mu.Lock()
	defer es.mu.Unlock()
	return len(es.UniqueErrors)
}

// normalizeErrorMessage normalizes error message by removing dynamic parts
func normalizeErrorMessage(msg string) string {
	// Can add more normalization rules here
	// For example, remove IP addresses, ports, timestamps, etc.
	return msg
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}