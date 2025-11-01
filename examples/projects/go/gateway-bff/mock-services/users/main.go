package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

var users = []User{
	{ID: "1", Name: "Alice Johnson", Email: "alice@example.com", Role: "admin"},
	{ID: "2", Name: "Bob Smith", Email: "bob@example.com", Role: "user"},
	{ID: "3", Name: "Charlie Brown", Email: "charlie@example.com", Role: "user"},
	{ID: "4", Name: "Diana Prince", Email: "diana@example.com", Role: "moderator"},
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func getUserByID(id string) *User {
	for _, user := range users {
		if user.ID == id {
			return &user
		}
	}
	return nil
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// List all users
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"total": len(users),
		})

	case "POST":
		// Create a new user
		var newUser User
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		newUser.ID = fmt.Sprintf("%d", len(users)+1)
		users = append(users, newUser)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User created successfully",
			"user":    newUser,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	userID := pathParts[3]

	user := getUserByID(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		// Get single user
		json.NewEncoder(w).Encode(user)

	case "PUT":
		// Update user
		var updatedUser User
		if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		updatedUser.ID = userID
		for i, u := range users {
			if u.ID == userID {
				users[i] = updatedUser
				break
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User updated successfully",
			"user":    updatedUser,
		})

	case "DELETE":
		// Delete user
		for i, u := range users {
			if u.ID == userID {
				users = append(users[:i], users[i+1:]...)
				break
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User deleted successfully",
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
		"service": "users-service",
		"port":    4000,
	})
}

func main() {
	http.HandleFunc("/api/users", handleUsers)
	http.HandleFunc("/api/users/", handleUserByID)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("🚀 Users Service running on http://localhost:4000")
	fmt.Println("📋 Endpoints:")
	fmt.Println("   - GET    /api/users        - List all users")
	fmt.Println("   - POST   /api/users        - Create user")
	fmt.Println("   - GET    /api/users/{id}   - Get user by ID")
	fmt.Println("   - PUT    /api/users/{id}   - Update user")
	fmt.Println("   - DELETE /api/users/{id}   - Delete user")
	fmt.Println("   - GET    /health           - Health check")

	if err := http.ListenAndServe(":4000", nil); err != nil {
		log.Fatal(err)
	}
}
