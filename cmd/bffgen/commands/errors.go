package commands

import (
	"fmt"
	"os"
	"strings"
)

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// HandleError logs an error message and exits with code 1
// Used for fatal errors that should stop execution
func HandleError(err error, context string) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s❌ Error%s", colorRed, colorReset)
	if context != "" {
		fmt.Fprintf(os.Stderr, " (%s)", context)
	}
	fmt.Fprintf(os.Stderr, ": %v\n", err)
	os.Exit(1)
}

// LogError logs an error message without exiting
// Used for non-fatal errors that can be recovered from
func LogError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s❌ Error%s: %v\n", colorRed, colorReset, err)
}

// LogSuccess logs a success message in green
// Used to indicate successful completion of operations
func LogSuccess(message string) {
	fmt.Printf("%s✅ %s%s\n", colorGreen, message, colorReset)
}

// LogWarning logs a warning message in yellow
// Used to indicate potentially problematic conditions
func LogWarning(message string) {
	fmt.Printf("%s⚠️  Warning%s: %s\n", colorYellow, colorReset, message)
}

// LogInfo logs an informational message in blue
// Used for general informational messages
func LogInfo(message string) {
	fmt.Printf("%s💡 %s%s\n", colorBlue, message, colorReset)
}

// ValidateError logs a validation error with context
// Used when user input validation fails
func ValidateError(err error, fieldName string) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "%s❌ Validation Error%s: %s - %v\n", colorRed, colorReset, fieldName, err)
}

// ErrorContext wraps an error with operation context
// Used to add context to errors for debugging
func ErrorContext(err error, operation string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s failed: %w", operation, err)
}

// LogVerbose logs a verbose message if verbose mode is enabled
// Used for debug output
func LogVerboseCommand(format string, args ...interface{}) {
	if globalConfig.Verbose {
		fmt.Printf("%s→ %s%s\n", colorCyan, fmt.Sprintf(format, args...), colorReset)
	}
}

// PromptUser displays a prompt and waits for user confirmation
// Returns true if user confirms, false otherwise
func PromptUser(message string) bool {
	fmt.Printf("%s?%s %s (y/n): ", colorCyan, colorReset, message)

	var response string
	_, _ = fmt.Scanln(&response)

	return response == "y" || response == "Y" || response == "yes" || response == "YES"
}

// PrintSeparator prints a visual separator for better output organization
func PrintSeparator() {
	fmt.Println("────────────────────────────────────────────────────────")
}

// PrintSection prints a section header with formatting
func PrintSection(title string) {
	fmt.Printf("\n%s━━━ %s %s━━━%s\n", colorCyan, title, colorCyan, colorReset)
}

// PrintTask prints a task status message
func PrintTask(task string, status string) {
	fmt.Printf("  %s├─ %s...%s %s\n", colorCyan, task, colorReset, status)
}

// toTitleCase converts a string to Title Case (replaces deprecated strings.Title)
func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
