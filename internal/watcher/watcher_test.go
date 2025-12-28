package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewConfigWatcher(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create the config file
	if err := os.WriteFile(configPath, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	changed := false
	onChange := func(path string) error {
		changed = true
		return nil
	}

	watcher, err := NewConfigWatcher([]string{configPath}, onChange)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Fatal("NewConfigWatcher() returned nil")
	}

	t.Run("initializes with stopped=false", func(t *testing.T) {
		if watcher.stopped {
			t.Error("Expected stopped to be false initially")
		}
	})

	t.Run("has correct debounce period", func(t *testing.T) {
		if watcher.debouncePeriod != 300*time.Millisecond {
			t.Errorf("Expected debounce period = 300ms, got %v", watcher.debouncePeriod)
		}
	})

	_ = changed // avoid unused warning
}

func TestConfigWatcher_Start(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	changeCount := 0
	onChange := func(path string) error {
		changeCount++
		return nil
	}

	watcher, err := NewConfigWatcher([]string{configPath}, onChange)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Give the watcher time to start
	time.Sleep(50 * time.Millisecond)
}

func TestConfigWatcher_Stop(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	watcher, err := NewConfigWatcher([]string{configPath}, func(path string) error { return nil })
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}

	if err := watcher.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if !watcher.stopped {
		t.Error("Expected stopped to be true after Stop()")
	}
}

func TestConfigWatcher_isWatchedFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	otherPath := filepath.Join(tempDir, "other.yaml")

	if err := os.WriteFile(configPath, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	if err := os.WriteFile(otherPath, []byte("other: value"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	watcher, err := NewConfigWatcher([]string{configPath}, func(path string) error { return nil })
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	t.Run("returns true for watched file", func(t *testing.T) {
		if !watcher.isWatchedFile(configPath) {
			t.Error("Expected isWatchedFile() = true for watched file")
		}
	})

	t.Run("returns false for non-watched file", func(t *testing.T) {
		if watcher.isWatchedFile(otherPath) {
			t.Error("Expected isWatchedFile() = false for non-watched file")
		}
	})

	t.Run("returns false for non-existent file", func(t *testing.T) {
		if watcher.isWatchedFile("/nonexistent/path") {
			t.Error("Expected isWatchedFile() = false for non-existent file")
		}
	})
}

func TestConfigWatcher_handleChange(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("test: value"), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	var mu sync.Mutex
	changeCount := 0
	onChange := func(path string) error {
		mu.Lock()
		changeCount++
		mu.Unlock()
		return nil
	}

	watcher, err := NewConfigWatcher([]string{configPath}, onChange)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	t.Run("debounces rapid changes", func(t *testing.T) {
		// Trigger multiple changes rapidly
		watcher.handleChange(configPath)
		watcher.handleChange(configPath)
		watcher.handleChange(configPath)

		// Wait for debounce to complete
		time.Sleep(400 * time.Millisecond)

		// Should only have one change due to debouncing
		mu.Lock()
		count := changeCount
		mu.Unlock()
		if count != 1 {
			t.Errorf("Expected 1 change after debounce, got %d", count)
		}
	})

	t.Run("does not trigger when stopped", func(t *testing.T) {
		mu.Lock()
		initialCount := changeCount
		mu.Unlock()
		watcher.stopped = true
		watcher.handleChange(configPath)
		time.Sleep(400 * time.Millisecond)

		mu.Lock()
		count := changeCount
		mu.Unlock()
		if count != initialCount {
			t.Error("handleChange should not trigger callback when stopped")
		}
	})
}

func TestNewConfigWatcher_InvalidPath(t *testing.T) {
	// Test with a path that doesn't exist
	invalidPath := "/this/path/definitely/does/not/exist/config.yaml"

	_, err := NewConfigWatcher([]string{invalidPath}, func(path string) error { return nil })
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}
