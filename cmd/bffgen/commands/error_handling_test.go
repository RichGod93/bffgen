package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/templates"
	"github.com/RichGod93/bffgen/internal/types"
)

// Error handling tests for edge cases

func TestGenerateProxyRoutesCode_NilConfig(t *testing.T) {
	// Should handle nil config gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil config")
		}
	}()
	
	generateProxyRoutesCode(nil)
}

func TestChiMethod_EdgeCases(t *testing.T) {
	// Test with very long method name
	result := chiMethod("VERYLONGMETHODNAME")
	if result != "Get" {
		t.Errorf("Expected 'Get' for unknown method, got %q", result)
	}

	// Test with special characters
	result = chiMethod("GET@#$%")
	// Should still try to process
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestConvertToEndpointData_EmptyInput(t *testing.T) {
	result := convertToEndpointData([]map[string]interface{}{})
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d items", len(result))
	}
}

func TestConvertToEndpointData_NilInput(t *testing.T) {
	result := convertToEndpointData(nil)
	
	if result == nil {
		t.Error("Expected non-nil result for nil input")
	}
	
	if len(result) != 0 {
		t.Errorf("Expected empty result for nil input, got %d items", len(result))
	}
}

func TestGenerateGoModContent_EmptyProjectName(t *testing.T) {
	result := generateGoModContent("", "chi")
	
	// Should still generate valid go.mod structure
	if result == "" {
		t.Error("Expected non-empty result for empty project name")
	}
	
	// Should contain go version
	if !strings.Contains(result, "go 1.21") {
		t.Error("Expected go version even with empty project name")
	}
}

func TestGeneratePackageJsonContent_EmptyProjectName(t *testing.T) {
	result := generatePackageJsonContent("", scaffolding.LanguageNodeExpress, "express")
	
	// Should still generate structure with empty name
	if !strings.Contains(result, `"name"`) {
		t.Error("Expected name field even with empty project name")
	}
}

func TestGenerateCORSConfig_EmptyFramework(t *testing.T) {
	result := generateCORSConfig([]string{"http://localhost:3000"}, "")
	
	// Should return empty or default for unknown framework
	_ = result // Just verify it doesn't panic
}

func TestRenderControllerTemplate_InvalidFramework(t *testing.T) {
	loader := templates.NewTemplateLoader(scaffolding.LanguageNodeExpress)
	data := &templates.ControllerTemplateData{
		ServiceName: "test",
		Endpoints:   []templates.EndpointData{},
	}
	
	// Try with invalid template name
	_, err := renderControllerTemplate(loader, "invalid-framework", "invalid-template.js", data)
	if err == nil {
		t.Error("Expected error for invalid framework/template")
	}
}

func TestRenderServiceTemplate_InvalidFramework(t *testing.T) {
	loader := templates.NewTemplateLoader(scaffolding.LanguageNodeExpress)
	data := &templates.ServiceTemplateData{
		ServiceName: "test",
		BaseURL:     "http://localhost:4000",
	}
	
	// Try with invalid template name
	_, err := renderServiceTemplate(loader, "invalid-framework", "invalid-template.js", data)
	if err == nil {
		t.Error("Expected error for invalid framework/template")
	}
}

func TestCreateProjectDirectories_EmptyProjectName(t *testing.T) {
	// Empty project name should error or handle gracefully
	err := createProjectDirectories("", scaffolding.LanguageGo)
	// Either succeeds (creates in current dir) or fails - both acceptable
	_ = err // Just ensure no panic
}

func TestGenerateProxyRoutesCode_EmptyServices(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{},
	}
	
	result := generateProxyRoutesCode(config)
	
	// Should have header but no routes
	if !strings.Contains(result, "Generated proxy routes") {
		t.Error("Expected header comment")
	}
}

func TestGenerateProxyRoutesCode_ServiceWithNoEndpoints(t *testing.T) {
	config := &types.BFFConfig{
		Services: map[string]types.Service{
			"empty": {
				BaseURL:   "http://localhost:4000",
				Endpoints: []types.Endpoint{},
			},
		},
	}
	
	result := generateProxyRoutesCode(config)
	
	// Should handle service with no endpoints
	if !strings.Contains(result, "empty service routes") {
		t.Error("Expected service comment even with no endpoints")
	}
}

func TestChiMethod_AllHTTPMethods(t *testing.T) {
	methods := map[string]string{
		"GET":     "Get",
		"POST":    "Post",
		"PUT":     "Put",
		"DELETE":  "Delete",
		"PATCH":   "Patch",
		"HEAD":    "Head",
		"OPTIONS": "Options",
		"CONNECT": "Get", // Unknown
		"TRACE":   "Get", // Unknown
	}
	
	for input, expected := range methods {
		result := chiMethod(input)
		if result != expected {
			t.Errorf("chiMethod(%q) = %q; expected %q", input, result, expected)
		}
	}
}

func TestGenerateCORSConfig_SpecialCharactersInOrigins(t *testing.T) {
	origins := []string{
		"http://localhost:3000",
		"https://example.com:8443",
		"http://192.168.1.1:3000",
	}
	
	result := generateCORSConfig(origins, "chi")
	
	// Should handle all origins
	for _, origin := range origins {
		if !strings.Contains(result, origin) {
			t.Errorf("Expected origin %s in result", origin)
		}
	}
}

func TestGenerateGoModContent_SpecialCharactersInProjectName(t *testing.T) {
	// Test with special characters (might not be valid but shouldn't crash)
	projectNames := []string{
		"test-project",
		"test_project",
		"test.project",
		"test/project",
	}
	
	for _, name := range projectNames {
		result := generateGoModContent(name, "chi")
		if result == "" {
			t.Errorf("Expected non-empty result for project name %q", name)
		}
	}
}

func TestConvertToEndpointData_MixedTypes(t *testing.T) {
	// Mix of valid and invalid types
	input := []map[string]interface{}{
		{
			"Path":   "/valid",
			"Method": "GET",
		},
		{
			"Path":   123, // Wrong type
			"Method": "POST",
		},
		{
			"Path":   "/another",
			"Method": []string{"GET"}, // Wrong type
		},
	}
	
	result := convertToEndpointData(input)
	
	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}
	
	// First should be valid
	if result[0].Path != "/valid" {
		t.Error("First endpoint should have valid path")
	}
	
	// Second should have empty path (type mismatch)
	if result[1].Path != "" {
		t.Error("Second endpoint should have empty path due to type mismatch")
	}
}

func TestGeneratePackageJsonContent_UnsupportedFramework(t *testing.T) {
	frameworks := []string{
		"unknown",
		"koa",
		"hapi",
		"",
	}
	
	for _, fw := range frameworks {
		result := generatePackageJsonContent("test", scaffolding.LanguageNodeExpress, fw)
		// Unknown frameworks return empty string
		if fw != "express" && fw != "fastify" && result != "" {
			t.Errorf("Expected empty result for unsupported framework %q", fw)
		}
	}
}

func TestGenerateCORSConfig_ManyOrigins(t *testing.T) {
	// Test with many origins
	origins := make([]string, 50)
	for i := 0; i < 50; i++ {
		origins[i] = fmt.Sprintf("http://localhost:%d", 3000+i)
	}
	
	result := generateCORSConfig(origins, "chi")
	
	// Should handle many origins without error
	if result == "" {
		t.Error("Expected non-empty result for many origins")
	}
	
	// Check some origins are present
	if !strings.Contains(result, "http://localhost:3000") {
		t.Error("Expected first origin in result")
	}
	if !strings.Contains(result, "http://localhost:3049") {
		t.Error("Expected last origin in result")
	}
}

