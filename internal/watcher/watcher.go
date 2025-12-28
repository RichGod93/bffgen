package watcher

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher monitors configuration files for changes
type ConfigWatcher struct {
	watcher        *fsnotify.Watcher
	configPaths    []string
	onChange       func(path string) error
	debounceTimer  *time.Timer
	debouncePeriod time.Duration
	stopped        bool
}

// NewConfigWatcher creates a new configuration file watcher
func NewConfigWatcher(configPaths []string, onChange func(path string) error) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	cw := &ConfigWatcher{
		watcher:        watcher,
		configPaths:    configPaths,
		onChange:       onChange,
		debouncePeriod: 300 * time.Millisecond,
		stopped:        false,
	}

	// Add all config files to watch
	for _, path := range configPaths {
		if err := watcher.Add(filepath.Dir(path)); err != nil {
			return nil, fmt.Errorf("failed to watch %s: %w", path, err)
		}
	}

	return cw, nil
}

// Start begins watching for file changes
func (cw *ConfigWatcher) Start() error {
	go func() {
		for {
			select {
			case event, ok := <-cw.watcher.Events:
				if !ok {
					return
				}

				// Only process writes to our config files
				if event.Has(fsnotify.Write) && cw.isWatchedFile(event.Name) {
					cw.handleChange(event.Name)
				}

			case err, ok := <-cw.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

// handleChange debounces file changes and triggers onChange callback
func (cw *ConfigWatcher) handleChange(path string) {
	if cw.stopped {
		return
	}

	// Cancel existing timer if any
	if cw.debounceTimer != nil {
		cw.debounceTimer.Stop()
	}

	// Set new timer
	cw.debounceTimer = time.AfterFunc(cw.debouncePeriod, func() {
		if err := cw.onChange(path); err != nil {
			fmt.Printf("Error processing change: %v\n", err)
		}
	})
}

// isWatchedFile checks if the changed file is one we're monitoring
func (cw *ConfigWatcher) isWatchedFile(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, configPath := range cw.configPaths {
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			continue
		}
		if absPath == absConfigPath {
			return true
		}
	}
	return false
}

// Stop stops the watcher
func (cw *ConfigWatcher) Stop() error {
	cw.stopped = true
	if cw.debounceTimer != nil {
		cw.debounceTimer.Stop()
	}
	return cw.watcher.Close()
}
