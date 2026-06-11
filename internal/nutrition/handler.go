package nutrition

import (
	"encoding/json"
	"net/http"

	"github.com/Bughay/egolifter/internal/lib"
)

// NutritionHandler handles nutrition-related HTTP requests.
type NutritionHandler struct {
	nutritionSvc NutritionService
}

// NewNutritionHandler creates a new NutritionHandler.
func NewNutritionHandler(nutritionSvc NutritionService) *NutritionHandler {
	return &NutritionHandler{nutritionSvc: nutritionSvc}
}

// RegisterRoutes attaches the food CRUD endpoints to the given mux.
func (h *NutritionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /food/create", h.CreateFood)
	mux.HandleFunc("GET /food/view", h.ViewFood)
	mux.HandleFunc("PUT /food/update", h.UpdateFood)
	mux.HandleFunc("DELETE /food/delete", h.DeleteFood)
}

func (h *NutritionHandler) CreateFood(w http.ResponseWriter, r *http.Request) {
	var req CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	food, err := h.nutritionSvc.CreateFood(r.Context(), &req)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to create food")
		return
	}

	lib.WriteJSON(w, http.StatusCreated, food)
}

// ViewFood returns a single food when ?id= is given, otherwise lists all foods.
func (h *NutritionHandler) ViewFood(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		foods, err := h.nutritionSvc.ListFoods(r.Context())
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, "failed to list foods")
			return
		}
		lib.WriteJSON(w, http.StatusOK, foods)
		return
	}

	food, err := h.nutritionSvc.GetFood(r.Context(), id)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to get food")
		return
	}
	if food == nil {
		lib.WriteError(w, http.StatusNotFound, "food not found")
		return
	}

	lib.WriteJSON(w, http.StatusOK, food)
}

func (h *NutritionHandler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	var req UpdateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	food, err := h.nutritionSvc.UpdateFood(r.Context(), &req)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to update food")
		return
	}
	if food == nil {
		lib.WriteError(w, http.StatusNotFound, "food not found")
		return
	}

	lib.WriteJSON(w, http.StatusOK, food)
}

func (h *NutritionHandler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if err := h.nutritionSvc.DeleteFood(r.Context(), id); err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to delete food")
		return
	}

	lib.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "food deleted successfully",
	})
}
