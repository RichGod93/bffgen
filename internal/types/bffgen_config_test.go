package types

import (
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	
	// Test Defaults
	if config.Defaults.Framework != "chi" {
		t.Errorf("Expected framework 'chi', got %s", config.Defaults.Framework)
	}
	
	if len(config.Defaults.CORSOrigins) != 2 {
		t.Errorf("Expected 2 CORS origins, got %d", len(config.Defaults.CORSOrigins))
	}
	
	expectedOrigins := []string{"localhost:3000", "localhost:3001"}
	for i, origin := range config.Defaults.CORSOrigins {
		if origin != expectedOrigins[i] {
			t.Errorf("Expected CORS origin %s, got %s", expectedOrigins[i], origin)
		}
	}
	
	if config.Defaults.JWTSecret != "your-secret-key-change-in-production" {
		t.Errorf("Expected JWT secret 'your-secret-key-change-in-production', got %s", config.Defaults.JWTSecret)
	}
	
	if config.Defaults.RedisURL != "redis://localhost:6379" {
		t.Errorf("Expected Redis URL 'redis://localhost:6379', got %s", config.Defaults.RedisURL)
	}
	
	if config.Defaults.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Defaults.Port)
	}
	
	if config.Defaults.RouteOption != "3" {
		t.Errorf("Expected route option '3', got %s", config.Defaults.RouteOption)
	}
	
	// Test User
	if config.User.Name != "" {
		t.Errorf("Expected empty user name, got %s", config.User.Name)
	}
	
	if config.User.Email != "" {
		t.Errorf("Expected empty user email, got %s", config.User.Email)
	}
	
	if config.User.GitHub != "" {
		t.Errorf("Expected empty user GitHub, got %s", config.User.GitHub)
	}
	
	// Test History
	if len(config.History.RecentProjects) != 0 {
		t.Errorf("Expected empty recent projects, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.LastUsed != "" {
		t.Errorf("Expected empty last used, got %s", config.History.LastUsed)
	}
}

func TestBFFGenConfig_Structure(t *testing.T) {
	config := &BFFGenConfig{
		Defaults: Defaults{
			Framework:   "echo",
			CORSOrigins: []string{"https://example.com"},
			JWTSecret:   "test-secret",
			RedisURL:    "redis://test:6379",
			Port:        3000,
			RouteOption: "1",
		},
		User: User{
			Name:    "Test User",
			Email:   "test@example.com",
			GitHub:  "testuser",
			Company: "Test Company",
		},
		History: History{
			RecentProjects: []string{"project1", "project2"},
			LastUsed:       "project1",
		},
	}
	
	// Test Defaults
	if config.Defaults.Framework != "echo" {
		t.Errorf("Expected framework 'echo', got %s", config.Defaults.Framework)
	}
	
	if len(config.Defaults.CORSOrigins) != 1 {
		t.Errorf("Expected 1 CORS origin, got %d", len(config.Defaults.CORSOrigins))
	}
	
	if config.Defaults.CORSOrigins[0] != "https://example.com" {
		t.Errorf("Expected CORS origin 'https://example.com', got %s", config.Defaults.CORSOrigins[0])
	}
	
	if config.Defaults.JWTSecret != "test-secret" {
		t.Errorf("Expected JWT secret 'test-secret', got %s", config.Defaults.JWTSecret)
	}
	
	if config.Defaults.RedisURL != "redis://test:6379" {
		t.Errorf("Expected Redis URL 'redis://test:6379', got %s", config.Defaults.RedisURL)
	}
	
	if config.Defaults.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", config.Defaults.Port)
	}
	
	if config.Defaults.RouteOption != "1" {
		t.Errorf("Expected route option '1', got %s", config.Defaults.RouteOption)
	}
	
	// Test User
	if config.User.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got %s", config.User.Name)
	}
	
	if config.User.Email != "test@example.com" {
		t.Errorf("Expected user email 'test@example.com', got %s", config.User.Email)
	}
	
	if config.User.GitHub != "testuser" {
		t.Errorf("Expected user GitHub 'testuser', got %s", config.User.GitHub)
	}
	
	if config.User.Company != "Test Company" {
		t.Errorf("Expected user company 'Test Company', got %s", config.User.Company)
	}
	
	// Test History
	if len(config.History.RecentProjects) != 2 {
		t.Errorf("Expected 2 recent projects, got %d", len(config.History.RecentProjects))
	}
	
	if config.History.RecentProjects[0] != "project1" {
		t.Errorf("Expected first recent project 'project1', got %s", config.History.RecentProjects[0])
	}
	
	if config.History.RecentProjects[1] != "project2" {
		t.Errorf("Expected second recent project 'project2', got %s", config.History.RecentProjects[1])
	}
	
	if config.History.LastUsed != "project1" {
		t.Errorf("Expected last used 'project1', got %s", config.History.LastUsed)
	}
}

func TestDefaults_Structure(t *testing.T) {
	defaults := Defaults{
		Framework:   "fiber",
		CORSOrigins: []string{"http://localhost:3000", "https://app.example.com"},
		JWTSecret:   "super-secret-key",
		RedisURL:    "redis://production:6379",
		Port:        9000,
		RouteOption: "2",
	}
	
	if defaults.Framework != "fiber" {
		t.Errorf("Expected framework 'fiber', got %s", defaults.Framework)
	}
	
	if len(defaults.CORSOrigins) != 2 {
		t.Errorf("Expected 2 CORS origins, got %d", len(defaults.CORSOrigins))
	}
	
	expectedOrigins := []string{"http://localhost:3000", "https://app.example.com"}
	for i, origin := range defaults.CORSOrigins {
		if origin != expectedOrigins[i] {
			t.Errorf("Expected CORS origin %s, got %s", expectedOrigins[i], origin)
		}
	}
	
	if defaults.JWTSecret != "super-secret-key" {
		t.Errorf("Expected JWT secret 'super-secret-key', got %s", defaults.JWTSecret)
	}
	
	if defaults.RedisURL != "redis://production:6379" {
		t.Errorf("Expected Redis URL 'redis://production:6379', got %s", defaults.RedisURL)
	}
	
	if defaults.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", defaults.Port)
	}
	
	if defaults.RouteOption != "2" {
		t.Errorf("Expected route option '2', got %s", defaults.RouteOption)
	}
}

func TestUser_Structure(t *testing.T) {
	user := User{
		Name:    "John Doe",
		Email:   "john.doe@company.com",
		GitHub:  "johndoe",
		Company: "Acme Corp",
	}
	
	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %s", user.Name)
	}
	
	if user.Email != "john.doe@company.com" {
		t.Errorf("Expected email 'john.doe@company.com', got %s", user.Email)
	}
	
	if user.GitHub != "johndoe" {
		t.Errorf("Expected GitHub 'johndoe', got %s", user.GitHub)
	}
	
	if user.Company != "Acme Corp" {
		t.Errorf("Expected company 'Acme Corp', got %s", user.Company)
	}
}

func TestHistory_Structure(t *testing.T) {
	history := History{
		RecentProjects: []string{"project-a", "project-b", "project-c"},
		LastUsed:       "project-a",
	}
	
	if len(history.RecentProjects) != 3 {
		t.Errorf("Expected 3 recent projects, got %d", len(history.RecentProjects))
	}
	
	expectedProjects := []string{"project-a", "project-b", "project-c"}
	for i, project := range history.RecentProjects {
		if project != expectedProjects[i] {
			t.Errorf("Expected project %s, got %s", expectedProjects[i], project)
		}
	}
	
	if history.LastUsed != "project-a" {
		t.Errorf("Expected last used 'project-a', got %s", history.LastUsed)
	}
}
