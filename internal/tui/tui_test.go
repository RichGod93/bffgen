package tui

import (
	"testing"

	"github.com/RichGod93/bffgen/internal/templates"
)

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single origin with http",
			input:    "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "single origin without scheme",
			input:    "localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "multiple origins",
			input:    "http://localhost:3000,https://example.com",
			expected: []string{"http://localhost:3000", "https://example.com"},
		},
		{
			name:     "origins with spaces",
			input:    "http://localhost:3000 , https://example.com",
			expected: []string{"http://localhost:3000", "https://example.com"},
		},
		{
			name:     "mixed schemes",
			input:    "localhost:3000,https://secure.com",
			expected: []string{"http://localhost:3000", "https://secure.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d origins, got %d", len(tt.expected), len(result))
				return
			}
			for i, origin := range result {
				if origin != tt.expected[i] {
					t.Errorf("Expected origin[%d] = %q, got %q", i, tt.expected[i], origin)
				}
			}
		})
	}
}

func TestTestTypeItem(t *testing.T) {
	item := TestTypeItem{
		title: "Test Title",
		desc:  "Test Description",
		value: "test-value",
	}

	t.Run("FilterValue returns title", func(t *testing.T) {
		if item.FilterValue() != "Test Title" {
			t.Errorf("Expected FilterValue() = 'Test Title', got %q", item.FilterValue())
		}
	})

	t.Run("Title returns title", func(t *testing.T) {
		if item.Title() != "Test Title" {
			t.Errorf("Expected Title() = 'Test Title', got %q", item.Title())
		}
	})

	t.Run("Description returns desc", func(t *testing.T) {
		if item.Description() != "Test Description" {
			t.Errorf("Expected Description() = 'Test Description', got %q", item.Description())
		}
	})
}

func TestTemplateSelectorModel(t *testing.T) {
	// Test with empty template list
	t.Run("NewTemplateSelector with empty list", func(t *testing.T) {
		model := NewTemplateSelector([]*templates.Template{})
		if model.choice != nil {
			t.Error("Expected nil choice initially")
		}
	})

	// Test with templates
	t.Run("NewTemplateSelector with templates", func(t *testing.T) {
		templateList := []*templates.Template{
			{Name: "template1", Version: "1.0.0", Language: "go"},
			{Name: "template2", Version: "1.0.0", Language: "nodejs-express"},
		}
		model := NewTemplateSelector(templateList)
		if model.quitting {
			t.Error("Expected quitting to be false initially")
		}
	})
}

func TestTemplateSelectorModel_Init(t *testing.T) {
	model := NewTemplateSelector([]*templates.Template{})
	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestTemplateSelectorModel_View(t *testing.T) {
	model := NewTemplateSelector([]*templates.Template{
		{Name: "test-template", Version: "1.0.0", Language: "go"},
	})
	view := model.View()
	if len(view) == 0 {
		t.Error("View() should return non-empty string")
	}
}

func TestRunTemplateSelector_EmptyList(t *testing.T) {
	_, err := RunTemplateSelector([]*templates.Template{})
	if err == nil {
		t.Error("Expected error for empty template list")
	}
}

func TestNewTestSelector(t *testing.T) {
	model := NewTestSelector()

	t.Run("initializes with empty choice", func(t *testing.T) {
		if model.choice != "" {
			t.Errorf("Expected empty choice, got %q", model.choice)
		}
	})

	t.Run("initializes with quitting false", func(t *testing.T) {
		if model.quitting {
			t.Error("Expected quitting to be false")
		}
	})

	t.Run("GetChoice returns empty string initially", func(t *testing.T) {
		if model.GetChoice() != "" {
			t.Errorf("Expected GetChoice() = '', got %q", model.GetChoice())
		}
	})
}

func TestTestSelectorModel_Init(t *testing.T) {
	model := NewTestSelector()
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestTestSelectorModel_View(t *testing.T) {
	model := NewTestSelector()
	view := model.View()

	if len(view) == 0 {
		t.Error("View() should return non-empty string")
	}
}

func TestTestDelegate(t *testing.T) {
	delegate := testDelegate{}

	t.Run("Height returns 2", func(t *testing.T) {
		if delegate.Height() != 2 {
			t.Errorf("Expected Height() = 2, got %d", delegate.Height())
		}
	})

	t.Run("Spacing returns 1", func(t *testing.T) {
		if delegate.Spacing() != 1 {
			t.Errorf("Expected Spacing() = 1, got %d", delegate.Spacing())
		}
	})
}
