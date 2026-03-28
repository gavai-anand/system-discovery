package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"system-discovery/internal/bootstrap"
	v1 "system-discovery/internal/routes/v1"
	"time"
)

func main() {
	// Create a root context that can be cancelled.
	// This context is used to control the lifecycle of the app and shutdown procedures.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure resources associated with the context are cleaned up on exit

	// Load environment variables from .env file and set them up for runtime
	settingEnvironment()

	// Initialize the application (load dependencies, services, DB connections, etc.)
	app := bootstrap.NewApp(ctx)
	// Initialize the HTTP router with all API routes
	router := v1.Router(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port), // Server will listen on port 8080
		Handler:      router,                   // Use the router as the main request handler
		ReadTimeout:  5 * time.Second,          // Max duration for reading the request
		WriteTimeout: 30 * time.Second,         // Max duration before timing out the response write
	}

	go app.DiscoveryService.StartHeartbeat(ctx)

	go bootstrap.AutoRegister(app)
	// Start a goroutine to handle OS interrupt signals (Ctrl+C, SIGTERM)
	// This will gracefully shutdown the server when the program is terminated.
	go terminate(cancel, server)

	// Start the HTTP server in a goroutine so it doesn't block main
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Fatal if server fails unexpectedly
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for the context to be cancelled (triggered by terminate)
	<-ctx.Done()
}

func settingEnvironment() {
	viper.SetConfigFile(".env") // explicitly load .env

	// Read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err)
	} else {
		fmt.Println("Config file loaded:", viper.ConfigFileUsed())
	}

	// Automatically read environment variables from OS environment
	viper.AutomaticEnv()
}

func terminate(cancel context.CancelFunc, server *http.Server) {
	// Channel to listen for OS signals
	sigCh := make(chan os.Signal, 1)

	// Notify the channel when receiving SIGINT or SIGTERM
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Block here until a signal is received
	<-sigCh
	log.Println("Received shutdown signal...")

	// Cancel the context to notify main and other routines
	cancel() // ✅ unblocks main

	// Attempt to gracefully shutdown the server
	closeServer(server)
}

func closeServer(server *http.Server) {
	// Create a new context with a 10-second timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		// Log error if shutdown fails
		log.Printf("Shutdown error: %v", err)
	}
}
