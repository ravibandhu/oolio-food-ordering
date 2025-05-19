package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/ravibandhu/oolio-food-ordering/internal/services"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new OrderHandler instance
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
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
// @Param order body models.PlaceOrderRequest true "Order to place"
// @Success 201 {object} models.PlaceOrderResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /order [post]
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req models.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := models.NewErrorResponse("INVALID_REQUEST", "Failed to parse request body").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Validate request
	if err := models.Validate(&req); err != nil {
		errResp := models.NewErrorResponse("INVALID_REQUEST", "Invalid request data").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Process order
	response, err := h.orderService.PlaceOrder(&req)
	if err != nil {
		errResp := models.NewErrorResponse("ORDER_FAILED", "Failed to place order").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errResp := models.NewErrorResponse("INTERNAL_ERROR", "Failed to encode response").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}
}
