package recipe

import (
	"context"
	"fmt"
	"strings"
)

// RecipeService defines the contract for recipe business logic.
type RecipeService interface {
	CreateRecipe(ctx context.Context, userID string, req *CreateRecipeRequest) (*Recipe, error)
	GetRecipe(ctx context.Context, userID, id string) (*Recipe, error)
	ListRecipes(ctx context.Context, userID string) ([]Recipe, error)
	UpdateRecipe(ctx context.Context, userID string, req *UpdateRecipeRequest) (*Recipe, error)
	DeleteRecipe(ctx context.Context, userID, id string) error
}

type recipeService struct {
	recipeRepo RecipeRepository
}

// NewRecipeService creates a new RecipeService.
func NewRecipeService(recipeRepo RecipeRepository) RecipeService {
	return &recipeService{recipeRepo: recipeRepo}
}

func (s *recipeService) CreateRecipe(ctx context.Context, userID string, req *CreateRecipeRequest) (*Recipe, error) {
	if err := validateRecipeInput(req.Name, req.Ingredients); err != nil {
		return nil, err
	}
	return s.recipeRepo.Create(ctx, userID, req)
}

func (s *recipeService) GetRecipe(ctx context.Context, userID, id string) (*Recipe, error) {
	return s.recipeRepo.FindByID(ctx, userID, id)
}

func (s *recipeService) ListRecipes(ctx context.Context, userID string) ([]Recipe, error) {
	return s.recipeRepo.List(ctx, userID)
}

func (s *recipeService) UpdateRecipe(ctx context.Context, userID string, req *UpdateRecipeRequest) (*Recipe, error) {
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("validation: recipe id is required")
	}
	if err := validateRecipeInput(req.Name, req.Ingredients); err != nil {
		return nil, err
	}
	return s.recipeRepo.Update(ctx, userID, req)
}

func (s *recipeService) DeleteRecipe(ctx context.Context, userID, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("validation: recipe id is required")
	}
	return s.recipeRepo.Delete(ctx, userID, id)
}

func validateRecipeInput(name string, ingredients []IngredientInput) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("validation: recipe name is required")
	}
	for i, ing := range ingredients {
		if strings.TrimSpace(ing.FoodID) == "" {
			return fmt.Errorf("validation: ingredient %d: food_id is required", i)
		}
		if ing.WeightG <= 0 {
			return fmt.Errorf("validation: ingredient %d: weight_g must be greater than zero", i)
		}
	}
	return nil
}
