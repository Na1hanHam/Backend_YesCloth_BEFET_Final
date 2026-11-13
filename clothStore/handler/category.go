package handler

import (
	"net/http"

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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load categories"})
		return
	}
	writeJSON(w, http.StatusOK, categories)
}