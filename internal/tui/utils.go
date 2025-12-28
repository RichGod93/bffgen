package tui

import (
	"strings"
)

// parseOrigins parses comma-separated CORS origins
func parseOrigins(input string) []string {
	parts := strings.Split(input, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		// Add http:// if no scheme provided
		if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			origin = "http://" + origin
		}
		origins = append(origins, origin)
	}

	return origins
}
