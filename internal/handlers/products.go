package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// swagger:parameters getProduct
type productIDParam struct {
	// ID of the product to retrieve
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:parameters createProduct
type createProductParam struct {
	// Product to create
	// in: body
	// required: true
	Body models.Product
}

// swagger:response productsResponse
type productsResponse struct {
	// List of products
	// in: body
	Body []models.Product
}

// swagger:response productResponse
type productResponse struct {
	// Single product
	// in: body
	Body models.Product
}

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	store *data.Store
}

// NewProductHandler creates a new ProductHandler instance
func NewProductHandler(store *data.Store) *ProductHandler {
	return &ProductHandler{
		store: store,
	}
}

// @Operation GET /products
// @Summary List all available products
// @Description Get a list of all available products in the system
// @Tags products
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} models.ErrorResponse
// @Router /products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Get all products from the store
	products := h.store.GetAllProducts()

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(products); err != nil {
		errResp := models.NewErrorResponse("INTERNAL_ERROR", "Failed to encode response").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}
}

// @Operation GET /products/{id}
// @Summary Get a specific product
// @Description Get detailed information about a specific product by its ID
// @Tags products
// @Param id path string true "Product ID"
// @Produce json
// @Success 200 {object} models.Product
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		errResp := models.NewErrorResponse("INVALID_REQUEST", "Invalid product ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}
	productID := parts[len(parts)-1]

	// Get product from store
	product, err := h.store.GetProduct(productID)
	if err != nil {
		errResp := models.NewErrorResponse("NOT_FOUND", "Product not found").
			AddDetail("productId", productID).
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(product); err != nil {
		errResp := models.NewErrorResponse("INTERNAL_ERROR", "Failed to encode response").
			AddDetail("error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}
}

// @Operation POST /products
// @Summary Create a new product
// @Description Create a new product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product object to create"
// @Success 201 {object} models.Product
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /products [post]
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}
