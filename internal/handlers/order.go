package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/ravibandhu/oolio-food-ordering/internal/services"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService services.OrderService
}

// NewOrderHandler creates a new OrderHandler instance
func NewOrderHandler(orderService services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// @Operation POST /order
// @Summary Place a new order
// @Description Place a new order with optional coupon code
// @Tags orders
// @Accept json
// @Produce json
// @Param order body models.OrderRequest true "Order to place"
// @Success 201 {object} models.Order
// @Failure 400 {object} models.ErrorResponse
// @Failure 422 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /order [post]
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// Set content type header for all responses
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := models.NewErrorResponse("INVALID_REQUEST", "Failed to parse request body").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Validate request
	if err := models.Validate(&req); err != nil {
		errResp := models.NewErrorResponse("VALIDATION_ERROR", "Invalid request data").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Process order
	order, err := h.orderService.PlaceOrder(&req)
	if err != nil {
		// Check if it's a known error type
		if errResp, ok := err.(*models.ErrorResponse); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(errResp)
			return
		}

		// Unknown error
		errResp := models.NewErrorResponse("ORDER_FAILED", "Failed to place order").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Return successful response
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		errResp := models.NewErrorResponse("INTERNAL_ERROR", "Failed to encode response").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}
}
