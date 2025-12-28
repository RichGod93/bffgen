package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateUnitTests generates unit tests for controllers/resolvers
func (g *Generator) GenerateUnitTests() error {
	switch g.config.Language {
	case "nodejs-express", "nodejs-fastify":
		return g.generateNodeJSUnitTests()
	case "go":
		return g.generateGoUnitTests()
	case "python-fastapi":
		return g.generatePythonUnitTests()
	default:
		return fmt.Errorf("unsupported language: %s", g.config.Language)
	}
}

func (g *Generator) generateNodeJSUnitTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "unit")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Generate controller unit Tests
	testFile := filepath.Join(testDir, "controllers.test.js")

	tmpl := template.Must(template.New("unit").Parse(nodeJSUnitTemplate))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
		"Routes":      g.config.Routes,
	}

	return tmpl.Execute(f, data)
}

func (g *Generator) generateGoUnitTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "unit")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	testFile := filepath.Join(testDir, "handlers_test.go")

	tmpl := template.Must(template.New("unit").Parse(goUnitTemplate))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
	}

	return tmpl.Execute(f, data)
}

func (g *Generator) generatePythonUnitTests() error {
	testDir := filepath.Join(g.config.OutputDir, "tests", "unit")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Create __init__.py
	initFile := filepath.Join(testDir, "__init__.py")
	if err := os.WriteFile(initFile, []byte(""), 0644); err != nil {
		return err
	}

	testFile := filepath.Join(testDir, "test_routes.py")

	tmpl := template.Must(template.New("unit").Parse(pythonUnitTemplate))

	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}
	defer f.Close()

	data := map[string]interface{}{
		"ProjectName": g.config.ProjectName,
	}

	return tmpl.Execute(f, data)
}

const nodeJSUnitTemplate = `import { jest } from '@jest/globals';

describe('Controller Unit Tests', () => {
  describe('User Controller', () => {
    let mockService;
    let controller;

    beforeEach(() => {
      // Mock the service layer
      mockService = {
        getUsers: jest.fn(),
        getUserById: jest.fn(),
        createUser: jest.fn(),
        updateUser: jest.fn(),
        deleteUser: jest.fn(),
      };

      // TODO: Initialize controller with mock service
      // controller = new UserController(mockService);
    });

    it('should get all users', async () => {
      const mockUsers = [{ id: 1, name: 'Test User' }];
      mockService.getUsers.mockResolvedValue(mockUsers);

      // TODO: Call controller method and assert results
      // const result = await controller.getUsers();
      // expect(result).toEqual(mockUsers);
      // expect(mockService.getUsers).toHaveBeenCalledTimes(1);
    });

    it('should handle service errors', async () => {
      mockService.getUsers.mockRejectedValue(new Error('Service error'));

      // TODO: Test error handling
      // await expect(controller.getUsers()).rejects.toThrow('Service error');
    });
  });
});
`

const goUnitTemplate = `package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockService is a mock for the service layer
type MockService struct {
	mock.Mock
}

func (m *MockService) GetUsers() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func TestUserHandler(t *testing.T) {
	t.Run("GetUsers returns users from service", func(t *testing.T) {
		mockService := new(MockService)
		mockUsers := []interface{}{
			map[string]interface{}{"id": 1, "name": "Test User"},
		}
		
		mockService.On("GetUsers").Return(mockUsers, nil)

		// TODO: Initialize handler with mock service
		// handler := NewUserHandler(mockService)
		
		// TODO: Test handler logic
		// users, err := handler.GetUsers()
		// assert.NoError(t, err)
		// assert.Equal(t, mockUsers, users)
		// mockService.AssertExpectations(t)
	})
}
`

const pythonUnitTemplate = `import pytest
from unittest.mock import Mock, patch

class TestRoutes:
    """Unit tests for API routes"""

    @pytest.fixture
    def mock_service(self):
        """Create mock service"""
        return Mock()

    def test_get_users(self, mock_service):
        """Test getting all users"""
        mock_service.get_users.return_value = [
            {"id": 1, "name": "Test User"}
        ]

        # TODO: Initialize route handler with mock service
        # result = await get_users_handler(mock_service)
        # assert result == [{"id": 1, "name": "Test User"}]
        # mock_service.get_users.assert_called_once()

    def test_error_handling(self, mock_service):
        """Test error handling in routes"""
        mock_service.get_users.side_effect = Exception("Service error")

        # TODO: Test that route handles service errors
        # with pytest.raises(HTTPException):
        #     await get_users_handler(mock_service)
`
