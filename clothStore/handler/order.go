package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"clothStore/mw"
	"clothStore/store"
)

type OrderHandler struct {
	Store *store.OrderStore
}

func NewOrderHandler(s *store.OrderStore) *OrderHandler {
	return &OrderHandler{Store: s}
}

// GET /api/cart
func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	items, err := h.Store.GetCart(r.Context(), userID)
	if err != nil {
		log.Printf("GetCart error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load cart"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// POST /api/cart
func (h *OrderHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var req struct {
		VariantID int `json:"variant_id"`
		Qty       int `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.VariantID <= 0 || req.Qty <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid variant_id and qty are required"})
		return
	}
	if err := h.Store.AddToCart(r.Context(), userID, req.VariantID, req.Qty); err != nil {
		log.Printf("AddToCart error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add to cart"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "added to cart"})
}

// PUT /api/cart
func (h *OrderHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var req struct {
		VariantID int `json:"variant_id"`
		Qty       int `json:"qty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.VariantID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid variant_id is required"})
		return
	}
	if err := h.Store.UpdateCartItem(r.Context(), userID, req.VariantID, req.Qty); err != nil {
		log.Printf("UpdateCart error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update cart"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "cart updated"})
}

// POST /api/checkout
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var req struct {
		Address string                  `json:"address"`
		Items   []store.PlaceOrderItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cart items are required"})
		return
	}
	orderID, err := h.Store.PlaceOrder(r.Context(), userID, req.Address, req.Items)
	if err != nil {
		log.Printf("Checkout error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to place order"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"message":  "order placed successfully",
		"order_id": fmt.Sprintf("YC-%05d", orderID),
	})
}

// GET /api/orders/mine
func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	orders, err := h.Store.ListMyOrders(r.Context(), userID)
	if err != nil {
		log.Printf("ListMyOrders error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load orders"})
		return
	}
	writeJSON(w, http.StatusOK, orders)
}