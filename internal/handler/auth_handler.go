// internal/handler/auth_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bughay/go-backend-layout/internal/lib"
	"github.com/Bughay/go-backend-layout/internal/model"
	"github.com/Bughay/go-backend-layout/internal/service"
)

// AuthHandler handles auth-related HTTP requests.
type AuthHandler struct {
	authSvc service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.authSvc.Register(r.Context(), &req)
	if err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	lib.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, refreshToken, err := h.authSvc.Login(r.Context(), &req)

	if err != nil {
		lib.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	lib.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Try to invalidate server-side (optional with JWT refresh tokens)
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = h.authSvc.Logout(r.Context(), cookie.Value)
	}

	// Clear the cookie regardless of server-side result
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // deletes immediately
		HttpOnly: true,
		Secure:   false, // true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	lib.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// Refresh handles token rotation
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		lib.WriteError(w, http.StatusUnauthorized, "refresh token required")
		return
	}

	resp, newRefreshToken, err := h.authSvc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		lib.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Rotate: set new refresh token cookie (old one is now invalid)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})

	lib.WriteJSON(w, http.StatusOK, resp)
}

// isValidationErr checks if the error originated from a validation rule.
func isValidationErr(err error) bool {
	return strings.HasPrefix(err.Error(), "validation:")
}
