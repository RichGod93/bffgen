package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Metric struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

var events = []Event{
	{
		ID:        "evt-1",
		Type:      "page_view",
		UserID:    "1",
		Data:      map[string]interface{}{"page": "/dashboard"},
		Timestamp: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "evt-2",
		Type:      "button_click",
		UserID:    "2",
		Data:      map[string]interface{}{"button": "submit", "form": "profile"},
		Timestamp: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "evt-3",
		Type:      "api_call",
		UserID:    "1",
		Data:      map[string]interface{}{"endpoint": "/api/users", "method": "GET"},
		Timestamp: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
	},
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	metrics := []Metric{
		{Name: "active_users", Value: 1234, Unit: "count", Timestamp: time.Now().Format(time.RFC3339)},
		{Name: "requests_per_minute", Value: 456.78, Unit: "rpm", Timestamp: time.Now().Format(time.RFC3339)},
		{Name: "avg_response_time", Value: 125.5, Unit: "ms", Timestamp: time.Now().Format(time.RFC3339)},
		{Name: "error_rate", Value: 0.02, Unit: "percent", Timestamp: time.Now().Format(time.RFC3339)},
		{Name: "cpu_usage", Value: 45.3, Unit: "percent", Timestamp: time.Now().Format(time.RFC3339)},
		{Name: "memory_usage", Value: 67.8, Unit: "percent", Timestamp: time.Now().Format(time.RFC3339)},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics":   metrics,
		"total":     len(metrics),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// List all events
		json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
			"total":  len(events),
		})

	case "POST":
		// Create a new event
		var newEvent Event
		if err := json.NewDecoder(r.Body).Decode(&newEvent); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		newEvent.ID = fmt.Sprintf("evt-%d", len(events)+1)
		newEvent.Timestamp = time.Now().Format(time.RFC3339)
		events = append(events, newEvent)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Event created successfully",
			"event":   newEvent,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "analytics-service",
		"port":    4001,
	})
}

func main() {
	http.HandleFunc("/api/metrics", handleMetrics)
	http.HandleFunc("/api/events", handleEvents)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("📊 Analytics Service running on http://localhost:4001")
	fmt.Println("📋 Endpoints:")
	fmt.Println("   - GET  /api/metrics  - Get system metrics")
	fmt.Println("   - GET  /api/events   - List all events")
	fmt.Println("   - POST /api/events   - Create event")
	fmt.Println("   - GET  /health       - Health check")

	if err := http.ListenAndServe(":4001", nil); err != nil {
		log.Fatal(err)
	}
}
