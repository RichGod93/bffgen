package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerationState tracks the state of generated files
type GenerationState struct {
	Version        string                    `json:"version"`
	LastGeneration time.Time                 `json:"lastGeneration"`
	GeneratedFiles map[string]*GeneratedFile `json:"generatedFiles"`
	Routes         map[string]*RouteState    `json:"routes"`
	ProjectType    string                    `json:"projectType"` // "go", "nodejs"
	Framework      string                    `json:"framework"`   // "chi", "express", "fastify", etc.
}

// GeneratedFile represents a file that was generated
type GeneratedFile struct {
	Path         string    `json:"path"`
	GeneratedAt  time.Time `json:"generatedAt"`
	Hash         string    `json:"hash"` // SHA256 hash of generated content
	HasMarkers   bool      `json:"hasMarkers"`
	UserModified bool      `json:"userModified"`
}

// RouteState represents the state of a route
type RouteState struct {
	Service   string    `json:"service"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	ExposeAs  string    `json:"exposeAs"`
	Generated time.Time `json:"generated"`
}

const (
	stateDir       = ".bffgen"
	stateFile      = "state.json"
	backupDir      = "backup"
	currentVersion = "2.0.0"
)

// LoadState loads the generation state from .bffgen/state.json
func LoadState() (*GenerationState, error) {
	statePath := filepath.Join(stateDir, stateFile)

	// If state file doesn't exist, return empty state
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &GenerationState{
			Version:        currentVersion,
			LastGeneration: time.Now(),
			GeneratedFiles: make(map[string]*GeneratedFile),
			Routes:         make(map[string]*RouteState),
		}, nil
	}

	// Read state file
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state GenerationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// SaveState saves the generation state to .bffgen/state.json
func SaveState(state *GenerationState) error {
	// Create .bffgen directory if it doesn't exist
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Update last generation time
	state.LastGeneration = time.Now()
	state.Version = currentVersion

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write state file
	statePath := filepath.Join(stateDir, stateFile)
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// TrackGeneratedFile adds or updates a generated file in the state
func (s *GenerationState) TrackGeneratedFile(path, contentHash string, hasMarkers bool) {
	if s.GeneratedFiles == nil {
		s.GeneratedFiles = make(map[string]*GeneratedFile)
	}

	s.GeneratedFiles[path] = &GeneratedFile{
		Path:         path,
		GeneratedAt:  time.Now(),
		Hash:         contentHash,
		HasMarkers:   hasMarkers,
		UserModified: false,
	}
}

// TrackRoute adds or updates a route in the state
func (s *GenerationState) TrackRoute(service, method, path, exposeAs string) {
	if s.Routes == nil {
		s.Routes = make(map[string]*RouteState)
	}

	routeKey := fmt.Sprintf("%s:%s:%s", service, method, exposeAs)
	s.Routes[routeKey] = &RouteState{
		Service:   service,
		Method:    method,
		Path:      path,
		ExposeAs:  exposeAs,
		Generated: time.Now(),
	}
}

// IsRouteGenerated checks if a route has already been generated
func (s *GenerationState) IsRouteGenerated(service, method, exposeAs string) bool {
	if s.Routes == nil {
		return false
	}

	routeKey := fmt.Sprintf("%s:%s:%s", service, method, exposeAs)
	_, exists := s.Routes[routeKey]
	return exists
}

// IsFileGenerated checks if a file has been generated before
func (s *GenerationState) IsFileGenerated(path string) bool {
	if s.GeneratedFiles == nil {
		return false
	}

	_, exists := s.GeneratedFiles[path]
	return exists
}

// GetBackupDir returns the path to the backup directory
func GetBackupDir() string {
	return filepath.Join(stateDir, backupDir)
}

// CleanupOldBackups removes backups older than specified duration
func CleanupOldBackups(olderThan time.Duration) error {
	backupPath := GetBackupDir()

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil // No backup directory, nothing to clean
	}

	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > olderThan {
			filePath := filepath.Join(backupPath, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// InitializeState initializes a new generation state for a project
func InitializeState(projectType, framework string) (*GenerationState, error) {
	state := &GenerationState{
		Version:        currentVersion,
		LastGeneration: time.Now(),
		GeneratedFiles: make(map[string]*GeneratedFile),
		Routes:         make(map[string]*RouteState),
		ProjectType:    projectType,
		Framework:      framework,
	}

	if err := SaveState(state); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	return state, nil
}

// GetStateDir returns the state directory path
func GetStateDir() string {
	return stateDir
}

// GetCurrentTimestamp returns the current timestamp as a string
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
