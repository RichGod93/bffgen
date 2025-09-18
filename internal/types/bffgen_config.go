package types

// BFFGenConfig represents the global bffgen configuration
type BFFGenConfig struct {
	Defaults Defaults `yaml:"defaults"`
	User     User     `yaml:"user"`
	History  History  `yaml:"history"`
}

// Defaults represents default settings for new projects
type Defaults struct {
	Framework     string   `yaml:"framework"`     // chi, echo, fiber
	CORSOrigins   []string `yaml:"cors_origins"`  // Default CORS origins
	JWTSecret     string   `yaml:"jwt_secret"`    // Default JWT secret
	RedisURL      string   `yaml:"redis_url"`     // Default Redis URL
	Port          int      `yaml:"port"`          // Default port
	RouteOption   string   `yaml:"route_option"`  // 1=manual, 2=template, 3=skip
}

// User represents user information
type User struct {
	Name    string `yaml:"name"`
	Email   string `yaml:"email"`
	GitHub  string `yaml:"github"`
	Company string `yaml:"company"`
}

// History represents recent project history
type History struct {
	RecentProjects []string `yaml:"recent_projects"`
	LastUsed       string   `yaml:"last_used"`
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *BFFGenConfig {
	return &BFFGenConfig{
		Defaults: Defaults{
			Framework:   "chi",
			CORSOrigins: []string{"localhost:3000", "localhost:3001"},
			JWTSecret:   "your-secret-key-change-in-production",
			RedisURL:    "redis://localhost:6379",
			Port:        8080,
			RouteOption: "3", // Skip by default
		},
		User: User{
			Name:   "",
			Email:  "",
			GitHub: "",
		},
		History: History{
			RecentProjects: []string{},
			LastUsed:       "",
		},
	}
}
