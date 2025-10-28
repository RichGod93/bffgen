package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/internal/types"
)

func TestGenerateProxyRoutesCode(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"users": {
				BaseURL: "http://localhost:4000/api",
				Endpoints: []types.Endpoint{
					{
						Path:     "/users",
						Method:   "GET",
						ExposeAs: "/api/users",
					},
					{
						Path:     "/users/:id",
						Method:   "GET",
						ExposeAs: "/api/users/:id",
					},
					{
						Path:     "/users",
						Method:   "POST",
						ExposeAs: "/api/users",
					},
				},
			},
			"products": {
				BaseURL: "http://localhost:5000/api",
				Endpoints: []types.Endpoint{
					{
						Path:     "/products",
						Method:   "GET",
						ExposeAs: "/api/products",
					},
				},
			},
		},
	}

	result := generateProxyRoutesCode(config)

	// Check that result contains expected patterns
	if !strings.Contains(result, "users service routes") {
		t.Error("Expected 'users service routes' comment")
	}
	if !strings.Contains(result, "products service routes") {
		t.Error("Expected 'products service routes' comment")
	}
	if !strings.Contains(result, `r.Get`) {
		t.Error("Expected r.Get method call")
	}
	if !strings.Contains(result, `r.Post`) {
		t.Error("Expected r.Post method call")
	}
	if !strings.Contains(result, `"/api/users"`) {
		t.Error("Expected API route path")
	}
	if !strings.Contains(result, `createProxyHandler`) {
		t.Error("Expected createProxyHandler function call")
	}
}

func TestChiMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"GET", "GET", "Get"},
		{"POST", "POST", "Post"},
		{"PUT", "PUT", "Put"},
		{"DELETE", "DELETE", "Delete"},
		{"PATCH", "PATCH", "Patch"},
		{"HEAD", "HEAD", "Head"},
		{"OPTIONS", "OPTIONS", "Options"},
		{"lowercase", "get", "Get"},
		{"MixedCase", "GeT", "Get"},
		{"Unknown", "UNKNOWN", "Get"},
		{"Empty", "", "Get"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chiMethod(tt.input)
			if result != tt.expected {
				t.Errorf("chiMethod(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToEndpointData(t *testing.T) {
	input := []map[string]interface{}{
		{
			"Path":              "/users",
			"Method":            "GET",
			"BackendPath":       "/users",
			"RequiresAuth":      false,
			"HandlerName":       "getUsers",
			"HandlerNamePascal": "GetUsers",
		},
		{
			"Path":              "/users/:id",
			"Method":            "GET",
			"BackendPath":       "/users/:id",
			"RequiresAuth":      true,
			"HandlerName":       "getUserById",
			"HandlerNamePascal": "GetUserById",
		},
	}

	result := convertToEndpointData(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(result))
	}

	// Check first endpoint
	if result[0].Path != "/users" {
		t.Errorf("Expected Path='/users', got %q", result[0].Path)
	}
	if result[0].Method != "GET" {
		t.Errorf("Expected Method='GET', got %q", result[0].Method)
	}
	if result[0].RequiresAuth != false {
		t.Errorf("Expected RequiresAuth=false, got %v", result[0].RequiresAuth)
	}

	// Check second endpoint
	if result[1].Path != "/users/:id" {
		t.Errorf("Expected Path='/users/:id', got %q", result[1].Path)
	}
	if result[1].RequiresAuth != true {
		t.Errorf("Expected RequiresAuth=true, got %v", result[1].RequiresAuth)
	}
}

func TestConvertToEndpointDataWithMissingFields(t *testing.T) {
	input := []map[string]interface{}{
		{
			"Path": "/users",
			// Missing other fields
		},
		{}, // Empty map
	}

	result := convertToEndpointData(input)

	// Should handle missing fields gracefully
	if len(result) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(result))
	}

	// First should have Path
	if result[0].Path != "/users" {
		t.Errorf("Expected Path='/users', got %q", result[0].Path)
	}

	// Second should be empty but still exist
	if result[1].Path != "" {
		t.Errorf("Expected empty Path, got %q", result[1].Path)
	}
}

func TestConvertToEndpointDataWithWrongTypes(t *testing.T) {
	input := []map[string]interface{}{
		{
			"Path":        123,             // Wrong type
			"Method":      []string{"GET"}, // Wrong type
			"BackendPath": true,            // Wrong type
		},
	}

	result := convertToEndpointData(input)

	// Should handle wrong types gracefully
	if len(result) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(result))
	}

	// All fields should be empty due to type mismatch
	if result[0].Path != "" {
		t.Errorf("Expected empty Path, got %q", result[0].Path)
	}
	if result[0].Method != "" {
		t.Errorf("Expected empty Method, got %q", result[0].Method)
	}
}

func TestGenerateProxyRoutesCode_EmptyConfig(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{},
	}

	result := generateProxyRoutesCode(config)

	// Should still have the header comment
	if !strings.Contains(result, "Generated proxy routes") {
		t.Error("Expected 'Generated proxy routes' header")
	}

	// But no service routes
	if strings.Contains(result, "service routes") {
		t.Error("Should not have any service routes")
	}
}

func TestGenerateProxyRoutesCode_SingleEndpoint(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"health": {
				BaseURL: "http://localhost:3000/api",
				Endpoints: []types.Endpoint{
					{
						Path:     "/health",
						Method:   "GET",
						ExposeAs: "/health",
					},
				},
			},
		},
	}

	result := generateProxyRoutesCode(config)

	// Check basic structure
	if !strings.Contains(result, "health service routes") {
		t.Error("Expected health service routes")
	}
	if !strings.Contains(result, `r.Get`) {
		t.Error("Expected r.Get method")
	}
	if !strings.Contains(result, `"/health"`) {
		t.Error("Expected /health path")
	}
}

func TestGenerateProxyRoutesCode_MultipleServices(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"service1": {
				BaseURL: "http://localhost:4000/api",
				Endpoints: []types.Endpoint{
					{Path: "/endpoint1", Method: "GET", ExposeAs: "/api/endpoint1"},
				},
			},
			"service2": {
				BaseURL: "http://localhost:5000/api",
				Endpoints: []types.Endpoint{
					{Path: "/endpoint2", Method: "POST", ExposeAs: "/api/endpoint2"},
				},
			},
			"service3": {
				BaseURL: "http://localhost:6000/api",
				Endpoints: []types.Endpoint{
					{Path: "/endpoint3", Method: "PUT", ExposeAs: "/api/endpoint3"},
				},
			},
		},
	}

	result := generateProxyRoutesCode(config)

	// Should contain all three services
	services := []string{"service1", "service2", "service3"}
	for _, svc := range services {
		if !strings.Contains(result, fmt.Sprintf("%s service routes", svc)) {
			t.Errorf("Expected %s service routes", svc)
		}
	}

	// Should contain all HTTP methods
	if !strings.Contains(result, "r.Get") {
		t.Error("Expected r.Get")
	}
	if !strings.Contains(result, "r.Post") {
		t.Error("Expected r.Post")
	}
	if !strings.Contains(result, "r.Put") {
		t.Error("Expected r.Put")
	}
}
