package nutrition

import (
	"context"
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
