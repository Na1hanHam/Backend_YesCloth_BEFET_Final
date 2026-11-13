package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"clothStore/store"
)

type AdminHandler struct {
	AdminStore   *store.AdminStore
	ProductStore *store.ProductStore
}

func NewAdminHandler(as *store.AdminStore, ps *store.ProductStore) *AdminHandler {
	return &AdminHandler{AdminStore: as, ProductStore: ps}
}

// GET /api/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.AdminStore.ListUsers(r.Context())
	if err != nil {
		log.Printf("Admin ListUsers error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load users"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// GET /api/admin/orders
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.AdminStore.ListOrders(r.Context())
	if err != nil {
		log.Printf("Admin ListOrders error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load orders"})
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

// GET /api/reviews (public so both ProductDetail and Admin can use it)
func (h *AdminHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.AdminStore.ListReviews(r.Context())
	if err != nil {
		log.Printf("ListReviews error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load reviews"})
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}

// POST /api/admin/products
func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var in store.UpsertProductInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product data"})
		return
	}
	id, err := h.ProductStore.Create(r.Context(), in)
	if err != nil {
		log.Printf("CreateProduct error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create product"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "message": "product created"})
}

// PUT /api/admin/products/{id}
func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}
	var in store.UpsertProductInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product data"})
		return
	}
	if err := h.ProductStore.Update(r.Context(), id, in); err != nil {
		log.Printf("UpdateProduct error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update product"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product updated"})
}

// DELETE /api/admin/products/{id}
func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}
	if err := h.ProductStore.Delete(r.Context(), id); err != nil {
		log.Printf("DeleteProduct error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete product"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product deleted"})
}