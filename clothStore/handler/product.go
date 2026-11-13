package handler

import (
	"log"
	"net/http"

	"clothStore/store"
)

type ProductHandler struct {
	Store *store.ProductStore
}

func NewProductHandler(s *store.ProductStore) *ProductHandler {
	return &ProductHandler{Store: s}
}

// GET /api/products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.Store.List(r.Context())
	if err != nil {
		log.Printf("List products error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":  "failed to load products",
			"detail": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, products)
}