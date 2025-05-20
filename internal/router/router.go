package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/handlers"
	"github.com/ravibandhu/oolio-food-ordering/internal/services"
)

// Router wraps the underlying router implementation and associated resources
type Router struct {
	engine *gin.Engine
	store  *data.Store
}

// NewRouter creates a new Router instance
func NewRouter(ctx context.Context, store *data.Store) *Router {
	r := &Router{
		engine: gin.Default(),
		store:  store,
	}

	// Set up routes
	r.setupRoutes(ctx)

	return r
}

// setupRoutes configures all the routes for the application
func (r *Router) setupRoutes(ctx context.Context) {
	// Create services
	orderService := services.NewOrderService(r.store)

	// Create handlers
	productHandler := handlers.NewProductHandler(r.store)
	orderHandler := handlers.NewOrderHandler(orderService)
	profileHandler := handlers.NewProfileHandler()

	// Product routes
	products := r.engine.Group("/products")
	{
		products.GET("", gin.WrapF(productHandler.ListProducts))
		products.GET("/:id", gin.WrapF(productHandler.GetProduct))
		// TODO: Add other product routes
	}

	// Order routes
	orders := r.engine.Group("/orders")
	{
		orders.POST("", gin.WrapF(orderHandler.PlaceOrder))
	}

	// Profile routes (protected, should be disabled in production)
	if gin.Mode() != gin.ReleaseMode {
		profile := r.engine.Group("/debug/profile")
		{
			profile.GET("/cpu", profileHandler.StartCPUProfile)
			profile.GET("/memory", profileHandler.GetMemoryProfile)
			profile.GET("/goroutine", profileHandler.GetGoroutineProfile)
		}
	}

	// Add middleware to check context cancellation
	r.engine.Use(func(c *gin.Context) {
		select {
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server is shutting down",
			})
			return
		default:
			c.Next()
		}
	})
}

// Engine returns the underlying gin.Engine instance
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// Shutdown performs cleanup when the router is being shut down
func (r *Router) Shutdown(ctx context.Context) error {
	// Close the store
	if err := r.store.Close(); err != nil {
		return err
	}
	return nil
}
