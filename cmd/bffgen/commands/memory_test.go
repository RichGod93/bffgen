package commands

import (
	"runtime"
	"testing"
)

func TestMemory_GenerateNoLeak(t *testing.T) {
	// Skip if not running memory tests
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Run project detection many times
	for i := 0; i < 100; i++ {
		// Simulate generate command operations
		_ = detectProjectType()
		_ = normalizeRuntime("go")
		_ = normalizeRuntime("nodejs-express")
		_ = normalizeRuntime("nodejs-fastify")
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 5*1024*1024 { // 5MB threshold
			t.Errorf("Possible memory leak: grew by %d bytes", growth)
		}
	}
}

func TestMemory_ProjectDetectionNoLeak(t *testing.T) {
	// Skip if not running memory tests
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Run detection many times
	for i := 0; i < 1000; i++ {
		runtime := detectProjectType()
		_ = runtime // Prevent optimization
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 2*1024*1024 { // 2MB threshold
			t.Errorf("Memory leak detected in project detection: grew by %d bytes", growth)
		}
	}
}

func TestMemory_ConfigValidationNoLeak(t *testing.T) {
	// Skip if not running memory tests
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Run validation operations many times
	for i := 0; i < 500; i++ {
		// Simulate validation operations
		_ = validateURL("http://localhost:3000")
		_ = validatePath("/api/users")
		_ = validatePath("/api/products/{id}")
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 1024*1024 { // 1MB threshold
			t.Errorf("Memory leak detected in validation: grew by %d bytes", growth)
		}
	}
}

func TestMemory_GlobalConfigNoLeak(t *testing.T) {
	// Skip if not running memory tests
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var before, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Initialize global config many times
	for i := 0; i < 1000; i++ {
		config := GlobalConfig{
			ConfigPath:      "/test/path",
			Verbose:         false,
			NoColor:         false,
			RuntimeOverride: "",
		}
		_ = config // Prevent optimization
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.Alloc > before.Alloc {
		growth := after.Alloc - before.Alloc
		if growth > 1024*1024 { // 1MB threshold
			t.Errorf("Memory leak detected: grew by %d bytes", growth)
		}
	}
}

// Long-running test for sustained operations
func TestLongRunning_CommandExecution(t *testing.T) {
	// Skip in normal test runs
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	var before, during, after runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&before)

	// Run operations for a longer period
	for i := 0; i < 5000; i++ {
		_ = detectProjectType()
		_ = normalizeRuntime("go")

		if i == 2500 {
			runtime.GC()
			runtime.ReadMemStats(&during)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	// Only check if memory actually grew (handle GC)
	if during.Alloc > before.Alloc && after.Alloc > before.Alloc {
		midGrowth := during.Alloc - before.Alloc
		finalGrowth := after.Alloc - before.Alloc

		// Check that memory stabilizes (doesn't keep growing)
		if finalGrowth > midGrowth*2 {
			t.Errorf("Memory continues to grow: mid=%d bytes, final=%d bytes", midGrowth, finalGrowth)
		}

		if finalGrowth > 10*1024*1024 { // 10MB threshold for long test
			t.Errorf("Excessive memory growth: grew by %d bytes", finalGrowth)
		}
	}
}

// Benchmark to track allocations
func BenchmarkMemory_ProjectDetection(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = detectProjectType()
	}
}

func BenchmarkMemory_URLValidation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = validateURL("http://localhost:3000")
	}
}

func BenchmarkMemory_PathValidation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = validatePath("/api/users/{id}")
	}
}
