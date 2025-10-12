package utils

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/RichGod93/bffgen/internal/scaffolding"
	"github.com/RichGod93/bffgen/internal/types"
)

// PromptConfig contains configuration for prompts
type PromptConfig struct {
	Reader   *bufio.Reader
	Defaults types.Defaults
}

// NewPromptConfig creates a new prompt configuration
func NewPromptConfig(reader *bufio.Reader, defaults types.Defaults) *PromptConfig {
	return &PromptConfig{
		Reader:   reader,
		Defaults: defaults,
	}
}

// PromptLanguageSelection prompts user to select programming language/runtime
func (p *PromptConfig) PromptLanguageSelection() (scaffolding.LanguageType, string, error) {
	fmt.Println("✔ Which language/runtime would you like to use?")
	
	supportedLanguages := scaffolding.GetSupportedLanguages()
	for i, lang := range supportedLanguages {
		fmt.Printf("   %d) %s\n", i+1, lang.Name)
	}
	
	defaultOption := "1" // Default to Go
	fmt.Printf("✔ Select option (1-%d) [%s]: ", len(supportedLanguages), defaultOption)
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = defaultOption
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(supportedLanguages) {
		return scaffolding.LanguageGo, "chi", fmt.Errorf("invalid choice: %s", input)
	}
	
	selectedLang := supportedLanguages[choice-1]
	
	// For Node.js, we already have the framework. For Go, prompt for framework
	if selectedLang.Type == scaffolding.LanguageGo {
		framework, err := p.promptGoFramework()
		if err != nil {
			return selectedLang.Type, "chi", err
		}
		return selectedLang.Type, framework, nil
	}
	
	return selectedLang.Type, selectedLang.Framework, nil
}

// promptGoFramework prompts user to select Go framework
func (p *PromptConfig) promptGoFramework() (string, error) {
	fmt.Printf("✔ Which Go framework? (chi/echo/fiber) [%s]: ", p.Defaults.Framework)
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	if input == "" {
		input = p.Defaults.Framework
	}
	
	if input != "chi" && input != "echo" && input != "fiber" {
		return p.Defaults.Framework, fmt.Errorf("unsupported framework: %s", input)
	}
	
	return input, nil
}

// PromptCORSSetting prompts user for CORS origins
func (p *PromptConfig) PromptCORSSetting() ([]string, error) {
	defaultCORS := strings.Join(p.Defaults.CORSOrigins, ",")
	fmt.Printf("✔ Frontend URLs (comma-separated) [%s]: ", defaultCORS)
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = defaultCORS
	}
	
	corsList := strings.Split(input, ",")
	for i, origin := range corsList {
		origin = strings.TrimSpace(origin)
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			corsList[i] = "http://" + origin
		} else {
			corsList[i] = origin
		}
	}
	
	return corsList, nil
}

// PromptBackendArchitecture prompts user for backend architecture
func (p *PromptConfig) PromptBackendArchitecture() (string, error) {
	fmt.Println("✔ What's your backend architecture?")
	fmt.Println("   1) Microservices (different ports/URLs)")
	fmt.Println("   2) Monolithic (single port/URL)")
	fmt.Println("   3) Hybrid (some services on same port)")
	fmt.Printf("✔ Select option (1-3) [1]: ")
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = "1"
	}
	
	if input != "1" && input != "2" && input != "3" {
		return "1", fmt.Errorf("invalid choice: %s", input)
	}
	
	return input, nil
}

// PromptRouteConfiguration prompts user for route configuration
func (p *PromptConfig) PromptRouteConfiguration() (string, error) {
	fmt.Println("✔ Configure routes now or later?")
	fmt.Println("   1) Define manually")
	fmt.Println("   2) Use a template")
	fmt.Println("   3) Skip for now")
	fmt.Printf("✔ Select option (1-3) [%s]: ", p.Defaults.RouteOption)
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = p.Defaults.RouteOption
	}
	
	return input, nil
}

// PromptServicename prompts user for service name
func (p *PromptConfig) PromptServicename(serviceNum int) (string, error) {
	if serviceNum == 0 {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders'): ")
	} else {
		fmt.Printf("✔ Service name (e.g., 'users', 'products', 'orders') [#%d]: ", serviceNum)
	}
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	return input, nil
}

// PromptServiceURL prompts user for service URL
func (p *PromptConfig) PromptServiceURL(serviceName string, defaultURL string) (string, error) {
	fmt.Printf("✔ Base URL for %s (e.g., '%s'): ", serviceName, defaultURL)
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = defaultURL
		fmt.Printf("   Using default: %s\n", defaultURL)
	}
	
	return input, nil
}

// PromptMonolithicURL prompts user for monolithic backend URL
func (p *PromptConfig) PromptMonolithicURL() (string, error) {
	fmt.Printf("✔ Backend base URL (e.g., 'http://localhost:3000/api'): ")
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = "http://localhost:3000/api"
		fmt.Printf("   Using default: %s\n", input)
	}
	
	return input, nil
}

// PromptHybridURL prompts user for hybrid service URL
func (p *PromptConfig) PromptHybridURL(serviceName string) (string, error) {
	fmt.Printf("✔ Base URL for %s (e.g., 'http://localhost:3000/api/%s'): ", serviceName, serviceName)
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" {
		input = fmt.Sprintf("http://localhost:3000/api/%s", serviceName)
		fmt.Printf("   Using default: %s\n", input)
	}
	
	return input, nil
}

// ConfirmAddMore prompts user to confirm adding more services
func (p *PromptConfig) ConfirmAddMore() bool {
	fmt.Printf("✔ Add another service? (y/N): ")
	
	input, _ := p.Reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	return input == "y" || input == "yes"
}
