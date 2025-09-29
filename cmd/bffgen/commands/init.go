package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
		framework, backendServices, err := initializeProject(projectName)
		if err != nil {			os.Exit(1)
		}

		fmt.Printf("✅ BFF project '%s' initialized successfully!\n", projectName)

		// Show backend configuration summary
		showBackendConfigSummary(backendServices)

		// Add setup instructions
		showSetupInstructions(backendServices, projectName)
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
		fmt.Println("🔐 Secure Authentication Setup:")
		fmt.Println("   1. Set encryption key: export ENCRYPTION_KEY=<base64-encoded-32-byte-key>")
		fmt.Println("   2. Set JWT secret: export JWT_SECRET=<base64-encoded-32-byte-key>")
		fmt.Println("   3. Keys will be auto-generated if not set (check console output)")
		fmt.Println("   4. Features: Encrypted JWT tokens, secure sessions, CSRF protection")
		fmt.Println("   5. Auth endpoints: /api/auth/login, /api/auth/refresh, /api/auth/logout")

		// Add global installation instructions
		fmt.Println()
		fmt.Println("💡 To make bffgen available globally:")
		fmt.Println("   macOS/Linux: sudo cp ../bffgen /usr/local/bin/")
		fmt.Println("   Windows: Add the bffgen directory to your PATH")
		fmt.Println("   Or use: go install github.com/RichGod93/bffgen/cmd/bffgen")

		// Add doctor command suggestion
		fmt.Println()
		fmt.Println("🔍 Run 'bffgen doctor' to check your project health")
	},
}
func initializeProject(projectName string) (string, []BackendService, error) {
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create project directory: %w", err)
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
		return "", nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
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


	// Backend architecture selection
	fmt.Println("✔ What's your backend architecture?")
	fmt.Println("   1) Microservices (different ports/URLs)")
	fmt.Println("   2) Monolithic (single port/URL)")
	fmt.Println("   3) Hybrid (some services on same port)")
	fmt.Printf("✔ Select option (1-3) [1]: ")
	backendArch, _ := reader.ReadString('\n')
	backendArch = strings.TrimSpace(backendArch)
	if backendArch == "" {
		backendArch = "1"
	}

	// Configure backend services based on selection
	backendServices, err := configureBackendServices(backendArch, reader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to configure backend services: %w", err)
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
		return "", nil, fmt.Errorf("failed to create main.go: %w", err)
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

	// Copy auth package to the generated project
	if err := copyAuthPackage(projectName); err != nil {
		fmt.Printf("⚠️  Warning: Could not copy auth package: %v\n", err)
	}

	// Create go.mod file
	goModContent := generateGoMod(projectName, framework)
	if err := os.WriteFile(filepath.Join(projectName, "go.mod"), []byte(goModContent), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create main.go based on framework
	var mainGoContent string
	switch framework {
	case "chi":
		mainGoContent = fmt.Sprintf(`package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"`+projectName+`/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	
	// Initialize secure auth
	secureAuth, err := auth.NewSecureAuth()
	if err != nil {
		log.Fatalf("Failed to initialize secure auth: %%v", err)
	}
	
	// CSRF Protection middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for GET, HEAD, OPTIONS
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Skip CSRF for public endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Validate CSRF token
			csrfToken := r.Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				http.Error(w, "CSRF token required", http.StatusForbidden)
				return
			}
			
			// Get session ID from encrypted token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}
			
			encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			if !auth.ValidateCSRFToken(csrfToken, claims.SessionID) {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// Secure JWT Authentication middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check and public endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/register" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Extract encrypted JWT token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
				return
			}
			
			encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
			if encryptedToken == "" {
				http.Error(w, "Empty token", http.StatusUnauthorized)
				return
			}
			
			// Validate encrypted token
			claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Add user info to request context for downstream handlers
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "session_id", claims.SessionID)
			r = r.WithContext(ctx)
			
			// Add CSRF token to response headers
			csrfToken := auth.GenerateCSRFToken(claims.SessionID)
			w.Header().Set("X-CSRF-Token", csrfToken)
			
			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "BFF server is running!")
	})
	
	// Auth endpoints with secure cookies
	r.Post("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Parse login request
		var loginReq struct {
			Email    string `+"`json:\"email\"`"+`
			Password string `+"`json:\"password\"`"+`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		
		// TODO: Validate credentials against your auth service
		// For demo purposes, accept any email/password
		if loginReq.Email == "" || loginReq.Password == "" {
			http.Error(w, "Email and password required", http.StatusBadRequest)
			return
		}
		
		// Create encrypted token
		accessToken, refreshToken, err := secureAuth.CreateEncryptedToken(loginReq.Email, loginReq.Email)
		if err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}
		
		// Set secure cookies
		accessCookie := auth.CreateSecureCookie("access_token", accessToken, 900) // 15 minutes
		refreshCookie := auth.CreateSecureCookie("refresh_token", refreshToken, 86400) // 24 hours
		
		http.SetCookie(w, &http.Cookie{
			Name:     accessCookie["Name"],
			Value:    accessCookie["Value"],
			Path:     accessCookie["Path"],
			MaxAge:   maxAgeToInt(accessCookie["MaxAge"]),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		http.SetCookie(w, &http.Cookie{
			Name:     refreshCookie["Name"],
			Value:    refreshCookie["Value"],
			Path:     refreshCookie["Path"],
			MaxAge:   maxAgeToInt(refreshCookie["MaxAge"]),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		// Return tokens in response
		response := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    "900", // 15 minutes
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	r.Post("/api/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie or header
		var refreshToken string
		
		if cookie, err := r.Cookie("refresh_token"); err == nil {
			refreshToken = cookie.Value
		} else {
			var refreshReq struct {
				RefreshToken string `+"`json:\"refresh_token\"`"+`
			}
			if err := json.NewDecoder(r.Body).Decode(&refreshReq); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			refreshToken = refreshReq.RefreshToken
		}
		
		// Refresh access token
		newAccessToken, err := secureAuth.RefreshToken(refreshToken)
		if err != nil {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		
		// Set new access token cookie
		accessCookie := auth.CreateSecureCookie("access_token", newAccessToken, 900)
		http.SetCookie(w, &http.Cookie{
			Name:     accessCookie["Name"],
			Value:    accessCookie["Value"],
			Path:     accessCookie["Path"],
			MaxAge:   maxAgeToInt(accessCookie["MaxAge"]),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		response := map[string]string{
			"access_token": newAccessToken,
			"token_type":   "Bearer",
			"expires_in":   "900",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	r.Post("/api/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from context
		sessionID, ok := r.Context().Value("session_id").(string)
		if ok {
			secureAuth.RevokeSession(sessionID)
		}
		
		// Clear cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Logged out successfully")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Helper function to convert string to int for cookie MaxAge
func maxAgeToInt(maxAge string) int {
	if val, err := strconv.Atoi(maxAge); err == nil {
		return val
	}
	return 0
}`, corsConfig)
	case "echo":
		mainGoContent = fmt.Sprintf(`package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"`+projectName+`/internal/auth"
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
	
	// Initialize secure auth
	secureAuth, err := auth.NewSecureAuth()
	if err != nil {
		log.Fatalf("Failed to initialize secure auth: %%v", err)
	}
	
	// CSRF Protection middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip CSRF for GET, HEAD, OPTIONS
			if c.Request().Method == "GET" || c.Request().Method == "HEAD" || c.Request().Method == "OPTIONS" {
				return next(c)
			}
			
			// Skip CSRF for public endpoints
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/api/auth/login" || c.Request().URL.Path == "/api/auth/register" {
				return next(c)
			}
			
			// Validate CSRF token
			csrfToken := c.Request().Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				return c.String(http.StatusForbidden, "CSRF token required")
			}
			
			// Get session ID from encrypted token
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.String(http.StatusUnauthorized, "Missing authorization header")
			}
			
			encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
			if err != nil {
				return c.String(http.StatusUnauthorized, "Invalid token")
			}
			
			if !auth.ValidateCSRFToken(csrfToken, claims.SessionID) {
				return c.String(http.StatusForbidden, "Invalid CSRF token")
			}
			
			return next(c)
		}
	})
	
	// Secure JWT Authentication middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip auth for health check and public endpoints
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/api/auth/login" || c.Request().URL.Path == "/api/auth/register" {
				return next(c)
			}
			
			// Extract encrypted JWT token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.String(http.StatusUnauthorized, "Missing or invalid authorization header")
			}
			
			encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
			if encryptedToken == "" {
				return c.String(http.StatusUnauthorized, "Empty token")
			}
			
			// Validate encrypted token
			claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
			if err != nil {
				return c.String(http.StatusUnauthorized, "Invalid token")
			}
			
			// Add user info to request context for downstream handlers
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("session_id", claims.SessionID)
			
			// Add CSRF token to response headers
			csrfToken := auth.GenerateCSRFToken(claims.SessionID)
			c.Response().Header().Set("X-CSRF-Token", csrfToken)
			
			return next(c)
		}
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "BFF server is running!")
	})
	
	// Auth endpoints with secure cookies
	e.POST("/api/auth/login", func(c echo.Context) error {
		// Parse login request
		var loginReq struct {
			Email    string `+"`json:\"email\"`"+`
			Password string `+"`json:\"password\"`"+`
		}
		
		if err := c.Bind(&loginReq); err != nil {
			return c.String(http.StatusBadRequest, "Invalid request")
		}
		
		// TODO: Validate credentials against your auth service
		if loginReq.Email == "" || loginReq.Password == "" {
			return c.String(http.StatusBadRequest, "Email and password required")
		}
		
		// Create encrypted token
		accessToken, refreshToken, err := secureAuth.CreateEncryptedToken(loginReq.Email, loginReq.Email)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to create token")
		}
		
		// Set secure cookies
		c.SetCookie(&http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			MaxAge:   900, // 15 minutes
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			MaxAge:   86400, // 24 hours
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		// Return tokens in response
		response := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    "900",
		}
		
		return c.JSON(http.StatusOK, response)
	})
	
	e.POST("/api/auth/refresh", func(c echo.Context) error {
		// Get refresh token from cookie or body
		var refreshToken string
		
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			refreshToken = cookie.Value
		} else {
			var refreshReq struct {
				RefreshToken string `+"`json:\"refresh_token\"`"+`
			}
			if err := c.Bind(&refreshReq); err != nil {
				return c.String(http.StatusBadRequest, "Invalid request")
			}
			refreshToken = refreshReq.RefreshToken
		}
		
		// Refresh access token
		newAccessToken, err := secureAuth.RefreshToken(refreshToken)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Invalid refresh token")
		}
		
		// Set new access token cookie
		c.SetCookie(&http.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Path:     "/",
			MaxAge:   900,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		response := map[string]string{
			"access_token": newAccessToken,
			"token_type":   "Bearer",
			"expires_in":   "900",
		}
		
		return c.JSON(http.StatusOK, response)
	})
	
	e.POST("/api/auth/logout", func(c echo.Context) error {
		// Get session ID from context
		sessionID := c.Get("session_id")
		if sessionID != nil {
			secureAuth.RevokeSession(sessionID.(string))
		}
		
		// Clear cookies
		c.SetCookie(&http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		
		return c.String(http.StatusOK, "Logged out successfully")
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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"`+projectName+`/internal/auth"
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
	
	// Initialize secure auth
	secureAuth, err := auth.NewSecureAuth()
	if err != nil {
		log.Fatalf("Failed to initialize secure auth: %%v", err)
	}
	
	// CSRF Protection middleware
	app.Use(func(c *fiber.Ctx) error {
		// Skip CSRF for GET, HEAD, OPTIONS
		if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
			return c.Next()
		}
		
		// Skip CSRF for public endpoints
		if c.Path() == "/health" || c.Path() == "/api/auth/login" || c.Path() == "/api/auth/register" {
			return c.Next()
		}
		
		// Validate CSRF token
		csrfToken := c.Get("X-CSRF-Token")
		if csrfToken == "" {
			return c.Status(403).SendString("CSRF token required")
		}
		
		// Get session ID from encrypted token
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).SendString("Missing authorization header")
		}
		
		encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
		if err != nil {
			return c.Status(401).SendString("Invalid token")
		}
		
		if !auth.ValidateCSRFToken(csrfToken, claims.SessionID) {
			return c.Status(403).SendString("Invalid CSRF token")
		}
		
		return c.Next()
	})
	
	// Secure JWT Authentication middleware
	app.Use(func(c *fiber.Ctx) error {
		// Skip auth for health check and public endpoints
		if c.Path() == "/health" || c.Path() == "/api/auth/login" || c.Path() == "/api/auth/register" {
			return c.Next()
		}
		
		// Extract encrypted JWT token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).SendString("Missing or invalid authorization header")
		}
		
		encryptedToken := strings.TrimPrefix(authHeader, "Bearer ")
		if encryptedToken == "" {
			return c.Status(401).SendString("Empty token")
		}
		
		// Validate encrypted token
		claims, err := secureAuth.ValidateEncryptedToken(encryptedToken)
		if err != nil {
			return c.Status(401).SendString("Invalid token")
		}
		
		// Add user info to request context for downstream handlers
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("session_id", claims.SessionID)
		
		// Add CSRF token to response headers
		csrfToken := auth.GenerateCSRFToken(claims.SessionID)
		c.Set("X-CSRF-Token", csrfToken)
		
		return c.Next()
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("BFF server is running!")
	})
	
	// Auth endpoints with secure cookies
	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		// Parse login request
		var loginReq struct {
			Email    string `+"`json:\"email\"`"+`
			Password string `+"`json:\"password\"`"+`
		}
		
		if err := c.BodyParser(&loginReq); err != nil {
			return c.Status(400).SendString("Invalid request")
		}
		
		// TODO: Validate credentials against your auth service
		if loginReq.Email == "" || loginReq.Password == "" {
			return c.Status(400).SendString("Email and password required")
		}
		
		// Create encrypted token
		accessToken, refreshToken, err := secureAuth.CreateEncryptedToken(loginReq.Email, loginReq.Email)
		if err != nil {
			return c.Status(500).SendString("Failed to create token")
		}
		
		// Set secure cookies
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			MaxAge:   900, // 15 minutes
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
		})
		
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			MaxAge:   86400, // 24 hours
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
		})
		
		// Return tokens in response
		response := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
			"expires_in":    "900",
		}
		
		return c.JSON(response)
	})
	
	app.Post("/api/auth/refresh", func(c *fiber.Ctx) error {
		// Get refresh token from cookie or body
		var refreshToken string
		
		if cookie := c.Cookies("refresh_token"); cookie != "" {
			refreshToken = cookie
		} else {
			var refreshReq struct {
				RefreshToken string `+"`json:\"refresh_token\"`"+`
			}
			if err := c.BodyParser(&refreshReq); err != nil {
				return c.Status(400).SendString("Invalid request")
			}
			refreshToken = refreshReq.RefreshToken
		}
		
		// Refresh access token
		newAccessToken, err := secureAuth.RefreshToken(refreshToken)
		if err != nil {
			return c.Status(401).SendString("Invalid refresh token")
		}
		
		// Set new access token cookie
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Path:     "/",
			MaxAge:   900,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
		})
		
		response := map[string]string{
			"access_token": newAccessToken,
			"token_type":   "Bearer",
			"expires_in":   "900",
		}
		
		return c.JSON(response)
	})
	
	app.Post("/api/auth/logout", func(c *fiber.Ctx) error {
		// Get session ID from context
		sessionID := c.Locals("session_id")
		if sessionID != nil {
			secureAuth.RevokeSession(sessionID.(string))
		}
		
		// Clear cookies
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
		})
		
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Strict",
		})
		
		return c.SendString("Logged out successfully")
	})

	// TODO: Add your aggregated routes here
	// Run 'bffgen add-route' or 'bffgen add-template' to add routes
	// Then run 'bffgen generate' to generate the code

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(app.Listen(":8080"))
}`, corsConfig)
	default:
		return "", nil, fmt.Errorf("unsupported framework: %s", framework)
	}

	// Write main.go file
	if err := os.WriteFile(filepath.Join(projectName, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create bff.config.yaml with enhanced backend configuration
	configContent := generateEnhancedBFFConfig(backendServices, projectName)

	if err := os.WriteFile(filepath.Join(projectName, "bff.config.yaml"), []byte(configContent), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to create bff.config.yaml: %w", err)
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
		return "", nil, fmt.Errorf("failed to create README.md: %w", err)
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
	case "3":
		fmt.Println()
		fmt.Println("💡 To add routes later, run:")
		fmt.Printf("   cd %s\n", projectName)
		fmt.Println("   bffgen add-route")
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

	return framework, backendServices, nil
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
	github.com/golang-jwt/jwt/v5 v5.3.0
	gopkg.in/yaml.v3 v3.0.1
)`
	case "echo":
		return `module ` + projectName + `

go 1.21

require (
	github.com/labstack/echo/v4 v4.11.4
	github.com/golang-jwt/jwt/v5 v5.3.0
	gopkg.in/yaml.v3 v3.0.1
)`
	case "fiber":
		return `module ` + projectName + `

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/golang-jwt/jwt/v5 v5.3.0
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

// copyAuthPackage copies the auth package to the generated project
func copyAuthPackage(projectName string) error {
	// Create internal/auth directory in the project
	authDir := filepath.Join(projectName, "internal", "auth")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	// Copy auth files
	authFiles := []string{
		"internal/auth/secure_auth.go",
		"internal/auth/secure_auth_test.go",
	}

	for _, srcFile := range authFiles {
		dstFile := filepath.Join(projectName, srcFile)

		// Check if source file exists
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue // Skip if source file doesn't exist
		}

		if err := copyFile(srcFile, dstFile); err != nil {
			return fmt.Errorf("failed to copy %s: %w", srcFile, err)
		}
	}

	return nil
}

// BackendService represents a backend service configuration
type BackendService struct {
	Name      string
	BaseURL   string
	Port      int
	Path      string
	Endpoints []string
}

func configureBackendServices(arch string, reader *bufio.Reader) ([]BackendService, error) {
	switch arch {
	case "1":
		return configureMicroservices(reader), nil
	case "2":
		return configureMonolithic(reader), nil
	case "3":
		return configureHybrid(reader), nil
	default:
		return configureMicroservices(reader), nil
	}
}

func configureMicroservices(reader *bufio.Reader) []BackendService {
	var services []BackendService

	fmt.Println("\n🔧 Configuring Microservices Backend")
	fmt.Println("Enter your backend services (press Enter with empty name to finish):")

	for {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders'): ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			break
		}

		fmt.Printf("✔ Base URL for %s (e.g., 'http://localhost:4000/api'): ", serviceName)
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:400%d/api", len(services)+1)
			fmt.Printf("   Using default: %s\n", baseURL)
		}

		// Extract port from URL
		port := 4000 + len(services)
		if strings.Contains(baseURL, ":") {
			parts := strings.Split(baseURL, ":")
			if len(parts) >= 3 {
				if p, err := strconv.Atoi(parts[2]); err == nil {
					port = p
				}
			}
		}

		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      "",
			Endpoints: getDefaultEndpoints(serviceName),
		}

		services = append(services, service)
		fmt.Printf("✅ Added %s service on %s\n", serviceName, baseURL)
	}

	return services
}

func configureMonolithic(reader *bufio.Reader) []BackendService {
	fmt.Println("\n🔧 Configuring Monolithic Backend")

	fmt.Printf("✔ Backend base URL (e.g., 'http://localhost:3000/api'): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://localhost:3000/api"
		fmt.Printf("   Using default: %s\n", baseURL)
	}

	// Extract port from URL
	port := 3000
	if strings.Contains(baseURL, ":") {
		parts := strings.Split(baseURL, ":")
		if len(parts) >= 3 {
			if p, err := strconv.Atoi(parts[2]); err == nil {
				port = p
			}
		}
	}

	// Define all services for monolithic backend
	serviceNames := []string{"users", "products", "orders", "cart", "auth"}
	var services []BackendService

	for _, serviceName := range serviceNames {
		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      "",
			Endpoints: getDefaultEndpoints(serviceName),
		}
		services = append(services, service)
	}

	fmt.Printf("✅ Configured monolithic backend on %s\n", baseURL)
	fmt.Printf("   Services: %s\n", strings.Join(serviceNames, ", "))

	return services
}

func configureHybrid(reader *bufio.Reader) []BackendService {
	var services []BackendService

	fmt.Println("\n🔧 Configuring Hybrid Backend")
	fmt.Println("Enter your backend services (press Enter with empty name to finish):")

	for {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders'): ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			break
		}

		fmt.Printf("✔ Base URL for %s (e.g., 'http://localhost:3000/api/users'): ", serviceName)
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:3000/api/%s", serviceName)
			fmt.Printf("   Using default: %s\n", baseURL)
		}

		// Extract port from URL
		port := 3000
		if strings.Contains(baseURL, ":") {
			parts := strings.Split(baseURL, ":")
			if len(parts) >= 3 {
				if p, err := strconv.Atoi(parts[2]); err == nil {
					port = p
				}
			}
		}

		// Extract path from URL
		path := ""
		if strings.Contains(baseURL, "/api/") {
			pathParts := strings.Split(baseURL, "/api/")
			if len(pathParts) > 1 {
				path = "/" + pathParts[1]
			}
		}

		service := BackendService{
			Name:      serviceName,
			BaseURL:   baseURL,
			Port:      port,
			Path:      path,
			Endpoints: getDefaultEndpoints(serviceName),
		}

		services = append(services, service)
		fmt.Printf("✅ Added %s service on %s\n", serviceName, baseURL)
	}

	return services
}

func getDefaultEndpoints(serviceName string) []string {
	switch serviceName {
	case "users":
		return []string{"GET /users", "GET /users/:id", "POST /users", "PUT /users/:id", "DELETE /users/:id"}
	case "products":
		return []string{"GET /products", "GET /products/:id", "POST /products", "PUT /products/:id", "DELETE /products/:id"}
	case "orders":
		return []string{"GET /orders", "GET /orders/:id", "POST /orders", "PUT /orders/:id"}
	case "cart":
		return []string{"GET /cart", "POST /cart/items", "DELETE /cart/items/:id", "POST /cart/checkout"}
	case "auth":
		return []string{"POST /auth/login", "POST /auth/register", "POST /auth/refresh", "POST /auth/logout"}
	default:
		return []string{"GET /" + serviceName, "GET /" + serviceName + "/:id", "POST /" + serviceName, "PUT /" + serviceName + "/:id"}
	}
}

func generateEnhancedBFFConfig(backendServices []BackendService, projectName string) string {
	var configContent strings.Builder

	configContent.WriteString("# BFF Configuration\n")
	configContent.WriteString("# Generated with enhanced backend architecture support\n\n")

	configContent.WriteString("services:\n")

	// Group services by port for better organization
	portGroups := make(map[int][]BackendService)
	for _, service := range backendServices {
		portGroups[service.Port] = append(portGroups[service.Port], service)
	}

	for _, services := range portGroups {
		if len(services) == 1 {
			// Single service on this port
			service := services[0]
			configContent.WriteString(fmt.Sprintf("  %s:\n", service.Name))
			configContent.WriteString(fmt.Sprintf("    baseUrl: %s\n", service.BaseURL))
			configContent.WriteString("    endpoints:\n")

			for _, endpoint := range service.Endpoints {
				parts := strings.Split(endpoint, " ")
				if len(parts) >= 2 {
					method := parts[0]
					path := parts[1]
					name := strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_")
					exposeAs := path

					configContent.WriteString(fmt.Sprintf("      - name: %s\n", name))
					configContent.WriteString(fmt.Sprintf("        path: %s\n", path))
					configContent.WriteString(fmt.Sprintf("        method: %s\n", method))
					configContent.WriteString(fmt.Sprintf("        exposeAs: %s\n", exposeAs))
				}
			}
		} else {
			// Multiple services on same port (hybrid)
			for _, service := range services {
				configContent.WriteString(fmt.Sprintf("  %s:\n", service.Name))
				configContent.WriteString(fmt.Sprintf("    baseUrl: %s\n", service.BaseURL))
				configContent.WriteString("    endpoints:\n")

				for _, endpoint := range service.Endpoints {
					parts := strings.Split(endpoint, " ")
					if len(parts) >= 2 {
						method := parts[0]
						path := parts[1]
						name := strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_")
						exposeAs := path

						configContent.WriteString(fmt.Sprintf("      - name: %s\n", name))
						configContent.WriteString(fmt.Sprintf("        path: %s\n", path))
						configContent.WriteString(fmt.Sprintf("        method: %s\n", method))
						configContent.WriteString(fmt.Sprintf("        exposeAs: %s\n", exposeAs))
					}
				}
			}
		}
	}

	configContent.WriteString("\n# Global settings\n")
	configContent.WriteString("settings:\n")
	configContent.WriteString("  port: 8080\n")
	configContent.WriteString("  timeout: 30s\n")
	configContent.WriteString("  retries: 3\n")
	configContent.WriteString("  max_request_size: 5MB\n")
	configContent.WriteString("  enable_hsts: true\n")
	configContent.WriteString("  disable_trace: true\n")

	return configContent.String()
}

func showBackendConfigSummary(backendServices []BackendService) {
	fmt.Println("\n📋 Backend Configuration Summary:")
	
	// Group services by port to determine architecture
	portGroups := make(map[int][]BackendService)
	for _, service := range backendServices {
		portGroups[service.Port] = append(portGroups[service.Port], service)
	}
	
	if len(portGroups) == 1 {
		// Single port - could be monolithic or hybrid
		var services []BackendService
		for _, s := range portGroups {
			services = s
		}
		
		if len(services) > 3 {
			fmt.Println("   Architecture: Monolithic")
			fmt.Printf("   - Backend: %s\n", services[0].BaseURL)
		} else {
			fmt.Println("   Architecture: Hybrid")
		}
		fmt.Printf("   - Services: %s\n", strings.Join(getServiceNames(services), ", "))
	} else {
		// Multiple ports - microservices
		fmt.Println("   Architecture: Microservices")
		for _, service := range backendServices {
			fmt.Printf("   - %s: %s\n", service.Name, service.BaseURL)
		}
	}
}

func getServiceNames(services []BackendService) []string {
	var names []string
	for _, service := range services {
		names = append(names, service.Name)
	}
	return names
}

func showSetupInstructions(backendServices []BackendService, projectName string) {
	fmt.Println("\n🔧 Setup Instructions:")
	
	// Group services by port
	portGroups := make(map[int][]BackendService)
	for _, service := range backendServices {
		portGroups[service.Port] = append(portGroups[service.Port], service)
	}
	
	if len(portGroups) == 1 {
		// Single port
		var services []BackendService
		for _, s := range portGroups {
			services = s
		}
		
		if len(services) > 3 {
			fmt.Println("   1. Start your monolithic backend:")
			fmt.Printf("      - Backend: %s\n", services[0].BaseURL)
		} else {
			fmt.Println("   1. Start your backend services:")
			fmt.Printf("      - Services: %s\n", strings.Join(getServiceNames(services), ", "))
		}
	} else {
		// Multiple ports - microservices
		fmt.Println("   1. Start your microservices on the configured ports:")
		for _, service := range backendServices {
			fmt.Printf("      - %s: %s\n", service.Name, service.BaseURL)
		}
	}
	
	fmt.Println("   2. Run the BFF server:")
	fmt.Printf("      cd %s\n", projectName)
	fmt.Println("      bffgen dev")
	fmt.Println("   3. Test the endpoints:")
	fmt.Println("      curl http://localhost:8080/health")
}
