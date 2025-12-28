package types

// GraphQLConfig represents GraphQL-specific configuration
type GraphQLConfig struct {
	Schema          SchemaConfig           `yaml:"schema" json:"schema"`
	DataSources     []DataSource           `yaml:"dataSources" json:"dataSources"`
	SchemaStitching *SchemaStitchingConfig `yaml:"schemaStitching,omitempty" json:"schemaStitching,omitempty"`
	TypeGeneration  *TypeGenerationConfig  `yaml:"typeGeneration,omitempty" json:"typeGeneration,omitempty"`
	Federation      *FederationConfig      `yaml:"federation,omitempty" json:"federation,omitempty"`
}

// SchemaConfig represents the main GraphQL schema configuration
type SchemaConfig struct {
	Path     string   `yaml:"path" json:"path"`                             // Path to schema file(s)
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"` // Glob patterns for multiple schema files
	Inline   string   `yaml:"inline,omitempty" json:"inline,omitempty"`     // Inline schema definition
}

// DataSource represents a backend service data source
type DataSource struct {
	Name       string               `yaml:"name" json:"name"`
	Type       string               `yaml:"type" json:"type"` // rest, graphql, grpc, database
	BaseURL    string               `yaml:"baseUrl" json:"baseUrl"`
	Headers    map[string]string    `yaml:"headers,omitempty" json:"headers,omitempty"`
	Cache      *GraphQLCacheConfig  `yaml:"cache,omitempty" json:"cache,omitempty"`
	RateLimit  *RateLimitConfig     `yaml:"rateLimit,omitempty" json:"rateLimit,omitempty"`
	Endpoints  []DataSourceEndpoint `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
	AuthConfig *DataSourceAuth      `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// DataSourceEndpoint maps REST endpoints to GraphQL types
type DataSourceEndpoint struct {
	Name        string                  `yaml:"name" json:"name"`
	Path        string                  `yaml:"path" json:"path"`
	Method      string                  `yaml:"method" json:"method"`
	GraphQLType string                  `yaml:"graphqlType" json:"graphqlType"` // Maps to GraphQL type
	Transform   *GraphQLTransformConfig `yaml:"transform,omitempty" json:"transform,omitempty"`
}

// DataSourceAuth represents authentication configuration for data sources
type DataSourceAuth struct {
	Type   string            `yaml:"type" json:"type"` // bearer, basic, apiKey, oauth2
	Token  string            `yaml:"token,omitempty" json:"token,omitempty"`
	Header string            `yaml:"header,omitempty" json:"header,omitempty"`
	Config map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
}

// GraphQLTransformConfig represents data transformation rules for GraphQL
type GraphQLTransformConfig struct {
	Mapping     map[string]string `yaml:"mapping,omitempty" json:"mapping,omitempty"`         // Field name mapping
	Exclude     []string          `yaml:"exclude,omitempty" json:"exclude,omitempty"`         // Fields to exclude
	Rename      map[string]string `yaml:"rename,omitempty" json:"rename,omitempty"`           // Rename fields
	CustomLogic string            `yaml:"customLogic,omitempty" json:"customLogic,omitempty"` // Custom transform function
}

// GraphQLCacheConfig represents caching configuration for GraphQL
type GraphQLCacheConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	TTL     int    `yaml:"ttl" json:"ttl"`                       // Time to live in seconds
	Type    string `yaml:"type,omitempty" json:"type,omitempty"` // memory, redis
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`
	MaxRequests int  `yaml:"maxRequests" json:"maxRequests"`
	WindowMs    int  `yaml:"windowMs" json:"windowMs"`
}

// SchemaStitchingConfig represents schema stitching configuration
type SchemaStitchingConfig struct {
	Enabled       bool                  `yaml:"enabled" json:"enabled"`
	RemoteSchemas []RemoteSchemaConfig  `yaml:"remoteSchemas" json:"remoteSchemas"`
	MergeStrategy string                `yaml:"mergeStrategy,omitempty" json:"mergeStrategy,omitempty"` // merge, delegate, extend
	TypeConflicts *TypeConflictStrategy `yaml:"typeConflicts,omitempty" json:"typeConflicts,omitempty"`
}

// RemoteSchemaConfig represents a remote GraphQL schema
type RemoteSchemaConfig struct {
	Name      string            `yaml:"name" json:"name"`
	URL       string            `yaml:"url" json:"url"`
	Headers   map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Namespace string            `yaml:"namespace,omitempty" json:"namespace,omitempty"` // Namespace for types
}

// TypeConflictStrategy defines how to handle type conflicts in schema stitching
type TypeConflictStrategy struct {
	Strategy string            `yaml:"strategy" json:"strategy"` // prefix, namespace, rename, error
	Mappings map[string]string `yaml:"mappings,omitempty" json:"mappings,omitempty"`
}

// TypeGenerationConfig represents type generation configuration
type TypeGenerationConfig struct {
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Language   string                 `yaml:"language" json:"language"` // typescript, go, both
	OutputPath string                 `yaml:"outputPath" json:"outputPath"`
	Plugins    []string               `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Config     map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
	WatchMode  bool                   `yaml:"watchMode,omitempty" json:"watchMode,omitempty"`
}

// FederationConfig represents Apollo Federation configuration
type FederationConfig struct {
	Enabled   bool             `yaml:"enabled" json:"enabled"`
	Version   string           `yaml:"version" json:"version"` // v1, v2
	Gateway   *GatewayConfig   `yaml:"gateway,omitempty" json:"gateway,omitempty"`
	Subgraphs []SubgraphConfig `yaml:"subgraphs,omitempty" json:"subgraphs,omitempty"`
}

// GatewayConfig represents Apollo Gateway configuration
type GatewayConfig struct {
	ServiceList  []ServiceListEntry `yaml:"serviceList" json:"serviceList"`
	PollInterval int                `yaml:"pollInterval,omitempty" json:"pollInterval,omitempty"` // In seconds
}

// ServiceListEntry represents a federated subgraph service
type ServiceListEntry struct {
	Name string `yaml:"name" json:"name"`
	URL  string `yaml:"url" json:"url"`
}

// SubgraphConfig represents a federated subgraph configuration
type SubgraphConfig struct {
	Name       string `yaml:"name" json:"name"`
	URL        string `yaml:"url" json:"url"`
	SchemaPath string `yaml:"schemaPath,omitempty" json:"schemaPath,omitempty"`
}

// ExtendBFFConfigWithGraphQL extends BFFConfig with GraphQL support
type BFFConfigWithGraphQL struct {
	BFFConfig
	GraphQL *GraphQLConfig `yaml:"graphql,omitempty" json:"graphql,omitempty"`
}
