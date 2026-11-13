package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"clothStore/auth"
	"clothStore/store"
)

type AuthHandler struct {
	Store *store.UserStore
}

func NewAuthHandler(s *store.UserStore) *AuthHandler {
	return &AuthHandler{Store: s}
}

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if len(strings.Fields(req.FullName)) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "full_name must include both name and surname"})
		return
	}

	if req.Email == "" || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email and password (min 6 chars) are required"})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to process password"})
		return
	}

	user, err := h.Store.CreateUser(r.Context(), req.FullName, req.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email is already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "user registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.Store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"token":   token,
		"user":    user,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}