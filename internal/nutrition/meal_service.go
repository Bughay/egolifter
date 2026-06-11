package nutrition

import (
	"context"
	"fmt"
	"strings"
)

// MealService defines the contract for meal business logic.
type MealService interface {
	CreateMeal(ctx context.Context, userID string, req *CreateMealRequest) (*Meal, error)
	GetMeal(ctx context.Context, userID, id string) (*Meal, error)
	ListMeals(ctx context.Context, userID string) ([]Meal, error)
}

type mealService struct {
	mealRepo MealRepository
}

// NewMealService creates a new MealService.
func NewMealService(mealRepo MealRepository) MealService {
	return &mealService{mealRepo: mealRepo}
}

func (s *mealService) CreateMeal(ctx context.Context, userID string, req *CreateMealRequest) (*Meal, error) {
	if err := validateMealInput(req.Name, req.Foods); err != nil {
		return nil, err
	}
	return s.mealRepo.Create(ctx, userID, req)
}

func (s *mealService) GetMeal(ctx context.Context, userID, id string) (*Meal, error) {
	return s.mealRepo.FindByID(ctx, userID, id)
}

func (s *mealService) ListMeals(ctx context.Context, userID string) ([]Meal, error) {
	return s.mealRepo.List(ctx, userID)
}

func validateMealInput(name string, foods []MealFoodInput) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("validation: meal name is required")
	}
	if len(foods) == 0 {
		return fmt.Errorf("validation: a meal must contain at least one food")
	}
	for i, f := range foods {
		if strings.TrimSpace(f.FoodID) == "" {
			return fmt.Errorf("validation: food %d: food_id is required", i)
		}
		if f.WeightG <= 0 {
			return fmt.Errorf("validation: food %d: weight_g must be greater than zero", i)
		}
	}
	return nil
}
