// Package main provides the entry point for the Oolio Food Ordering API server
package main

import (
	"log"
	"os"

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
	// Load configuration
	os.Setenv("CONFIG_PATH", "/Users/ravibandhu/personal/go/oolio-food-ordering/config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Print("Configuration loaded successfully")

	// Create data store
	store, err := data.NewStore(cfg)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	log.Print("Store created successfully")

	// Setup router
	r := router.SetupRouter(store)
	log.Print("Router setup successfully")

	// Start server
	log.Printf("Starting server on %s", cfg.Server.Port)
	if err := r.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
