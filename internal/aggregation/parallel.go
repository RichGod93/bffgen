package aggregation

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FetchFunc represents a function that fetches data
type FetchFunc func(ctx context.Context) (interface{}, error)

// FetchRequest represents a single fetch request
type FetchRequest struct {
	Name    string
	Fetch   FetchFunc
	Timeout time.Duration
}

// FetchResult represents the result of a fetch operation
type FetchResult struct {
	Service  string
	Data     interface{}
	Error    error
	Success  bool
	Duration time.Duration
}

// ParallelAggregator executes multiple service calls in parallel
type ParallelAggregator struct {
	Timeout  time.Duration
	FailFast bool
}

// NewParallelAggregator creates a new parallel aggregator
func NewParallelAggregator(timeout time.Duration) *ParallelAggregator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ParallelAggregator{
		Timeout:  timeout,
		FailFast: false,
	}
}

// FetchParallel executes multiple fetch requests in parallel
func (pa *ParallelAggregator) FetchParallel(requests []FetchRequest) []FetchResult {
	results := make([]FetchResult, len(requests))
	var wg sync.WaitGroup
	errChan := make(chan error, len(requests))

	for i, req := range requests {
		wg.Add(1)

		go func(index int, request FetchRequest) {
			defer wg.Done()

			// Use request-specific timeout or default
			timeout := request.Timeout
			if timeout == 0 {
				timeout = pa.Timeout
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			startTime := time.Now()

			// Execute fetch
			data, err := request.Fetch(ctx)
			duration := time.Since(startTime)

			result := FetchResult{
				Service:  request.Name,
				Data:     data,
				Error:    err,
				Success:  err == nil,
				Duration: duration,
			}

			results[index] = result

			// Send error to channel if fail-fast is enabled
			if pa.FailFast && err != nil {
				errChan <- err
			}
		}(i, req)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	return results
}

// FetchWaterfall executes requests in sequence, passing previous results
type WaterfallFetchFunc func(ctx context.Context, previousResults []FetchResult) (interface{}, error)

// WaterfallRequest represents a waterfall fetch request
type WaterfallRequest struct {
	Name    string
	Fetch   WaterfallFetchFunc
	Timeout time.Duration
}

// FetchWaterfall executes requests in sequence, each receiving previous results
func (pa *ParallelAggregator) FetchWaterfall(requests []WaterfallRequest) []FetchResult {
	results := make([]FetchResult, 0, len(requests))

	for _, req := range requests {
		timeout := req.Timeout
		if timeout == 0 {
			timeout = pa.Timeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		startTime := time.Now()

		data, err := req.Fetch(ctx, results)
		duration := time.Since(startTime)
		cancel()

		result := FetchResult{
			Service:  req.Name,
			Data:     data,
			Error:    err,
			Success:  err == nil,
			Duration: duration,
		}

		results = append(results, result)

		// Stop on error if fail-fast is enabled
		if pa.FailFast && err != nil {
			break
		}
	}

	return results
}

// GetSuccessfulResults filters results to only successful ones
func GetSuccessfulResults(results []FetchResult) []FetchResult {
	successful := make([]FetchResult, 0, len(results))
	for _, result := range results {
		if result.Success {
			successful = append(successful, result)
		}
	}
	return successful
}

// GetFailedResults filters results to only failed ones
func GetFailedResults(results []FetchResult) []FetchResult {
	failed := make([]FetchResult, 0, len(results))
	for _, result := range results {
		if !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}

// FindResult finds a result by service name
func FindResult(results []FetchResult, serviceName string) *FetchResult {
	for i, result := range results {
		if result.Service == serviceName {
			return &results[i]
		}
	}
	return nil
}

// GetData gets data from a result by service name, with fallback
func GetData(results []FetchResult, serviceName string, fallback interface{}) interface{} {
	result := FindResult(results, serviceName)
	if result != nil && result.Success {
		return result.Data
	}
	return fallback
}

// ResultsToMap converts results to a map for easy access
func ResultsToMap(results []FetchResult) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for _, result := range results {
		if result.Success {
			resultMap[result.Service] = result.Data
		}
	}
	return resultMap
}

// GetTotalDuration returns the total duration of all requests
func GetTotalDuration(results []FetchResult) time.Duration {
	var total time.Duration
	for _, result := range results {
		total += result.Duration
	}
	return total
}

// GetMaxDuration returns the maximum duration among all requests
func GetMaxDuration(results []FetchResult) time.Duration {
	var max time.Duration
	for _, result := range results {
		if result.Duration > max {
			max = result.Duration
		}
	}
	return max
}

// Stats returns statistics about the fetch results
type FetchStats struct {
	Total       int
	Succeeded   int
	Failed      int
	AvgDuration time.Duration
	MaxDuration time.Duration
}

// GetStats calculates statistics from results
func GetStats(results []FetchResult) FetchStats {
	stats := FetchStats{
		Total: len(results),
	}

	var totalDuration time.Duration
	for _, result := range results {
		if result.Success {
			stats.Succeeded++
		} else {
			stats.Failed++
		}
		totalDuration += result.Duration
		if result.Duration > stats.MaxDuration {
			stats.MaxDuration = result.Duration
		}
	}

	if len(results) > 0 {
		stats.AvgDuration = totalDuration / time.Duration(len(results))
	}

	return stats
}

// String provides a string representation of stats
func (fs FetchStats) String() string {
	return fmt.Sprintf("Total: %d, Succeeded: %d, Failed: %d, Avg: %s, Max: %s",
		fs.Total, fs.Succeeded, fs.Failed, fs.AvgDuration, fs.MaxDuration)
}
