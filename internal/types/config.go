package types

// BFFConfig represents the complete BFF configuration
type BFFConfig struct {
	Services map[string]Service `yaml:"services"`
	Settings Settings           `yaml:"settings"`
}

// Service represents a backend service configuration
type Service struct {
	BaseURL   string     `yaml:"baseUrl"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

// Endpoint represents a single API endpoint
type Endpoint struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	Method   string `yaml:"method"`
	ExposeAs string `yaml:"exposeAs"`
}

// Settings represents global BFF settings
type Settings struct {
	Port    int    `yaml:"port"`
	Timeout string `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}
