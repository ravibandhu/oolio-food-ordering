// Package main provides the entry point for the Oolio Food Ordering API server
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/router"
)

// @title Oolio Food Ordering API
// @version 1.0.0
// @description This is the API server for the Oolio Food Ordering system. It provides endpoints for managing products, orders, and coupons.
// @license.name MIT
// @license.url http://opensource.org/licenses/MIT
// @contact.name API Support
// @contact.email support@oolio.com
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Enable maximum CPU usage
	runtime.GOMAXPROCS(runtime.NumCPU() - 2)

	// Load configuration
	os.Setenv("CONFIG_PATH", "/Users/ravibandhu/personal/go/oolio-food-ordering/config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Print("Configuration loaded successfully")

	// Create data store with context
	store, err := data.NewStore(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	log.Print("Store created successfully")
	defer store.Close()

	// Create router with context
	r := router.NewRouter(ctx, store)
	log.Print("Router created successfully")

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r.Engine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Channel to receive any errors returned from starting the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Channel to receive OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		log.Printf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal: %v", sig)
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Initiate graceful shutdown
	log.Print("Initiating graceful shutdown...")

	// First, shut down the router (and store)
	if err := r.Shutdown(shutdownCtx); err != nil {
		log.Printf("Router shutdown error: %v", err)
	}

	// Then, shut down the HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		// If we get here, we exceeded shutdown timeout
		if err := srv.Close(); err != nil {
			log.Printf("Server force close error: %v", err)
		}
	}

	// Wait for any in-flight requests to complete
	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Print("Shutdown timed out")
	} else {
		log.Print("Shutdown completed successfully")
	}
}
