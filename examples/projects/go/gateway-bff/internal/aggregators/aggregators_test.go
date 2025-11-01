package aggregators

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDashboard(t *testing.T) {
	// Create mock servers
	usersServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    "1",
			"name":  "Test User",
			"email": "test@example.com",
		})
	}))
	defer usersServer.Close()

	analyticsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metrics": []map[string]interface{}{
				{"name": "active_users", "value": 100},
			},
		})
	}))
	defer analyticsServer.Close()

	notificationsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notifications": []map[string]interface{}{
				{"id": "1", "message": "Test notification"},
			},
			"unread": 1,
		})
	}))
	defer notificationsServer.Close()

	// Test successful aggregation
	baseURLs := map[string]string{
		"users":         usersServer.URL,
		"analytics":     analyticsServer.URL,
		"notifications": notificationsServer.URL,
	}

	dashboard, err := GetDashboard("1", baseURLs)
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dashboard.User == nil {
		t.Error("Expected user data, got nil")
	}

	if dashboard.Metrics == nil {
		t.Error("Expected metrics data, got nil")
	}

	if dashboard.Notifications == nil {
		t.Error("Expected notifications data, got nil")
	}

	if dashboard.Timestamp == "" {
		t.Error("Expected timestamp, got empty string")
	}
}

func TestGetDashboard_RequiredServiceFailure(t *testing.T) {
	// Test with invalid user service URL (should fail since users is required)
	baseURLs := map[string]string{
		"users":         "http://invalid-url",
		"analytics":     "http://localhost:4001/api",
		"notifications": "http://localhost:4002/api",
	}

	_, err := GetDashboard("1", baseURLs)
	if err == nil {
		t.Error("Expected error when required service fails, got nil")
	}
}

func TestGetDashboard_OptionalServiceFailure(t *testing.T) {
	// Create mock users server (required service)
	usersServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "1",
			"name": "Test User",
		})
	}))
	defer usersServer.Close()

	// Test with invalid optional service URLs (should succeed)
	baseURLs := map[string]string{
		"users":         usersServer.URL,
		"analytics":     "http://invalid-analytics-url",
		"notifications": "http://invalid-notifications-url",
	}

	dashboard, err := GetDashboard("1", baseURLs)
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}

	if dashboard.User == nil {
		t.Error("Expected user data, got nil")
	}

	// Optional services should be nil when they fail
	if dashboard.Metrics != nil {
		t.Error("Expected nil metrics when service fails")
	}

	if dashboard.Notifications != nil {
		t.Error("Expected nil notifications when service fails")
	}
}
