package aggregators

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Aggregator defines the interface for data aggregation
type Aggregator interface {
	Aggregate(w http.ResponseWriter, r *http.Request) error
	GetName() string
	GetPath() string
}

// BaseAggregator provides common functionality for aggregators
type BaseAggregator struct {
	Name        string
	Path        string
	Description string
	Services    []string // List of services this aggregator depends on
}

// GetName returns the aggregator name
func (ba *BaseAggregator) GetName() string {
	return ba.Name
}

// GetPath returns the aggregator path
func (ba *BaseAggregator) GetPath() string {
	return ba.Path
}

// UserDashboardAggregator aggregates user data from multiple services
type UserDashboardAggregator struct {
	BaseAggregator
	UserServiceURL        string
	OrdersServiceURL      string
	PreferencesServiceURL string
}

// NewUserDashboardAggregator creates a new user dashboard aggregator
func NewUserDashboardAggregator() *UserDashboardAggregator {
	return &UserDashboardAggregator{
		BaseAggregator: BaseAggregator{
			Name:        "user-dashboard",
			Path:        "/api/user-dashboard/:id",
			Description: "Aggregates user, orders, and preferences data",
			Services:    []string{"users", "orders", "preferences"},
		},
		UserServiceURL:        "http://localhost:4000/api",
		OrdersServiceURL:      "http://localhost:5000/api",
		PreferencesServiceURL: "http://localhost:6000/api",
	}
}

// Aggregate combines data from multiple services
func (uda *UserDashboardAggregator) Aggregate(w http.ResponseWriter, r *http.Request) error {
	// Extract user ID from path
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return fmt.Errorf("user ID is required")
	}

	// Create aggregated response
	dashboard := map[string]interface{}{
		"user":        uda.fetchUserData(userID),
		"orders":      uda.fetchOrdersData(userID),
		"preferences": uda.fetchPreferencesData(userID),
		"timestamp":   time.Now().Unix(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return aggregated data
	return json.NewEncoder(w).Encode(dashboard)
}

// fetchUserData retrieves user information
func (uda *UserDashboardAggregator) fetchUserData(userID string) map[string]interface{} {
	// Use HTTP client to fetch from user service
	client := NewHTTPClient(5 * time.Second)
	userServiceURL := os.Getenv("USER_SERVICE_URL")

	if userServiceURL == "" {
		// Fallback to mock data if service URL not configured
		return map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
			"role":  "user",
		}
	}

	url := fmt.Sprintf("%s/users/%s", userServiceURL, userID)
	resp, err := client.Get(url)
	if err != nil || resp == nil {
		// Return mock data on error
		return map[string]interface{}{
			"id":    userID,
			"error": "Failed to fetch user data",
		}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't fail the request
			_ = closeErr
		}
	}()

	// Parse response
	var userData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return map[string]interface{}{
			"id":    userID,
			"error": "Failed to parse user data",
		}
	}

	return userData
}

// fetchOrdersData retrieves user orders
func (uda *UserDashboardAggregator) fetchOrdersData(userID string) []map[string]interface{} {
	// In a real implementation, this would make HTTP calls to the orders service
	// For now, return mock data
	return []map[string]interface{}{
		{
			"id":     "order-1",
			"amount": 99.99,
			"status": "completed",
			"date":   "2024-01-15",
		},
		{
			"id":     "order-2",
			"amount": 149.99,
			"status": "pending",
			"date":   "2024-01-20",
		},
	}
}

// fetchPreferencesData retrieves user preferences
func (uda *UserDashboardAggregator) fetchPreferencesData(userID string) map[string]interface{} {
	// In a real implementation, this would make HTTP calls to the preferences service
	// For now, return mock data
	return map[string]interface{}{
		"theme":         "dark",
		"language":      "en",
		"notifications": true,
		"timezone":      "UTC",
	}
}

// EcommerceAggregator aggregates e-commerce data
type EcommerceAggregator struct {
	BaseAggregator
	ProductsServiceURL  string
	CartServiceURL      string
	InventoryServiceURL string
}

// NewEcommerceAggregator creates a new e-commerce aggregator
func NewEcommerceAggregator() *EcommerceAggregator {
	return &EcommerceAggregator{
		BaseAggregator: BaseAggregator{
			Name:        "ecommerce-catalog",
			Path:        "/api/catalog/:category",
			Description: "Aggregates products, inventory, and cart data",
			Services:    []string{"products", "inventory", "cart"},
		},
		ProductsServiceURL:  "http://localhost:4000/api",
		CartServiceURL:      "http://localhost:5000/api",
		InventoryServiceURL: "http://localhost:6000/api",
	}
}

// Aggregate combines e-commerce data
func (ea *EcommerceAggregator) Aggregate(w http.ResponseWriter, r *http.Request) error {
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "all"
	}

	// Create aggregated response
	catalog := map[string]interface{}{
		"category":    category,
		"products":    ea.fetchProductsData(category),
		"inventory":   ea.fetchInventoryData(category),
		"cartSummary": ea.fetchCartSummary(),
		"timestamp":   time.Now().Unix(),
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return aggregated data
	return json.NewEncoder(w).Encode(catalog)
}

// fetchProductsData retrieves product information
func (ea *EcommerceAggregator) fetchProductsData(category string) []map[string]interface{} {
	// Mock data - in real implementation, would call products service
	return []map[string]interface{}{
		{
			"id":          "prod-1",
			"name":        "Sample Product",
			"category":    category,
			"price":       29.99,
			"description": "A sample product",
		},
	}
}

// fetchInventoryData retrieves inventory information
func (ea *EcommerceAggregator) fetchInventoryData(category string) map[string]interface{} {
	// Mock data - in real implementation, would call inventory service
	return map[string]interface{}{
		"totalItems": 100,
		"inStock":    85,
		"lowStock":   15,
	}
}

// fetchCartSummary retrieves cart summary
func (ea *EcommerceAggregator) fetchCartSummary() map[string]interface{} {
	// Mock data - in real implementation, would call cart service
	return map[string]interface{}{
		"itemCount": 3,
		"total":     89.97,
		"currency":  "USD",
	}
}

// Registry manages available aggregators
type Registry struct {
	aggregators map[string]Aggregator
}

// NewRegistry creates a new aggregator registry
func NewRegistry() *Registry {
	return &Registry{
		aggregators: make(map[string]Aggregator),
	}
}

// Register adds an aggregator to the registry
func (r *Registry) Register(aggregator Aggregator) {
	r.aggregators[aggregator.GetName()] = aggregator
}

// Get retrieves an aggregator by name
func (r *Registry) Get(name string) (Aggregator, bool) {
	aggregator, exists := r.aggregators[name]
	return aggregator, exists
}

// List returns all registered aggregators
func (r *Registry) List() []Aggregator {
	var aggregators []Aggregator
	for _, agg := range r.aggregators {
		aggregators = append(aggregators, agg)
	}
	return aggregators
}

// DefaultRegistry returns a registry with default aggregators
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register default aggregators
	registry.Register(NewUserDashboardAggregator())
	registry.Register(NewEcommerceAggregator())

	return registry
}
