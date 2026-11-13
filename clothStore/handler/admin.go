package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"clothStore/auth"
	"clothStore/mw"
	"clothStore/store"
)

type AdminHandler struct {
	AdminStore   *store.AdminStore
	ProductStore *store.ProductStore
}

func NewAdminHandler(as *store.AdminStore, ps *store.ProductStore) *AdminHandler {
	return &AdminHandler{AdminStore: as, ProductStore: ps}
}

// ---------- Users (Members) ----------

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

// POST /api/admin/users
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var in store.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" || in.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}
	if len(in.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 6 characters"})
		return
	}
	hashed, err := auth.HashPassword(in.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}
	u, err := h.AdminStore.CreateUser(r.Context(), in, hashed)
	if err != nil {
		log.Printf("Admin CreateUser error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user (email may already exist)"})
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// PUT /api/admin/users/{id}
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	var in store.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" || in.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}
	if in.Role == "" {
		in.Role = "customer"
	}
	if err := h.AdminStore.UpdateUser(r.Context(), id, in); err != nil {
		log.Printf("Admin UpdateUser error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update user"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user updated"})
}

// DELETE /api/admin/users/{id}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	if id == mw.GetUserID(r) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot delete your own admin account"})
		return
	}
	if err := h.AdminStore.DeleteUser(r.Context(), id); err != nil {
		log.Printf("Admin DeleteUser error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete user"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// ---------- Orders ----------

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

// ---------- Reviews ----------

// GET /api/reviews (public)
func (h *AdminHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.AdminStore.ListReviews(r.Context())
	if err != nil {
		log.Printf("ListReviews error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load reviews"})
		return
	}
	writeJSON(w, http.StatusOK, reviews)
}

// POST /api/reviews (authenticated customer/admin)
func (h *AdminHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var req struct {
		ProductID int    `json:"productId"`
		Rating    int    `json:"rating"`
		Comment   string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID <= 0 || req.Rating < 1 || req.Rating > 5 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid productId and rating (1-5) are required"})
		return
	}
	rv, err := h.AdminStore.CreateReview(r.Context(), userID, req.ProductID, req.Rating, req.Comment)
	if err != nil {
		log.Printf("CreateReview error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to submit review"})
		return
	}
	writeJSON(w, http.StatusCreated, rv)
}

// PUT /api/admin/reviews/{id}
func (h *AdminHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid review id"})
		return
	}
	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Rating < 1 || req.Rating > 5 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "rating must be between 1 and 5"})
		return
	}
	if err := h.AdminStore.UpdateReview(r.Context(), id, req.Rating, req.Comment); err != nil {
		log.Printf("UpdateReview error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update review"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "review updated"})
}

// DELETE /api/admin/reviews/{id}
func (h *AdminHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid review id"})
		return
	}
	if err := h.AdminStore.DeleteReview(r.Context(), id); err != nil {
		log.Printf("DeleteReview error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete review"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "review deleted"})
}

// ---------- Products (Admin) ----------

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