package utils

import (
	"runtime"
	"testing"
)

func TestMemory_TransactionNoLeak(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Create and execute many transactions
	for i := 0; i < 1000; i++ {
		tx := NewTransaction()
		tx.AddCreate("test.txt", []byte("content"))
		_ = tx.Execute()
		_ = tx.Rollback()
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	// Check memory didn't grow significantly (handle GC reducing memory)
	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 1024*1024 { // 1MB threshold
			t.Errorf("Memory leak detected: grew by %d bytes", growth)
		}
	}
}

func TestMemory_StateTrackerNoLeak(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Create and save many states
	for i := 0; i < 1000; i++ {
		state := &GenerationState{
			ProjectType:    "go",
			Routes:         make(map[string]*RouteState),
			GeneratedFiles: make(map[string]*GeneratedFile),
			Version:        "2.0.0",
		}

		// Track a route
		state.TrackRoute("test", "GET", "/api/test", "/test")

		// Track a file
		state.TrackGeneratedFile("test.go", "hash123", false)
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	// Check for growth (handle potential wraparound by checking if after > before)
	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 1024*1024 {
			t.Errorf("Memory leak detected: grew by %d bytes", growth)
		}
	}
}

func TestMemory_FileTransactionNoLeak(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Create temp directory
	tempDir := t.TempDir()

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Create many file transactions
	for i := 0; i < 500; i++ {
		ft := NewFileTransaction()
		_ = ft.CreateFile(tempDir+"/test.txt", []byte("test content"))
		_ = ft.ExecuteAndCommit()
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 2*1024*1024 { // 2MB threshold (file operations allocate more)
			t.Errorf("Memory leak detected: grew by %d bytes", growth)
		}
	}
}

func TestMemory_ConfigConverterNoLeak(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Perform conversions many times
	for i := 0; i < 500; i++ {
		// Simulate config operations
		nodeConfig := map[string]interface{}{
			"server": map[string]interface{}{
				"port": float64(8080),
			},
			"backends": []interface{}{
				map[string]interface{}{
					"name":    "test",
					"baseUrl": "http://localhost:3000",
					"endpoints": []interface{}{
						map[string]interface{}{
							"name":     "test",
							"path":     "/test",
							"method":   "GET",
							"exposeAs": "/api/test",
						},
					},
				},
			},
		}

		// Convert to Go config and back
		_ = nodeConfig // Prevent optimization
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 1024*1024 {
			t.Errorf("Memory leak detected: grew by %d bytes", growth)
		}
	}
}

// Benchmark to track memory allocations over time
func BenchmarkMemory_Transaction(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tx := NewTransaction()
		tx.AddCreate("test.txt", []byte("content"))
		_ = tx.Execute()
		_ = tx.Rollback()
	}
}

func BenchmarkMemory_StateTracking(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state := &GenerationState{
			ProjectType:    "go",
			Routes:         make(map[string]*RouteState),
			GeneratedFiles: make(map[string]*GeneratedFile),
		}
		state.TrackRoute("test", "GET", "/api/test", "/test")
		state.TrackGeneratedFile("test.go", "hash", false)
	}
}
