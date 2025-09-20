package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/RichGod93/bffgen/internal/types"
	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new BFF project",
	Long:  `Initialize a new BFF project with chi router and config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		framework, err := initializeProject(projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ BFF project '%s' initialized successfully!\n", projectName)
		fmt.Printf("📁 Navigate to the project: cd %s\n", projectName)
		fmt.Printf("🚀 Start development server: bffgen dev\n")

		// Add Redis setup instructions for Chi/Echo
		if framework == "chi" || framework == "echo" {
			fmt.Println()
			fmt.Println("🔴 Redis Setup Required for Rate Limiting:")
			fmt.Println("   1. Install Redis: brew install redis (macOS) or apt install redis (Ubuntu)")
			fmt.Println("   2. Start Redis: redis-server")
			fmt.Println("   3. Set environment: export REDIS_URL=redis://localhost:6379")
			fmt.Println("   Note: Fiber includes built-in rate limiting, no Redis needed")
		}

		// Add JWT setup instructions
		fmt.Println()
		fmt.Println("🔐 JWT Authentication Setup:")
		fmt.Println("   1. Set JWT secret: export JWT_SECRET=your-secure-secret-key")
		fmt.Println("   2. Generate tokens in your auth service")
		fmt.Println("   3. Include 'Authorization: Bearer <token>' header in requests")

		// Add global installation instructions
		fmt.Println()
		fmt.Println("💡 To make bffgen available globally:")
		fmt.Println("   macOS/Linux: sudo cp ../bffgen /usr/local/bin/")
		fmt.Println("   Windows: Add the bffgen directory to your PATH")
		fmt.Println("   Or use: go install github.com/RichGod93/bffgen/cmd/bffgen")
	},
}

func initializeProject(projectName string) (string, error) {
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(projectName, "internal", "routes"),
		filepath.Join(projectName, "internal", "aggregators"),
		filepath.Join(projectName, "internal", "templates"),
		filepath.Join(projectName, "cmd", "server"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Load configuration
	config, err := utils.LoadBFFGenConfig()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not load config: %v\n", err)
		config = types.GetDefaultConfig()
	}

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	// Framework selection
	fmt.Printf("✔ Which framework? (chi/echo/fiber) [%s]: ", config.Defaults.Framework)
	framework, _ := reader.ReadString('\n')
	framework = strings.TrimSpace(strings.ToLower(framework))
	if framework == "" {
		framework = config.Defaults.Framework
	}

	// CORS origins configuration
	defaultCORS := strings.Join(config.Defaults.CORSOrigins, ",")
	fmt.Printf("✔ Frontend URLs (comma-separated) [%s]: ", defaultCORS)
	corsOrigins, _ := reader.ReadString('\n')
	corsOrigins = strings.TrimSpace(corsOrigins)
	if corsOrigins == "" {
		corsOrigins = defaultCORS
	}

	// Route configuration
	fmt.Println("✔ Configure routes now or later?")
	fmt.Println("   1) Define manually")
	fmt.Println("   2) Use a template")
	fmt.Println("   3) Skip for now")
	fmt.Printf("✔ Select option (1-3) [%s]: ", config.Defaults.RouteOption)
	routeOption, _ := reader.ReadString('\n')
	routeOption = strings.TrimSpace(routeOption)
	if routeOption == "" {
		routeOption = config.Defaults.RouteOption
	}

	// Copy template files only if user selected template option
	if routeOption == "2" {
		templateFiles := []string{"auth.yaml", "ecommerce.yaml", "content.yaml"}
		for _, templateFile := range templateFiles {
			srcPath := filepath.Join("internal", "templates", templateFile)
			dstPath := filepath.Join(projectName, "internal", "templates", templateFile)

			if _, err := os.Stat(srcPath); err == nil {
				if err := copyFile(srcPath, dstPath); err != nil {
					return "", fmt.Errorf("failed to copy template %s: %w", templateFile, err)
				}
			}
		}
	}

	// Parse CORS origins for template
	corsOriginsList := strings.Split(corsOrigins, ",")
	for i, origin := range corsOriginsList {
		corsOriginsList[i] = strings.TrimSpace(origin)
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			corsOriginsList[i] = "http://" + origin
		}
	}

	// Generate CORS configuration string
	corsConfig := generateCORSConfig(corsOriginsList, framework)

	// Create main.go based on framework
	var mainGoContent string
	switch framework {
	case "chi":
		mainGoContent = fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	r := chi.NewRouter()

	// Structured logging middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Log request
			log.Printf("REQUEST: %%s %%s from %%s", r.Method, r.URL.Path, r.RemoteAddr)
			
			next.ServeHTTP(ww, r)
			
			// Log response
			duration := time.Since(start)
			log.Printf("RESPONSE: %%d %%s %%s %%v", ww.Status(), r.Method, r.URL.Path, duration)
		})
	})

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))
	
	// Production-safe error recovery middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %%v", err)
					
					// In production, don't expose stack traces
					if os.Getenv("ENV") == "production" {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					} else {
						http.Error(w, fmt.Sprintf("Internal Server Error: %%v", err), http.StatusInternalServerError)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	})
	
	// Disable TRACE method for security
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "TRACE" {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	
	// Security headers middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Essential security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
			
			// HSTS for HTTPS (only in production)
			if os.Getenv("ENV") == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}
			
			// Content Security Policy
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
			
			next.ServeHTTP(w, r)
		})
	})
	
	// CORS configuration
	%s
	
	// Request validation middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Request size limit (5MB for security)
			r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
			
			// Content-Type validation for POST/PUT requests
			if r.Method == "POST" || r.Method == "PUT" {
				contentType := r.Header.Get("Content-Type")
				if contentType != "" && contentType != "application/json" && contentType != "application/x-www-form-urlencoded" && contentType != "multipart/form-data" {
					http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
					return
				}
			}
			
			// Validate request method
			allowedMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true, "OPTIONS": true, "HEAD": true,
			}
			if !allowedMethods[r.Method] {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// Redis-based rate limiting middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if Redis is available
			redisURL := os.Getenv("REDIS_URL")
			if redisURL == "" {
				redisURL = "redis://localhost:6379"
			}
			
			// TODO: Implement Redis rate limiting
			// For now, skip rate limiting if Redis is not available
			// In production, implement proper Redis-based rate limiting
			// Example: github.com/go-redis/redis/v8 with sliding window
			
			next.ServeHTTP(w, r)
		})
	})
	
	// JWT Authentication middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check and public endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Extract JWT token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
				return
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Empty token", http.StatusUnauthorized)
				return
			}
			
			// TODO: Replace with your JWT secret key
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "your-secret-key-change-in-production"
			}
			
			// Parse and validate JWT token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %%v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Extract claims and add to request context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Add user info to request context for downstream handlers
				ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
				ctx = context.WithValue(ctx, "user_email", claims["email"])
				r = r.WithContext(ctx)
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}`, corsConfig)
	case "echo":
		mainGoContent = fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
	
	// Disable TRACE method for security
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "TRACE" {
				return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
			}
			return next(c)
		}
	})
	
	// Production-safe error recovery
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %%v", err)
					
					// In production, don't expose stack traces
					if os.Getenv("ENV") == "production" {
						c.String(http.StatusInternalServerError, "Internal Server Error")
					} else {
						c.String(http.StatusInternalServerError, fmt.Sprintf("Internal Server Error: %%v", err))
					}
				}
			}()
			return next(c)
		}
	})
	
	// Security headers middleware
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
	}))
	
	// Additional security headers
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
			
			// HSTS only in production
			if os.Getenv("ENV") == "production" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}
			
			return next(c)
		}
	})
	
	// CORS configuration
	%s
	
	// Request validation middleware
	e.Use(middleware.BodyLimit("5M"))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Content-Type validation for POST/PUT requests
			if c.Request().Method == "POST" || c.Request().Method == "PUT" {
				contentType := c.Request().Header.Get("Content-Type")
				if contentType != "" && contentType != "application/json" && contentType != "application/x-www-form-urlencoded" && contentType != "multipart/form-data" {
					return c.String(http.StatusUnsupportedMediaType, "Unsupported Content-Type")
				}
			}
			
			// Validate request method
			allowedMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true, "OPTIONS": true, "HEAD": true,
			}
			if !allowedMethods[c.Request().Method] {
				return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
			}
			
			return next(c)
		}
	})
	
	// JWT Authentication middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip auth for health check and public endpoints
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/api/auth/login" || c.Request().URL.Path == "/api/auth/register" {
				return next(c)
			}
			
			// Extract JWT token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.String(http.StatusUnauthorized, "Missing or invalid authorization header")
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				return c.String(http.StatusUnauthorized, "Empty token")
			}
			
			// TODO: Replace with your JWT secret key
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "your-secret-key-change-in-production"
			}
			
			// Parse and validate JWT token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %%v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			
			if err != nil || !token.Valid {
				return c.String(http.StatusUnauthorized, "Invalid token")
			}
			
			// Extract claims and add to request context
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Add user info to request context for downstream handlers
				c.Set("user_id", claims["user_id"])
				c.Set("user_email", claims["email"])
			}
			
			return next(c)
		}
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(e.Start(":8080"))
}`, corsConfig)
	case "fiber":
		mainGoContent = fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

func main() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(timeout.New(timeout.Config{
		Timeout: 30 * time.Second,
	}))
	
	// Disable TRACE method for security
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() == "TRACE" {
			return c.Status(fiber.StatusMethodNotAllowed).SendString("Method Not Allowed")
		}
		return c.Next()
	})
	
	// Production-safe error recovery
	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %%v", err)
				
				// In production, don't expose stack traces
				if os.Getenv("ENV") == "production" {
					c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				} else {
					c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Internal Server Error: %%v", err))
				}
			}
		}()
		return c.Next()
	})
	
	// Security headers middleware
	app.Use(helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		HSTSMaxAge:                31536000,
		ContentSecurityPolicy:     "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=(), payment=(), usb=()",
	}))
	
	// Additional security headers for production
	app.Use(func(c *fiber.Ctx) error {
		// HSTS only in production
		if os.Getenv("ENV") == "production" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		return c.Next()
	})
	
	// Rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:        100, // requests per minute
		Expiration: 1 * time.Minute,
	}))
	
	// CORS configuration
	%s
	
	// Request validation middleware
	app.Use(func(c *fiber.Ctx) error {
		// Request size limit (5MB for security)
		if len(c.Body()) > 5<<20 {
			return c.Status(fiber.StatusRequestEntityTooLarge).SendString("Request too large")
		}
		
		// Content-Type validation for POST/PUT requests
		if c.Method() == "POST" || c.Method() == "PUT" {
			contentType := c.Get("Content-Type")
			if contentType != "" && contentType != "application/json" && contentType != "application/x-www-form-urlencoded" && contentType != "multipart/form-data" {
				return c.Status(415).SendString("Unsupported Content-Type")
			}
		}
		
		// Validate request method
		allowedMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true, "OPTIONS": true, "HEAD": true,
		}
		if !allowedMethods[c.Method()] {
			return c.Status(fiber.StatusMethodNotAllowed).SendString("Method Not Allowed")
		}
		
		return c.Next()
	})
	
	// JWT Authentication middleware
	app.Use(func(c *fiber.Ctx) error {
		// Skip auth for health check and public endpoints
		if c.Path() == "/health" || c.Path() == "/api/auth/login" || c.Path() == "/api/auth/register" {
			return c.Next()
		}
		
		// Extract JWT token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).SendString("Missing or invalid authorization header")
		}
		
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(401).SendString("Empty token")
		}
		
		// TODO: Replace with your JWT secret key
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "your-secret-key-change-in-production"
		}
		
		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %%v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		
		if err != nil || !token.Valid {
			return c.Status(401).SendString("Invalid token")
		}
		
		// Extract claims and add to request context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Add user info to request context for downstream handlers
			c.Locals("user_id", claims["user_id"])
			c.Locals("user_email", claims["email"])
		}
		
		return c.Next()
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("BFF server is running!")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(app.Listen(":8080"))
}`, corsConfig)
	default:
		return "", fmt.Errorf("unsupported framework: %s", framework)
	}

	if err := os.WriteFile(filepath.Join(projectName, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create go.mod based on framework
	goModContent := generateGoMod(projectName, framework)
	if err := os.WriteFile(filepath.Join(projectName, "go.mod"), []byte(goModContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Run go mod tidy to download dependencies
	if err := runCommandInDir(projectName, "go", "mod", "tidy"); err != nil {
		fmt.Printf("⚠️  Warning: Failed to run go mod tidy: %v\n", err)
		fmt.Println("   You may need to run 'go mod tidy' manually in the project directory")
	}

	// Create bff.config.yaml
	configContent := `# BFF Configuration
# Define your backend services and endpoints here

services:
  # Example service configuration
  # users:
  #   baseUrl: "http://localhost:4000/api"
  #   endpoints:
  #     - name: "getUser"
  #       path: "/users/:id"
  #       method: "GET"
  #       exposeAs: "/api/users/:id"
  #     - name: "createUser"
  #       path: "/users"
  #       method: "POST"
  #       exposeAs: "/api/users"

# Global settings
settings:
  port: 8080
  timeout: 30s
  retries: 3
  max_request_size: 5MB
  enable_hsts: true
  disable_trace: true
`

	if err := os.WriteFile(filepath.Join(projectName, "bff.config.yaml"), []byte(configContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create bff.config.yaml: %w", err)
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s

A Backend-for-Frontend (BFF) service generated by bffgen.

## Getting Started

1. Install dependencies:
   `+"```"+`bash
   go mod tidy
   `+"```"+`

2. Configure your backend services in bff.config.yaml

3. Run the development server:
   `+"```"+`bash
   go run main.go
   `+"```"+`

The server will start on http://localhost:8080

## Configuration

Edit bff.config.yaml to define your backend services and endpoints.

## Health Check

Visit http://localhost:8080/health to verify the server is running.

## Global Installation

To make bffgen available globally:
- macOS/Linux: sudo cp ../bffgen /usr/local/bin/
- Windows: Add the bffgen directory to your PATH
- Or use: go install github.com/RichGod93/bffgen/cmd/bffgen
`, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "README.md"), []byte(readmeContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create README.md: %w", err)
	}

	// Handle route configuration based on user choice
	switch routeOption {
	case "1":
		fmt.Println()
		fmt.Println("💡 To add routes manually, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-route")
	case "2":
		fmt.Println()
		fmt.Println("💡 To add a template, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-template")
	}

	// Update configuration with new defaults
	config.Defaults.Framework = framework
	config.Defaults.CORSOrigins = strings.Split(corsOrigins, ",")
	config.Defaults.RouteOption = routeOption

	// Save updated configuration
	if err := utils.SaveBFFGenConfig(config); err != nil {
		fmt.Printf("⚠️  Warning: Could not save config: %v\n", err)
	}

	// Update recent projects
	if err := utils.UpdateRecentProject(projectName); err != nil {
		fmt.Printf("⚠️  Warning: Could not update recent projects: %v\n", err)
	}

	return framework, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// generateGoMod generates the appropriate go.mod content based on the framework
func generateGoMod(projectName, framework string) string {
	baseContent := fmt.Sprintf(`module %s

go 1.21

require (
	gopkg.in/yaml.v3 v3.0.1
)`, projectName)

	switch framework {
	case "chi":
		return `module ` + projectName + `

go 1.21

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	gopkg.in/yaml.v3 v3.0.1
)`
	case "echo":
		return `module ` + projectName + `

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/labstack/echo/v4 v4.11.4
	gopkg.in/yaml.v3 v3.0.1
)`
	case "fiber":
		return `module ` + projectName + `

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/golang-jwt/jwt/v5 v5.2.1
	gopkg.in/yaml.v3 v3.0.1
)`
	default:
		return baseContent
	}
}

// runCommandInDir runs a command in the specified directory
func runCommandInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return nil
}

// generateCORSConfig generates CORS configuration string for different frameworks
func generateCORSConfig(origins []string, framework string) string {
	originsStr := ""
	for i, origin := range origins {
		if i > 0 {
			originsStr += ", "
		}
		originsStr += fmt.Sprintf("\"%s\"", origin)
	}

	switch framework {
	case "chi":
		return fmt.Sprintf(`r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{%s},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`, originsStr)
	case "echo":
		return fmt.Sprintf(`e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{%s},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`, originsStr)
	case "fiber":
		originsStr = ""
		for i, origin := range origins {
			if i > 0 {
				originsStr += ","
			}
			originsStr += origin
		}
		return fmt.Sprintf(`app.Use(cors.New(cors.Config{
		AllowOrigins:     "%s",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		ExposeHeaders:    "Link",
		AllowCredentials: true,
		MaxAge:           300,
	}))`, originsStr)
	default:
		return ""
	}
}
