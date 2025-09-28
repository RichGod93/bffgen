package types

// BFFGenV1Config represents the v1 configuration schema
type BFFGenV1Config struct {
	Version     string               `yaml:"version" json:"version"`
	Project     ProjectConfig        `yaml:"project" json:"project"`
	Server      *ServerConfig        `yaml:"server,omitempty" json:"server,omitempty"`
	Auth        *AuthConfig          `yaml:"auth,omitempty" json:"auth,omitempty"`
	CORS        *CORSConfig          `yaml:"cors,omitempty" json:"cors,omitempty"`
	Security    *SecurityConfig      `yaml:"security,omitempty" json:"security,omitempty"`
	Logging     *LoggingConfig       `yaml:"logging,omitempty" json:"logging,omitempty"`
	Monitoring  *MonitoringConfig    `yaml:"monitoring,omitempty" json:"monitoring,omitempty"`
	Services    map[string]ServiceV1 `yaml:"services,omitempty" json:"services,omitempty"`
	Aggregators []AggregatorConfig   `yaml:"aggregators,omitempty" json:"aggregators,omitempty"`
	Middleware  []MiddlewareConfig   `yaml:"middleware,omitempty" json:"middleware,omitempty"`
	Environment *EnvironmentConfig   `yaml:"environment,omitempty" json:"environment,omitempty"`
	Build       *BuildConfig         `yaml:"build,omitempty" json:"build,omitempty"`
	Deployment  *DeploymentConfig    `yaml:"deployment,omitempty" json:"deployment,omitempty"`
}

// ProjectConfig represents project-level configuration
type ProjectConfig struct {
	Name        string        `yaml:"name" json:"name"`
	Description string        `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string        `yaml:"version,omitempty" json:"version,omitempty"`
	Language    string        `yaml:"language,omitempty" json:"language,omitempty"`
	Framework   string        `yaml:"framework" json:"framework"`
	Output      *OutputConfig `yaml:"output,omitempty" json:"output,omitempty"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	Package   string `yaml:"package,omitempty" json:"package,omitempty"`
	Module    string `yaml:"module,omitempty" json:"module,omitempty"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port             int                     `yaml:"port,omitempty" json:"port,omitempty"`
	Host             string                  `yaml:"host,omitempty" json:"host,omitempty"`
	Timeout          *TimeoutConfig          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	GracefulShutdown *GracefulShutdownConfig `yaml:"graceful_shutdown,omitempty" json:"graceful_shutdown,omitempty"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	Read  string `yaml:"read,omitempty" json:"read,omitempty"`
	Write string `yaml:"write,omitempty" json:"write,omitempty"`
	Idle  string `yaml:"idle,omitempty" json:"idle,omitempty"`
}

// GracefulShutdownConfig represents graceful shutdown configuration
type GracefulShutdownConfig struct {
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Mode    string         `yaml:"mode,omitempty" json:"mode,omitempty"`
	JWT     *JWTConfig     `yaml:"jwt,omitempty" json:"jwt,omitempty"`
	Session *SessionConfig `yaml:"session,omitempty" json:"session,omitempty"`
	CSRF    *CSRFConfig    `yaml:"csrf,omitempty" json:"csrf,omitempty"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret            string            `yaml:"secret,omitempty" json:"secret,omitempty"`
	Expiration        string            `yaml:"expiration,omitempty" json:"expiration,omitempty"`
	RefreshExpiration string            `yaml:"refresh_expiration,omitempty" json:"refresh_expiration,omitempty"`
	Encryption        *EncryptionConfig `yaml:"encryption,omitempty" json:"encryption,omitempty"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Enabled   bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Algorithm string `yaml:"algorithm,omitempty" json:"algorithm,omitempty"`
}

// SessionConfig represents session configuration
type SessionConfig struct {
	Store      string `yaml:"store,omitempty" json:"store,omitempty"`
	Expiration string `yaml:"expiration,omitempty" json:"expiration,omitempty"`
	Secure     bool   `yaml:"secure,omitempty" json:"secure,omitempty"`
}

// CSRFConfig represents CSRF configuration
type CSRFConfig struct {
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Header  string `yaml:"header,omitempty" json:"header,omitempty"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Origins     []string `yaml:"origins,omitempty" json:"origins,omitempty"`
	Methods     []string `yaml:"methods,omitempty" json:"methods,omitempty"`
	Headers     []string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Credentials bool     `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	MaxAge      int      `yaml:"max_age,omitempty" json:"max_age,omitempty"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Headers           *SecurityHeadersConfig   `yaml:"headers,omitempty" json:"headers,omitempty"`
	RateLimiting      *RateLimitingConfig      `yaml:"rate_limiting,omitempty" json:"rate_limiting,omitempty"`
	RequestValidation *RequestValidationConfig `yaml:"request_validation,omitempty" json:"request_validation,omitempty"`
}

// SecurityHeadersConfig represents security headers configuration
type SecurityHeadersConfig struct {
	Enabled            bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	ContentTypeOptions string `yaml:"content_type_options,omitempty" json:"content_type_options,omitempty"`
	FrameOptions       string `yaml:"frame_options,omitempty" json:"frame_options,omitempty"`
	XSSProtection      string `yaml:"xss_protection,omitempty" json:"xss_protection,omitempty"`
	ReferrerPolicy     string `yaml:"referrer_policy,omitempty" json:"referrer_policy,omitempty"`
	PermissionsPolicy  string `yaml:"permissions_policy,omitempty" json:"permissions_policy,omitempty"`
}

// RateLimitingConfig represents rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	RequestsPerMinute int    `yaml:"requests_per_minute,omitempty" json:"requests_per_minute,omitempty"`
	Burst             int    `yaml:"burst,omitempty" json:"burst,omitempty"`
	Store             string `yaml:"store,omitempty" json:"store,omitempty"`
}

// RequestValidationConfig represents request validation configuration
type RequestValidationConfig struct {
	MaxBodySize         string   `yaml:"max_body_size,omitempty" json:"max_body_size,omitempty"`
	AllowedContentTypes []string `yaml:"allowed_content_types,omitempty" json:"allowed_content_types,omitempty"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level          string                `yaml:"level,omitempty" json:"level,omitempty"`
	Format         string                `yaml:"format,omitempty" json:"format,omitempty"`
	Output         string                `yaml:"output,omitempty" json:"output,omitempty"`
	File           *LogFileConfig        `yaml:"file,omitempty" json:"file,omitempty"`
	RequestLogging *RequestLoggingConfig `yaml:"request_logging,omitempty" json:"request_logging,omitempty"`
}

// LogFileConfig represents log file configuration
type LogFileConfig struct {
	Path       string `yaml:"path,omitempty" json:"path,omitempty"`
	MaxSize    string `yaml:"max_size,omitempty" json:"max_size,omitempty"`
	MaxBackups int    `yaml:"max_backups,omitempty" json:"max_backups,omitempty"`
	MaxAge     int    `yaml:"max_age,omitempty" json:"max_age,omitempty"`
}

// RequestLoggingConfig represents request logging configuration
type RequestLoggingConfig struct {
	Enabled    bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	LogBody    bool `yaml:"log_body,omitempty" json:"log_body,omitempty"`
	LogHeaders bool `yaml:"log_headers,omitempty" json:"log_headers,omitempty"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	HealthCheck *HealthCheckConfig `yaml:"health_check,omitempty" json:"health_check,omitempty"`
	Metrics     *MetricsConfig     `yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Tracing     *TracingConfig     `yaml:"tracing,omitempty" json:"tracing,omitempty"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Enabled bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Path    string   `yaml:"path,omitempty" json:"path,omitempty"`
	Checks  []string `yaml:"checks,omitempty" json:"checks,omitempty"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Path       string `yaml:"path,omitempty" json:"path,omitempty"`
	Prometheus bool   `yaml:"prometheus,omitempty" json:"prometheus,omitempty"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled bool          `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Jaeger  *JaegerConfig `yaml:"jaeger,omitempty" json:"jaeger,omitempty"`
}

// JaegerConfig represents Jaeger configuration
type JaegerConfig struct {
	Endpoint    string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	ServiceName string `yaml:"service_name,omitempty" json:"service_name,omitempty"`
}

// ServiceV1 represents a backend service configuration (v1)
type ServiceV1 struct {
	BaseURL        string                `yaml:"base_url" json:"base_url"`
	Timeout        string                `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries        int                   `yaml:"retries,omitempty" json:"retries,omitempty"`
	CircuitBreaker *CircuitBreakerConfig `yaml:"circuit_breaker,omitempty" json:"circuit_breaker,omitempty"`
	Endpoints      []EndpointV1          `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	FailureThreshold int    `yaml:"failure_threshold,omitempty" json:"failure_threshold,omitempty"`
	RecoveryTimeout  string `yaml:"recovery_timeout,omitempty" json:"recovery_timeout,omitempty"`
}

// EndpointV1 represents a single API endpoint (v1)
type EndpointV1 struct {
	Name         string           `yaml:"name" json:"name"`
	Path         string           `yaml:"path" json:"path"`
	Method       string           `yaml:"method" json:"method"`
	ExposeAs     string           `yaml:"expose_as" json:"expose_as"`
	AuthRequired bool             `yaml:"auth_required,omitempty" json:"auth_required,omitempty"`
	Cache        *CacheConfig     `yaml:"cache,omitempty" json:"cache,omitempty"`
	Transform    *TransformConfig `yaml:"transform,omitempty" json:"transform,omitempty"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Enabled     bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	TTL         string `yaml:"ttl,omitempty" json:"ttl,omitempty"`
	KeyTemplate string `yaml:"key_template,omitempty" json:"key_template,omitempty"`
}

// TransformConfig represents transformation configuration
type TransformConfig struct {
	Request  map[string]interface{} `yaml:"request,omitempty" json:"request,omitempty"`
	Response map[string]interface{} `yaml:"response,omitempty" json:"response,omitempty"`
}

// AggregatorConfig represents aggregator configuration
type AggregatorConfig struct {
	Name         string              `yaml:"name" json:"name"`
	Endpoint     string              `yaml:"endpoint" json:"endpoint"`
	Method       string              `yaml:"method,omitempty" json:"method,omitempty"`
	AuthRequired bool                `yaml:"auth_required,omitempty" json:"auth_required,omitempty"`
	Services     []AggregatorService `yaml:"services" json:"services"`
	Response     *AggregatorResponse `yaml:"response,omitempty" json:"response,omitempty"`
}

// AggregatorService represents a service call within an aggregator
type AggregatorService struct {
	Service  string `yaml:"service" json:"service"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Required bool   `yaml:"required,omitempty" json:"required,omitempty"`
	Timeout  string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// AggregatorResponse represents aggregator response configuration
type AggregatorResponse struct {
	MergeStrategy string `yaml:"merge_strategy,omitempty" json:"merge_strategy,omitempty"`
	Template      string `yaml:"template,omitempty" json:"template,omitempty"`
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	Name    string                 `yaml:"name" json:"name"`
	Type    string                 `yaml:"type" json:"type"`
	Enabled bool                   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Order   int                    `yaml:"order,omitempty" json:"order,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// EnvironmentConfig represents environment configuration
type EnvironmentConfig struct {
	Variables map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	Files     []string          `yaml:"files,omitempty" json:"files,omitempty"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	GoVersion  string   `yaml:"go_version,omitempty" json:"go_version,omitempty"`
	BuildTags  []string `yaml:"build_tags,omitempty" json:"build_tags,omitempty"`
	LDFlags    []string `yaml:"ldflags,omitempty" json:"ldflags,omitempty"`
	CGOEnabled bool     `yaml:"cgo_enabled,omitempty" json:"cgo_enabled,omitempty"`
}

// DeploymentConfig represents deployment configuration
type DeploymentConfig struct {
	Docker     *DockerConfig     `yaml:"docker,omitempty" json:"docker,omitempty"`
	Kubernetes *KubernetesConfig `yaml:"kubernetes,omitempty" json:"kubernetes,omitempty"`
}

// DockerConfig represents Docker configuration
type DockerConfig struct {
	Enabled    bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	BaseImage  string `yaml:"base_image,omitempty" json:"base_image,omitempty"`
	FinalImage string `yaml:"final_image,omitempty" json:"final_image,omitempty"`
	MultiStage bool   `yaml:"multi_stage,omitempty" json:"multi_stage,omitempty"`
}

// KubernetesConfig represents Kubernetes configuration
type KubernetesConfig struct {
	Enabled   bool                 `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Namespace string               `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Replicas  int                  `yaml:"replicas,omitempty" json:"replicas,omitempty"`
	Resources *KubernetesResources `yaml:"resources,omitempty" json:"resources,omitempty"`
}

// KubernetesResources represents Kubernetes resource configuration
type KubernetesResources struct {
	Requests *ResourceLimits `yaml:"requests,omitempty" json:"requests,omitempty"`
	Limits   *ResourceLimits `yaml:"limits,omitempty" json:"limits,omitempty"`
}

// ResourceLimits represents resource limits
type ResourceLimits struct {
	CPU    string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty" json:"memory,omitempty"`
}

// GetDefaultBFFGenV1Config returns a default v1 configuration
func GetDefaultBFFGenV1Config() *BFFGenV1Config {
	return &BFFGenV1Config{
		Version: "1.0",
		Project: ProjectConfig{
			Name:      "my-bff",
			Version:   "1.0.0",
			Language:  "go",
			Framework: "chi",
		},
		Server: &ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
			Timeout: &TimeoutConfig{
				Read:  "30s",
				Write: "30s",
				Idle:  "120s",
			},
			GracefulShutdown: &GracefulShutdownConfig{
				Enabled: true,
				Timeout: "30s",
			},
		},
		Auth: &AuthConfig{
			Mode: "jwt",
			JWT: &JWTConfig{
				Expiration:        "15m",
				RefreshExpiration: "24h",
				Encryption: &EncryptionConfig{
					Enabled:   true,
					Algorithm: "AES-GCM",
				},
			},
			CSRF: &CSRFConfig{
				Enabled: true,
				Header:  "X-CSRF-Token",
			},
		},
		CORS: &CORSConfig{
			Enabled:     true,
			Origins:     []string{"http://localhost:3000", "http://localhost:3001"},
			Methods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			Headers:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			Credentials: true,
			MaxAge:      86400,
		},
		Security: &SecurityConfig{
			Headers: &SecurityHeadersConfig{
				Enabled:            true,
				ContentTypeOptions: "nosniff",
				FrameOptions:       "DENY",
				XSSProtection:      "1; mode=block",
				ReferrerPolicy:     "strict-origin-when-cross-origin",
				PermissionsPolicy:  "geolocation=(), microphone=(), camera=()",
			},
			RateLimiting: &RateLimitingConfig{
				Enabled:           true,
				RequestsPerMinute: 100,
				Burst:             10,
				Store:             "memory",
			},
			RequestValidation: &RequestValidationConfig{
				MaxBodySize:         "10mb",
				AllowedContentTypes: []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"},
			},
		},
		Logging: &LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
			RequestLogging: &RequestLoggingConfig{
				Enabled:    true,
				LogBody:    false,
				LogHeaders: true,
			},
		},
		Monitoring: &MonitoringConfig{
			HealthCheck: &HealthCheckConfig{
				Enabled: true,
				Path:    "/health",
			},
			Metrics: &MetricsConfig{
				Enabled:    true,
				Path:       "/metrics",
				Prometheus: true,
			},
		},
		Services:    make(map[string]ServiceV1),
		Aggregators: []AggregatorConfig{},
		Middleware:  []MiddlewareConfig{},
		Build: &BuildConfig{
			GoVersion:  "1.21",
			CGOEnabled: false,
		},
		Deployment: &DeploymentConfig{
			Docker: &DockerConfig{
				Enabled:    true,
				BaseImage:  "golang:1.21-alpine",
				FinalImage: "alpine:latest",
				MultiStage: true,
			},
		},
	}
}
