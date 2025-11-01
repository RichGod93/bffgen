package aggregators

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DashboardResponse represents the aggregated dashboard data
type DashboardResponse struct {
	User          interface{} `json:"user"`
	Metrics       interface{} `json:"metrics"`
	Notifications interface{} `json:"notifications"`
	Timestamp     string      `json:"timestamp"`
}

// ServiceResponse holds the response from a backend service
type ServiceResponse struct {
	Data  interface{}
	Error error
}

// GetDashboard fetches and aggregates data from multiple services
func GetDashboard(userID string, baseURLs map[string]string) (*DashboardResponse, error) {
	var wg sync.WaitGroup
	responses := make(map[string]ServiceResponse)
	mu := sync.Mutex{}

	// Helper function to fetch data from a service
	fetchService := func(name, url string, required bool) {
		defer wg.Done()

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get(url)
		if err != nil {
			mu.Lock()
			responses[name] = ServiceResponse{Error: err}
			mu.Unlock()
			if required {
				fmt.Printf("❌ Failed to fetch %s: %v\n", name, err)
			}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			mu.Lock()
			responses[name] = ServiceResponse{Error: err}
			mu.Unlock()
			return
		}

		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			mu.Lock()
			responses[name] = ServiceResponse{Error: err}
			mu.Unlock()
			return
		}

		mu.Lock()
		responses[name] = ServiceResponse{Data: data}
		mu.Unlock()
	}

	// Fetch user data (required)
	wg.Add(1)
	usersURL := fmt.Sprintf("%s/users/%s", baseURLs["users"], userID)
	go fetchService("users", usersURL, true)

	// Fetch analytics metrics (optional)
	wg.Add(1)
	analyticsURL := fmt.Sprintf("%s/metrics", baseURLs["analytics"])
	go fetchService("analytics", analyticsURL, false)

	// Fetch notifications (optional)
	wg.Add(1)
	notificationsURL := fmt.Sprintf("%s/notifications?user_id=%s", baseURLs["notifications"], userID)
	go fetchService("notifications", notificationsURL, false)

	// Wait for all requests to complete
	wg.Wait()

	// Check if required service (users) succeeded
	if resp, ok := responses["users"]; !ok || resp.Error != nil {
		return nil, fmt.Errorf("failed to fetch user data: %v", resp.Error)
	}

	// Build aggregated response
	dashboard := &DashboardResponse{
		User:      responses["users"].Data,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add optional services if available
	if resp, ok := responses["analytics"]; ok && resp.Error == nil {
		dashboard.Metrics = resp.Data
	}

	if resp, ok := responses["notifications"]; ok && resp.Error == nil {
		dashboard.Notifications = resp.Data
	}

	return dashboard, nil
}

