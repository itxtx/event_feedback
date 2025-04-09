package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yourusername/event-feedback/internal/database"
	"github.com/yourusername/event-feedback/internal/handlers"
	"github.com/yourusername/event-feedback/internal/middleware"
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create router
	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Register handlers
	handlers.RegisterHandlers(mux, db)

	// Apply middleware
	handler := middleware.LogRequest(mux)

	// Start server
	fmt.Printf("Server starting on port %s...\n", port)
	err = http.ListenAndServe(":"+port, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
