package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout/stderr during test execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureErrorOutput captures stderr during test execution
func captureErrorOutput(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestLogSuccess verifies success message formatting
func TestLogSuccess(t *testing.T) {
	output := captureOutput(func() {
		LogSuccess("Operation completed")
	})

	if !strings.Contains(output, "✅") {
		t.Error("Expected success emoji in output")
	}

	if !strings.Contains(output, "Operation completed") {
		t.Error("Expected message text in output")
	}
}

// TestLogWarning verifies warning message formatting
func TestLogWarning(t *testing.T) {
	output := captureOutput(func() {
		LogWarning("This is a warning")
	})

	if !strings.Contains(output, "⚠️") {
		t.Error("Expected warning emoji in output")
	}

	if !strings.Contains(output, "Warning") {
		t.Error("Expected 'Warning' text in output")
	}

	if !strings.Contains(output, "This is a warning") {
		t.Error("Expected warning message in output")
	}
}

// TestLogInfo verifies info message formatting
func TestLogInfo(t *testing.T) {
	output := captureOutput(func() {
		LogInfo("This is info")
	})

	if !strings.Contains(output, "💡") {
		t.Error("Expected info emoji in output")
	}

	if !strings.Contains(output, "This is info") {
		t.Error("Expected info message in output")
	}
}

// TestLogError verifies error message formatting (without exit)
func TestLogError(t *testing.T) {
	output := captureErrorOutput(func() {
		LogError(fmt.Errorf("test error"))
	})

	if !strings.Contains(output, "❌") {
		t.Error("Expected error emoji in output")
	}

	if !strings.Contains(output, "test error") {
		t.Error("Expected error message in output")
	}
}

// TestLogErrorWithNil verifies nil error handling
func TestLogErrorWithNil(t *testing.T) {
	// Should not panic or produce output
	output := captureErrorOutput(func() {
		LogError(nil)
	})

	if len(output) > 0 {
		t.Error("Expected no output for nil error")
	}
}

// TestValidateError verifies validation error formatting
func TestValidateError(t *testing.T) {
	output := captureErrorOutput(func() {
		ValidateError(fmt.Errorf("invalid format"), "email")
	})

	if !strings.Contains(output, "Validation Error") {
		t.Error("Expected 'Validation Error' in output")
	}

	if !strings.Contains(output, "email") {
		t.Error("Expected field name in output")
	}

	if !strings.Contains(output, "invalid format") {
		t.Error("Expected error message in output")
	}
}

// TestErrorContext wraps errors with context
func TestErrorContext(t *testing.T) {
	baseErr := fmt.Errorf("connection refused")
	wrappedErr := ErrorContext(baseErr, "database connection")

	if wrappedErr == nil {
		t.Fatal("Expected wrapped error, got nil")
	}

	errMsg := wrappedErr.Error()
	if !strings.Contains(errMsg, "database connection") {
		t.Error("Expected operation context in error")
	}

	if !strings.Contains(errMsg, "connection refused") {
		t.Error("Expected original error message in error")
	}
}

// TestErrorContextWithNil verifies nil error handling
func TestErrorContextWithNil(t *testing.T) {
	result := ErrorContext(nil, "operation")

	if result != nil {
		t.Error("Expected nil for nil error input")
	}
}

// TestLogVerboseCommandWhenDisabled verifies no output when verbose is disabled
func TestLogVerboseCommandWhenDisabled(t *testing.T) {
	// Save original config
	origVerbose := globalConfig.Verbose
	globalConfig.Verbose = false
	defer func() { globalConfig.Verbose = origVerbose }()

	output := captureOutput(func() {
		LogVerboseCommand("verbose message")
	})

	if len(output) > 0 {
		t.Error("Expected no output when verbose is disabled")
	}
}

// TestLogVerboseCommandWhenEnabled verifies output when verbose is enabled
func TestLogVerboseCommandWhenEnabled(t *testing.T) {
	// Save original config
	origVerbose := globalConfig.Verbose
	globalConfig.Verbose = true
	defer func() { globalConfig.Verbose = origVerbose }()

	output := captureOutput(func() {
		LogVerboseCommand("verbose message")
	})

	if len(output) == 0 {
		t.Error("Expected output when verbose is enabled")
	}

	if !strings.Contains(output, "verbose message") {
		t.Error("Expected message in output")
	}
}

// TestPrintSeparator verifies separator output
func TestPrintSeparator(t *testing.T) {
	output := captureOutput(func() {
		PrintSeparator()
	})

	if !strings.Contains(output, "─") {
		t.Error("Expected separator characters in output")
	}
}

// TestPrintSection verifies section header formatting
func TestPrintSection(t *testing.T) {
	output := captureOutput(func() {
		PrintSection("Test Section")
	})

	if !strings.Contains(output, "Test Section") {
		t.Error("Expected section title in output")
	}

	if !strings.Contains(output, "━") {
		t.Error("Expected section border characters in output")
	}
}

// TestPrintTask verifies task status formatting
func TestPrintTask(t *testing.T) {
	output := captureOutput(func() {
		PrintTask("Initialize project", "✅ Done")
	})

	if !strings.Contains(output, "Initialize project") {
		t.Error("Expected task name in output")
	}

	if !strings.Contains(output, "✅ Done") {
		t.Error("Expected task status in output")
	}

	if !strings.Contains(output, "├") {
		t.Error("Expected task marker in output")
	}
}

// TestConsistentFormatting verifies all output functions use consistent formatting
func TestConsistentFormatting(t *testing.T) {
	tests := []struct {
		name   string
		output func() string
	}{
		{
			name: "LogSuccess",
			output: func() string {
				return captureOutput(func() {
					LogSuccess("test")
				})
			},
		},
		{
			name: "LogWarning",
			output: func() string {
				return captureOutput(func() {
					LogWarning("test")
				})
			},
		},
		{
			name: "LogInfo",
			output: func() string {
				return captureOutput(func() {
					LogInfo("test")
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.output()
			// All should end with newline
			if !strings.HasSuffix(output, "\n") {
				t.Error("Expected newline at end of output")
			}
		})
	}
}
