package router

import (
	"github.com/gin-gonic/gin"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/handlers"
	"github.com/ravibandhu/oolio-food-ordering/internal/services"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter(store *data.Store) *gin.Engine {
	// Create default gin router
	r := gin.Default()

	// Create services
	orderService := services.NewOrderService(store)

	// Create handlers
	productHandler := handlers.NewProductHandler(store)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Product routes
	products := r.Group("/products")
	{
		products.GET("", gin.WrapF(productHandler.ListProducts))
		products.GET("/:id", gin.WrapF(productHandler.GetProduct))
		// TODO: Add other product routes
	}

	// Order routes
	orders := r.Group("/orders")
	{
		orders.POST("", gin.WrapF(orderHandler.PlaceOrder))
	}

	return r
}
