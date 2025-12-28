package types

// BackendService represents a backend service configuration
// This is a shared type used by both commands and TUI to avoid import cycles
type BackendService struct {
	Name      string
	BaseURL   string
	Port      int
	Path      string
	Endpoints []string
}
