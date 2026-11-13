package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"clothStore/mw"
	"clothStore/store"
)

type FavoriteHandler struct {
	Store *store.FavoriteStore
}

func NewFavoriteHandler(s *store.FavoriteStore) *FavoriteHandler {
	return &FavoriteHandler{Store: s}
}

// GET /api/favorites -> [1, 3, ...]
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	ids, err := h.Store.ListProductIDs(r.Context(), userID)
	if err != nil {
		log.Printf("List favorites error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load favorites"})
		return
	}
	writeJSON(w, http.StatusOK, ids)
}

// POST /api/favorites  body: {"product_id": 1}
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var req struct {
		ProductID int `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProductID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid product_id is required"})
		return
	}
	if err := h.Store.Add(r.Context(), userID, req.ProductID); err != nil {
		log.Printf("Add favorite error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add favorite"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "added to favorites"})
}

// DELETE /api/favorites/{productId}
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	productID, err := strconv.Atoi(r.PathValue("productId"))
	if err != nil || productID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}
	if err := h.Store.Remove(r.Context(), userID, productID); err != nil {
		log.Printf("Remove favorite error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove favorite"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "removed from favorites"})
}