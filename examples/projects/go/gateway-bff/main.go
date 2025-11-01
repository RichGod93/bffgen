package main

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
	fmt.Println("⚠️  This is a stub. Use: go run cmd/server/main.go")
	log.Fatal(http.ListenAndServe(":8080", nil))
}