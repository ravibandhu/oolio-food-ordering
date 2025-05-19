// Package main provides the entry point for the Oolio Food Ordering API server
package main

import (
	"log"
	"net/http"

	_ "github.com/ravibandhu/oolio-food-ordering/docs"
	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/handlers"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Register routes
	http.HandleFunc("/api/v1/products", handlers.ListProducts)
	http.HandleFunc("/api/v1/products/", handlers.GetProduct)

	// Start the server
	log.Printf("Starting server on %s", cfg.Server.Port)
	if err := http.ListenAndServe(cfg.Server.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
