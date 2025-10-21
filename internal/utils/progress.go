package utils

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker tracks progress of multi-step operations
type ProgressTracker struct {
	bar         *progressbar.ProgressBar
	startTime   time.Time
	currentStep int
	totalSteps  int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(description string, totalSteps int) *ProgressTracker {
	bar := progressbar.NewOptions(totalSteps,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetElapsedTime(true),
	)

	return &ProgressTracker{
		bar:        bar,
		startTime:  time.Now(),
		totalSteps: totalSteps,
	}
}

// Step increments the progress by one step
func (pt *ProgressTracker) Step(message string) {
	pt.currentStep++
	pt.bar.Add(1)
	if message != "" {
		pt.bar.Describe(message)
	}
}

// SetStep sets the current step with a message
func (pt *ProgressTracker) SetStep(step int, message string) {
	pt.currentStep = step
	pt.bar.Set(step)
	if message != "" {
		pt.bar.Describe(message)
	}
}

// Finish completes the progress bar
func (pt *ProgressTracker) Finish() {
	pt.bar.Finish()
	elapsed := time.Since(pt.startTime)
	fmt.Printf("\n✅ Completed in %s\n", elapsed.Round(time.Millisecond))
}

// Fail marks the progress as failed
func (pt *ProgressTracker) Fail(message string) {
	pt.bar.Describe(fmt.Sprintf("❌ Failed: %s", message))
	fmt.Println()
}

// GetElapsed returns the elapsed time
func (pt *ProgressTracker) GetElapsed() time.Duration {
	return time.Since(pt.startTime)
}

// Spinner provides a simple spinner for indeterminate operations
type Spinner struct {
	message   string
	stopChan  chan bool
	isRunning bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		stopChan: make(chan bool),
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	s.isRunning = true
	go func() {
		chars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-s.stopChan:
				return
			default:
				fmt.Printf("\r%s %s", chars[i%len(chars)], s.message)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.isRunning {
		return
	}
	s.isRunning = false
	s.stopChan <- true
	fmt.Print("\r")
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}

// StepProgress represents a simple step-based progress indicator
type StepProgress struct {
	currentStep int
	totalSteps  int
	startTime   time.Time
}

// NewStepProgress creates a new step progress tracker
func NewStepProgress(totalSteps int) *StepProgress {
	return &StepProgress{
		currentStep: 0,
		totalSteps:  totalSteps,
		startTime:   time.Now(),
	}
}

// Step advances to the next step with a message
func (sp *StepProgress) Step(message string) {
	sp.currentStep++
	percentage := float64(sp.currentStep) / float64(sp.totalSteps) * 100
	elapsed := time.Since(sp.startTime)

	fmt.Printf("[%d/%d] (%.0f%%) %s", sp.currentStep, sp.totalSteps, percentage, message)

	// Show elapsed time if more than 1 second
	if elapsed > time.Second {
		fmt.Printf(" [%s]", elapsed.Round(time.Millisecond))
	}

	fmt.Println()
}

// Finish completes the progress
func (sp *StepProgress) Finish(message string) {
	elapsed := time.Since(sp.startTime)
	fmt.Printf("\n✅ %s (completed in %s)\n", message, elapsed.Round(time.Millisecond))
}

// Fail marks a failure
func (sp *StepProgress) Fail(message string) {
	elapsed := time.Since(sp.startTime)
	fmt.Printf("\n❌ %s (failed after %s)\n", message, elapsed.Round(time.Millisecond))
}

// QuietProgress provides minimal progress output (just icons)
type QuietProgress struct {
	verbose bool
}

// NewQuietProgress creates a quiet progress tracker
func NewQuietProgress(verbose bool) *QuietProgress {
	return &QuietProgress{verbose: verbose}
}

// Start indicates an operation is starting
func (qp *QuietProgress) Start(message string) {
	if qp.verbose {
		fmt.Printf("⏳ %s...\n", message)
	}
}

// Success indicates successful completion
func (qp *QuietProgress) Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Warning indicates a warning
func (qp *QuietProgress) Warning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error indicates an error
func (qp *QuietProgress) Error(message string) {
	fmt.Printf("❌ %s\n", message)
}

// Info provides information
func (qp *QuietProgress) Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}
