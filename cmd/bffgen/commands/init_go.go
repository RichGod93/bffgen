// File: init_go.go
// Purpose: Go-specific project initialization
// Contains all logic for scaffolding Go BFF projects

package commands

import (
	"os"
	"path/filepath"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/utils"
)

// createGoModFile creates go.mod file for Go projects
func createGoModFile(projectName, framework string) error {
	content := generateGoModContent(projectName, framework)
	return os.WriteFile(filepath.Join(projectName, "go.mod"), []byte(content), utils.ProjectFilePerm)
}

// generateGoMod generates go.mod file content (legacy wrapper for tests)
func generateGoMod(projectName, framework string) string {
	return generateGoModContent(projectName, framework)
}

// generateGoModContent creates content for go.mod file
func generateGoModContent(projectName, framework string) string {
	baseContent := `module ` + projectName + `

go 1.21

require (
	gopkg.in/yaml.v3 v3.0.1
)`

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

// createGoMainFile creates Go main.go file
func createGoMainFile(projectName, framework, corsConfig string) error {
	// For now, just create a placeholder - this would need the full Go template logic
	content := `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "BFF server is running!")
	})

	fmt.Println("🚀 BFF server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`
	return os.WriteFile(filepath.Join(projectName, "main.go"), []byte(content), utils.ProjectFilePerm)
}

// generateCORSConfig generates CORS config for Go frameworks (legacy wrapper for tests)
func generateCORSConfig(origins []string, framework string) string {
	return generateCORSConfigWithLang(origins, framework, scaffolding.LanguageGo)
}

// generateCORSConfigWithLang generates CORS config with language support
func generateCORSConfigWithLang(origins []string, framework string, langType scaffolding.LanguageType) string {
	// For Node.js, return empty for now (it will be in the template)
	if langType != scaffolding.LanguageGo {
		return ""
	}

	originsStr := ""
	for i, origin := range origins {
		if i > 0 {
			originsStr += ", "
		}
		originsStr += `"` + origin + `"`
	}

	switch framework {
	case "chi":
		return `r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{` + originsStr + `},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`
	case "echo":
		return `e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{` + originsStr + `},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))`
	case "fiber":
		originsStr = ""
		for i, origin := range origins {
			if i > 0 {
				originsStr += ","
			}
			originsStr += origin
		}
		return `app.Use(cors.New(cors.Config{
		AllowOrigins:     "` + originsStr + `",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-CSRF-Token",
		ExposeHeaders:    "Link",
		AllowCredentials: true,
		MaxAge:           300,
	}))`
	default:
		return ""
	}
}

