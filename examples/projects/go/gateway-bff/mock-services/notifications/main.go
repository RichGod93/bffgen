package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

var notifications = []Notification{
	{
		ID:        "notif-1",
		UserID:    "1",
		Title:     "Welcome!",
		Message:   "Welcome to the Gateway API platform",
		Type:      "info",
		Read:      false,
		CreatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "notif-2",
		UserID:    "1",
		Title:     "New Feature",
		Message:   "Check out our new analytics dashboard",
		Type:      "feature",
		Read:      false,
		CreatedAt: time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "notif-3",
		UserID:    "2",
		Title:     "System Update",
		Message:   "System will be under maintenance tonight",
		Type:      "warning",
		Read:      true,
		CreatedAt: time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
	},
	{
		ID:        "notif-4",
		UserID:    "1",
		Title:     "Password Changed",
		Message:   "Your password was successfully updated",
		Type:      "security",
		Read:      false,
		CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
	},
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func getNotificationByID(id string) *Notification {
	for i := range notifications {
		if notifications[i].ID == id {
			return &notifications[i]
		}
	}
	return nil
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
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

	// Filter by user_id if provided
	userID := r.URL.Query().Get("user_id")
	var filtered []Notification

	if userID != "" {
		for _, notif := range notifications {
			if notif.UserID == userID {
				filtered = append(filtered, notif)
			}
		}
	} else {
		filtered = notifications
	}

	unreadCount := 0
	for _, notif := range filtered {
		if !notif.Read {
			unreadCount++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": filtered,
		"total":         len(filtered),
		"unread":        unreadCount,
	})
}

func handleNotificationByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extract notification ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	notifID := pathParts[3]

	notif := getNotificationByID(notifID)
	if notif == nil {
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		json.NewEncoder(w).Encode(notif)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleMarkRead(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extract notification ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	notifID := pathParts[3]

	for i := range notifications {
		if notifications[i].ID == notifID {
			notifications[i].Read = true
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":      "Notification marked as read",
				"notification": notifications[i],
			})
			return
		}
	}

	http.Error(w, "Notification not found", http.StatusNotFound)
}

func handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get user_id from query or body
	userID := r.URL.Query().Get("user_id")

	markedCount := 0
	for i := range notifications {
		if userID == "" || notifications[i].UserID == userID {
			if !notifications[i].Read {
				notifications[i].Read = true
				markedCount++
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("%d notifications marked as read", markedCount),
		"count":   markedCount,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "notifications-service",
		"port":    4002,
	})
}

func main() {
	http.HandleFunc("/api/notifications", handleNotifications)
	http.HandleFunc("/api/notifications/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/read") {
			handleMarkRead(w, r)
		} else {
			handleNotificationByID(w, r)
		}
	})
	http.HandleFunc("/api/notifications/read-all", handleMarkAllRead)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("🔔 Notifications Service running on http://localhost:4002")
	fmt.Println("📋 Endpoints:")
	fmt.Println("   - GET  /api/notifications           - List notifications")
	fmt.Println("   - GET  /api/notifications/{id}      - Get notification by ID")
	fmt.Println("   - POST /api/notifications/{id}/read - Mark as read")
	fmt.Println("   - POST /api/notifications/read-all  - Mark all as read")
	fmt.Println("   - GET  /health                      - Health check")

	if err := http.ListenAndServe(":4002", nil); err != nil {
		log.Fatal(err)
	}
}
