package recipe

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bughay/egolifter/internal/auth"
	"github.com/Bughay/egolifter/internal/lib"
)

// RecipeHandler handles recipe-related HTTP requests.
type RecipeHandler struct {
	recipeSvc RecipeService
}

// NewRecipeHandler creates a new RecipeHandler.
func NewRecipeHandler(recipeSvc RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeSvc: recipeSvc}
}

// RegisterRoutes attaches the recipe CRUD endpoints to the given mux,
// wrapping every route with the provided middleware (JWT auth).
func (h *RecipeHandler) RegisterRoutes(mux *http.ServeMux, mw func(http.Handler) http.Handler) {
	mux.Handle("POST /recipe/create", mw(http.HandlerFunc(h.CreateRecipe)))
	mux.Handle("GET /recipe/view", mw(http.HandlerFunc(h.ViewRecipe)))
	mux.Handle("PUT /recipe/update", mw(http.HandlerFunc(h.UpdateRecipe)))
	mux.Handle("DELETE /recipe/delete", mw(http.HandlerFunc(h.DeleteRecipe)))
}

// userID extracts the authenticated user's ID from the JWT claims in the context.
func userID(r *http.Request) (string, bool) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	recipe, err := h.recipeSvc.CreateRecipe(r.Context(), uid, &req)
	if err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to create recipe")
		return
	}

	lib.WriteJSON(w, http.StatusCreated, recipe)
}

// ViewRecipe returns a single recipe (with ingredients) when ?id= is given,
// otherwise lists all of the user's recipes.
func (h *RecipeHandler) ViewRecipe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		recipes, err := h.recipeSvc.ListRecipes(r.Context(), uid)
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, "failed to list recipes")
			return
		}
		lib.WriteJSON(w, http.StatusOK, recipes)
		return
	}

	recipe, err := h.recipeSvc.GetRecipe(r.Context(), uid, id)
	if err != nil {
		lib.WriteError(w, http.StatusInternalServerError, "failed to get recipe")
		return
	}
	if recipe == nil {
		lib.WriteError(w, http.StatusNotFound, "recipe not found")
		return
	}

	lib.WriteJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	recipe, err := h.recipeSvc.UpdateRecipe(r.Context(), uid, &req)
	if err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to update recipe")
		return
	}
	if recipe == nil {
		lib.WriteError(w, http.StatusNotFound, "recipe not found")
		return
	}

	lib.WriteJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(r)
	if !ok {
		lib.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id := r.URL.Query().Get("id")
	if err := h.recipeSvc.DeleteRecipe(r.Context(), uid, id); err != nil {
		if isValidationErr(err) {
			lib.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		lib.WriteError(w, http.StatusInternalServerError, "failed to delete recipe")
		return
	}

	lib.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "recipe deleted successfully",
	})
}

// isValidationErr checks if the error originated from a validation rule.
func isValidationErr(err error) bool {
	return strings.HasPrefix(err.Error(), "validation:")
}
