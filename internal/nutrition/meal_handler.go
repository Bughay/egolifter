package nutrition

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bughay/egolifter/internal/auth"
	"github.com/Bughay/egolifter/internal/lib"
)

// MealHandler handles meal-related HTTP requests.
type MealHandler struct {
	mealSvc MealService
}

// NewMealHandler creates a new MealHandler.
func NewMealHandler(mealSvc MealService) *MealHandler {
	return &MealHandler{mealSvc: mealSvc}
}

// RegisterRoutes attaches the meal endpoints to the given mux,
// wrapping every route with the provided middleware (JWT auth).
func (h *MealHandler) RegisterRoutes(mux *http.ServeMux, mw func(http.Handler) http.Handler) {
	mux.Handle("POST /meal/create", mw(http.HandlerFunc(h.CreateMeal)))
	mux.Handle("GET /meal/view", mw(http.HandlerFunc(h.ViewMeal)))
}

// userID extracts the authenticated user's ID from the JWT claims in the context.
func userID(r *http.Request) (string, bool) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func (h *MealHandler) CreateMeal(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateMealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	meal, err := h.mealSvc.CreateMeal(r.Context(), uid, &req)
	if err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to create meal")
		return
	}

	lib.WriteJSON(w, http.StatusCreated, meal)
}

// ViewMeal returns a single meal (with its foods) when ?id= is given,
// otherwise lists all of the user's meals with aggregate totals.
func (h *MealHandler) ViewMeal(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		meals, err := h.mealSvc.ListMeals(r.Context(), uid)
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, "failed to list meals")
			return
		}
		lib.WriteJSON(w, http.StatusOK, meals)
		return
	}

	meal, err := h.mealSvc.GetMeal(r.Context(), uid, id)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to get meal")
		return
	}
	if meal == nil {
		lib.WriteError(w, http.StatusNotFound, "meal not found")
		return
	}

	lib.WriteJSON(w, http.StatusOK, meal)
}

// isValidationErr checks if the error originated from a validation rule.
func isValidationErr(err error) bool {
	return strings.HasPrefix(err.Error(), "validation:")
}
