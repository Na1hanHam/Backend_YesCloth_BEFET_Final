package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"clothStore/store"
)

type CategoryHandler struct {
	Store *store.CategoryStore
}

func NewCategoryHandler(s *store.CategoryStore) *CategoryHandler {
	return &CategoryHandler{Store: s}
}

// GET /api/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Store.List(r.Context())
	if err != nil {
		log.Printf("List categories error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load categories"})
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

// POST /api/admin/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryName string `json:"category_name"`
		ImageURL     string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.CategoryName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category_name is required"})
		return
	}
	cat, err := h.Store.Create(r.Context(), req.CategoryName, req.ImageURL)
	if err != nil {
		log.Printf("CreateCategory error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

// PUT /api/admin/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category id"})
		return
	}
	var req struct {
		CategoryName string `json:"category_name"`
		ImageURL     string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.CategoryName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category_name is required"})
		return
	}
	if err := h.Store.Update(r.Context(), id, req.CategoryName, req.ImageURL); err != nil {
		log.Printf("UpdateCategory error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category updated"})
}

// DELETE /api/admin/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category id"})
		return
	}
	if err := h.Store.Delete(r.Context(), id); err != nil {
		log.Printf("DeleteCategory error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "category deleted"})
}