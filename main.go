package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/sonyarianto/gobete/internal/modules/scheduler"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/http"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found or error loading .env file")
	}

	// Initialize the database connection
	db.ConnectMySQL()

	// Start the user session cleanup scheduler
	scheduler.StartCleanupUserSessionScheduler()

	// Create and configure the Fiber app
	app := http.NewApp()

	// Determine the port to listen on
	port := os.Getenv("APP_PORT")
	if port == "" {
		// Default port if not specified
		port = "9000"
	}

	// Start the server in a separate goroutine
	log.Printf("Server is starting on :%s", port)
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()
	log.Printf("Server is listening on :%s", port)

	// Wait for shutdown signal and gracefully shut down the server
	http.WaitForShutdown(app)
}
