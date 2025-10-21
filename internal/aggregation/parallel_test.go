package aggregation

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestParallelAggregator(t *testing.T) {
	t.Run("FetchParallel_Success", func(t *testing.T) {
		agg := NewParallelAggregator(5 * time.Second)

		requests := []FetchRequest{
			{
				Name: "service1",
				Fetch: func(ctx context.Context) (interface{}, error) {
					return map[string]string{"data": "service1"}, nil
				},
			},
			{
				Name: "service2",
				Fetch: func(ctx context.Context) (interface{}, error) {
					return map[string]string{"data": "service2"}, nil
				},
			},
		}

		results := agg.FetchParallel(requests)

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		for _, result := range results {
			if !result.Success {
				t.Errorf("Service %s failed: %v", result.Service, result.Error)
			}
			if result.Data == nil {
				t.Errorf("Service %s returned nil data", result.Service)
			}
		}
	})

	t.Run("FetchParallel_WithErrors", func(t *testing.T) {
		agg := NewParallelAggregator(5 * time.Second)

		requests := []FetchRequest{
			{
				Name: "success",
				Fetch: func(ctx context.Context) (interface{}, error) {
					return "ok", nil
				},
			},
			{
				Name: "failure",
				Fetch: func(ctx context.Context) (interface{}, error) {
					return nil, errors.New("service error")
				},
			},
		}

		results := agg.FetchParallel(requests)

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		successful := GetSuccessfulResults(results)
		if len(successful) != 1 {
			t.Errorf("Expected 1 successful result, got %d", len(successful))
		}

		failed := GetFailedResults(results)
		if len(failed) != 1 {
			t.Errorf("Expected 1 failed result, got %d", len(failed))
		}
	})

	t.Run("FetchParallel_Timeout", func(t *testing.T) {
		agg := NewParallelAggregator(100 * time.Millisecond)

		requests := []FetchRequest{
			{
				Name: "slow",
				Fetch: func(ctx context.Context) (interface{}, error) {
					select {
					case <-time.After(1 * time.Second):
						return "done", nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				},
			},
		}

		results := agg.FetchParallel(requests)

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].Success {
			t.Error("Expected timeout error")
		}
	})

	t.Run("FetchWaterfall", func(t *testing.T) {
		agg := NewParallelAggregator(5 * time.Second)

		requests := []WaterfallRequest{
			{
				Name: "first",
				Fetch: func(ctx context.Context, prev []FetchResult) (interface{}, error) {
					return "first-result", nil
				},
			},
			{
				Name: "second",
				Fetch: func(ctx context.Context, prev []FetchResult) (interface{}, error) {
					if len(prev) == 0 {
						return nil, errors.New("no previous results")
					}
					return "second-result", nil
				},
			},
		}

		results := agg.FetchWaterfall(requests)

		if len(results) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(results))
		}

		for i, result := range results {
			if !result.Success {
				t.Errorf("Request %d failed: %v", i, result.Error)
			}
		}
	})
}

func TestFetchHelpers(t *testing.T) {
	results := []FetchResult{
		{Service: "service1", Data: "data1", Success: true},
		{Service: "service2", Data: nil, Success: false, Error: errors.New("error")},
		{Service: "service3", Data: "data3", Success: true},
	}

	t.Run("GetSuccessfulResults", func(t *testing.T) {
		successful := GetSuccessfulResults(results)
		if len(successful) != 2 {
			t.Errorf("Expected 2 successful results, got %d", len(successful))
		}
	})

	t.Run("GetFailedResults", func(t *testing.T) {
		failed := GetFailedResults(results)
		if len(failed) != 1 {
			t.Errorf("Expected 1 failed result, got %d", len(failed))
		}
	})

	t.Run("FindResult", func(t *testing.T) {
		result := FindResult(results, "service1")
		if result == nil {
			t.Fatal("Should find service1")
		}
		if result.Service != "service1" {
			t.Errorf("Expected service1, got %s", result.Service)
		}

		notFound := FindResult(results, "nonexistent")
		if notFound != nil {
			t.Error("Should not find nonexistent service")
		}
	})

	t.Run("GetData", func(t *testing.T) {
		data := GetData(results, "service1", "fallback")
		if data != "data1" {
			t.Errorf("Expected 'data1', got %v", data)
		}

		fallbackData := GetData(results, "service2", "fallback")
		if fallbackData != "fallback" {
			t.Errorf("Expected 'fallback', got %v", fallbackData)
		}
	})

	t.Run("ResultsToMap", func(t *testing.T) {
		resultMap := ResultsToMap(results)

		if len(resultMap) != 2 {
			t.Errorf("Expected 2 entries in map, got %d", len(resultMap))
		}

		if resultMap["service1"] != "data1" {
			t.Error("service1 data incorrect")
		}

		if _, exists := resultMap["service2"]; exists {
			t.Error("Failed service should not be in map")
		}
	})
}

func TestFetchStats(t *testing.T) {
	results := []FetchResult{
		{Service: "s1", Success: true, Duration: 100 * time.Millisecond},
		{Service: "s2", Success: true, Duration: 200 * time.Millisecond},
		{Service: "s3", Success: false, Duration: 50 * time.Millisecond},
	}

	stats := GetStats(results)

	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}

	if stats.Succeeded != 2 {
		t.Errorf("Expected 2 succeeded, got %d", stats.Succeeded)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}

	if stats.MaxDuration != 200*time.Millisecond {
		t.Errorf("Expected max duration 200ms, got %v", stats.MaxDuration)
	}

	expectedAvg := (100 + 200 + 50) / 3
	actualAvg := stats.AvgDuration.Milliseconds()
	if actualAvg != int64(expectedAvg) {
		t.Errorf("Expected avg duration ~%dms, got %dms", expectedAvg, actualAvg)
	}

	// Test String() method
	str := stats.String()
	if str == "" {
		t.Error("Stats string should not be empty")
	}
}

func TestGetTotalDuration(t *testing.T) {
	results := []FetchResult{
		{Duration: 100 * time.Millisecond},
		{Duration: 200 * time.Millisecond},
	}

	total := GetTotalDuration(results)
	if total != 300*time.Millisecond {
		t.Errorf("Expected 300ms, got %v", total)
	}
}

func TestGetMaxDuration(t *testing.T) {
	results := []FetchResult{
		{Duration: 100 * time.Millisecond},
		{Duration: 200 * time.Millisecond},
		{Duration: 50 * time.Millisecond},
	}

	max := GetMaxDuration(results)
	if max != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", max)
	}
}
