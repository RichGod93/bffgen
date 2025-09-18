package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/RichGod93/bffgen/internal/types"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if path == "" {
		t.Fatal("Expected non-empty config path")
	}
	
	// Check that path contains expected components
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}
	
	// Check that directory was created
	configDir := filepath.Dir(path)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected config directory to exist: %s", configDir)
	}
}

func TestLoadBFFGenConfig_NoFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Temporarily change the config directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)
	
	// Load config when file doesn't exist
	config, err := LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	
	// Should return default config
	if config.Defaults.Framework != "chi" {
		t.Errorf("Expected default framework 'chi', got %s", config.Defaults.Framework)
	}
}

func TestLoadBFFGenConfig_WithFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Temporarily change the config directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)
	
	// Create a test config file
	testConfig := &types.BFFGenConfig{
		Defaults: types.Defaults{
			Framework:   "echo",
			CORSOrigins: []string{"https://test.com"},
			JWTSecret:   "test-secret",
			RedisURL:    "redis://test:6379",
			Port:        3000,
			RouteOption: "1",
		},
		User: types.User{
			Name:    "Test User",
			Email:   "test@example.com",
			GitHub:  "testuser",
			Company: "Test Company",
		},
		History: types.History{
			RecentProjects: []string{"test-project"},
			LastUsed:       "test-project",
		},
	}
	
	// Save the config first
	err := SaveBFFGenConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	// Load the config
	config, err := LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	
	// Verify loaded config matches saved config
	if config.Defaults.Framework != "echo" {
		t.Errorf("Expected framework 'echo', got %s", config.Defaults.Framework)
	}
	
	if len(config.Defaults.CORSOrigins) != 1 {
		t.Errorf("Expected 1 CORS origin, got %d", len(config.Defaults.CORSOrigins))
	}
	
	if config.Defaults.CORSOrigins[0] != "https://test.com" {
		t.Errorf("Expected CORS origin 'https://test.com', got %s", config.Defaults.CORSOrigins[0])
	}
	
	if config.User.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got %s", config.User.Name)
	}
	
	if len(config.History.RecentProjects) != 1 {
		t.Errorf("Expected 1 recent project, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.RecentProjects[0] != "test-project" {
		t.Errorf("Expected recent project 'test-project', got %s", config.History.RecentProjects[0])
	}
}

func TestSaveBFFGenConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Temporarily change the config directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)
	
	config := &types.BFFGenConfig{
		Defaults: types.Defaults{
			Framework:   "fiber",
			CORSOrigins: []string{"http://localhost:3000"},
			JWTSecret:   "save-test-secret",
			RedisURL:    "redis://save-test:6379",
			Port:        4000,
			RouteOption: "2",
		},
		User: types.User{
			Name:    "Save Test User",
			Email:   "save@example.com",
			GitHub:  "savetest",
			Company: "Save Test Company",
		},
		History: types.History{
			RecentProjects: []string{"save-project-1", "save-project-2"},
			LastUsed:       "save-project-1",
		},
	}
	
	err := SaveBFFGenConfig(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify file was created
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Expected config file to exist")
	}
	
	// Load and verify saved config
	loadedConfig, err := LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	
	if loadedConfig.Defaults.Framework != "fiber" {
		t.Errorf("Expected saved framework 'fiber', got %s", loadedConfig.Defaults.Framework)
	}
	
	if loadedConfig.User.Name != "Save Test User" {
		t.Errorf("Expected saved user name 'Save Test User', got %s", loadedConfig.User.Name)
	}
	
	if len(loadedConfig.History.RecentProjects) != 2 {
		t.Errorf("Expected 2 saved recent projects, got %d", len(loadedConfig.History.RecentProjects))
	}
}

func TestUpdateRecentProject(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Temporarily change the config directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)
	
	// Test adding first project
	err := UpdateRecentProject("project-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	config, err := LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if len(config.History.RecentProjects) != 1 {
		t.Errorf("Expected 1 recent project, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.RecentProjects[0] != "project-1" {
		t.Errorf("Expected recent project 'project-1', got %s", config.History.RecentProjects[0])
	}
	
	if config.History.LastUsed != "project-1" {
		t.Errorf("Expected last used 'project-1', got %s", config.History.LastUsed)
	}
	
	// Test adding second project
	err = UpdateRecentProject("project-2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	config, err = LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if len(config.History.RecentProjects) != 2 {
		t.Errorf("Expected 2 recent projects, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.RecentProjects[0] != "project-2" {
		t.Errorf("Expected first recent project 'project-2', got %s", config.History.RecentProjects[0])
	}
	
	if config.History.RecentProjects[1] != "project-1" {
		t.Errorf("Expected second recent project 'project-1', got %s", config.History.RecentProjects[1])
	}
	
	if config.History.LastUsed != "project-2" {
		t.Errorf("Expected last used 'project-2', got %s", config.History.LastUsed)
	}
	
	// Test adding duplicate project (should move to front)
	err = UpdateRecentProject("project-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	config, err = LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if len(config.History.RecentProjects) != 2 {
		t.Errorf("Expected 2 recent projects, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.RecentProjects[0] != "project-1" {
		t.Errorf("Expected first recent project 'project-1', got %s", config.History.RecentProjects[0])
	}
	
	if config.History.RecentProjects[1] != "project-2" {
		t.Errorf("Expected second recent project 'project-2', got %s", config.History.RecentProjects[1])
	}
	
	if config.History.LastUsed != "project-1" {
		t.Errorf("Expected last used 'project-1', got %s", config.History.LastUsed)
	}
}

func TestUpdateRecentProject_Limit(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Temporarily change the config directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempDir)
	
	// Add more than 10 projects to test the limit
	for i := 1; i <= 12; i++ {
		projectName := fmt.Sprintf("project-%d", i)
		err := UpdateRecentProject(projectName)
		if err != nil {
			t.Fatalf("Expected no error for project %d, got %v", i, err)
		}
	}
	
	config, err := LoadBFFGenConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Should only keep the last 10 projects
	if len(config.History.RecentProjects) != 10 {
		t.Errorf("Expected 10 recent projects, got %d", len(config.History.RecentProjects))
	}
	
	// First project should be the most recent
	if config.History.RecentProjects[0] != "project-12" {
		t.Errorf("Expected first recent project 'project-12', got %s", config.History.RecentProjects[0])
	}
	
	// Last project should be project-3 (project-1 and project-2 should be removed)
	if config.History.RecentProjects[9] != "project-3" {
		t.Errorf("Expected last recent project 'project-3', got %s", config.History.RecentProjects[9])
	}
	
	if config.History.LastUsed != "project-12" {
		t.Errorf("Expected last used 'project-12', got %s", config.History.LastUsed)
	}
}
