package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerationState(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	t.Run("InitializeState", func(t *testing.T) {
		state, err := InitializeState("nodejs-express", "express")
		if err != nil {
			t.Fatalf("Failed to initialize state: %v", err)
		}

		if state.ProjectType != "nodejs-express" {
			t.Errorf("Expected project type nodejs-express, got %s", state.ProjectType)
		}

		if state.Framework != "express" {
			t.Errorf("Expected framework express, got %s", state.Framework)
		}

		if state.GeneratedFiles == nil {
			t.Error("GeneratedFiles should be initialized")
		}

		if state.Routes == nil {
			t.Error("Routes should be initialized")
		}
	})

	t.Run("LoadAndSaveState", func(t *testing.T) {
		// Create a state
		state := &GenerationState{
			Version:        "2.0.0",
			LastGeneration: time.Now(),
			GeneratedFiles: make(map[string]*GeneratedFile),
			Routes:         make(map[string]*RouteState),
			ProjectType:    "go",
			Framework:      "chi",
		}

		// Save state
		err := SaveState(state)
		if err != nil {
			t.Fatalf("Failed to save state: %v", err)
		}

		// Load state
		loaded, err := LoadState()
		if err != nil {
			t.Fatalf("Failed to load state: %v", err)
		}

		if loaded.ProjectType != "go" {
			t.Errorf("Expected project type go, got %s", loaded.ProjectType)
		}

		if loaded.Framework != "chi" {
			t.Errorf("Expected framework chi, got %s", loaded.Framework)
		}
	})

	t.Run("TrackGeneratedFile", func(t *testing.T) {
		state := &GenerationState{
			GeneratedFiles: make(map[string]*GeneratedFile),
		}

		state.TrackGeneratedFile("main.go", "abc123", true)

		if len(state.GeneratedFiles) != 1 {
			t.Errorf("Expected 1 generated file, got %d", len(state.GeneratedFiles))
		}

		file, exists := state.GeneratedFiles["main.go"]
		if !exists {
			t.Error("main.go should be tracked")
		}

		if file.Hash != "abc123" {
			t.Errorf("Expected hash abc123, got %s", file.Hash)
		}

		if !file.HasMarkers {
			t.Error("Expected file to have markers")
		}
	})

	t.Run("TrackRoute", func(t *testing.T) {
		state := &GenerationState{
			Routes: make(map[string]*RouteState),
		}

		state.TrackRoute("users", "GET", "/users", "/api/users")

		if len(state.Routes) != 1 {
			t.Errorf("Expected 1 route, got %d", len(state.Routes))
		}

		if !state.IsRouteGenerated("users", "GET", "/api/users") {
			t.Error("Route should be marked as generated")
		}

		if state.IsRouteGenerated("users", "POST", "/api/users") {
			t.Error("POST route should not be marked as generated")
		}
	})

	t.Run("CleanupOldBackups", func(t *testing.T) {
		// Create backup directory with old file
		backupDir := GetBackupDir()
		os.MkdirAll(backupDir, 0755)

		oldFile := filepath.Join(backupDir, "old.txt")
		os.WriteFile(oldFile, []byte("old"), 0644)

		// Wait a moment
		time.Sleep(10 * time.Millisecond)

		// Create new file
		newFile := filepath.Join(backupDir, "new.txt")
		os.WriteFile(newFile, []byte("new"), 0644)

		// Cleanup files older than 5ms
		err := CleanupOldBackups(5 * time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to cleanup backups: %v", err)
		}

		// Old file should be removed
		if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
			t.Error("Old file should have been removed")
		}

		// New file should still exist
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			t.Error("New file should still exist")
		}
	})
}

func TestGetStateDir(t *testing.T) {
	dir := GetStateDir()
	if dir != ".bffgen" {
		t.Errorf("Expected .bffgen, got %s", dir)
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	timestamp := GetCurrentTimestamp()
	if timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	// Parse timestamp to ensure it's valid RFC3339
	_, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}
