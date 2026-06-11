package nutrition

import "context"

// NutritionService defines the contract for nutrition business logic.
type NutritionService interface {
	CreateFood(ctx context.Context, req *CreateFoodRequest) (*Food, error)
	GetFood(ctx context.Context, id string) (*Food, error)
	ListFoods(ctx context.Context) ([]Food, error)
	UpdateFood(ctx context.Context, req *UpdateFoodRequest) (*Food, error)
	DeleteFood(ctx context.Context, id string) error
}

type nutritionService struct {
	foodRepo FoodRepository
}

// NewNutritionService creates a new NutritionService.
func NewNutritionService(foodRepo FoodRepository) NutritionService {
	return &nutritionService{foodRepo: foodRepo}
}

func (s *nutritionService) CreateFood(ctx context.Context, req *CreateFoodRequest) (*Food, error) {
	return s.foodRepo.Create(ctx, req)
}

func (s *nutritionService) GetFood(ctx context.Context, id string) (*Food, error) {
	return s.foodRepo.FindByID(ctx, id)
}

func (s *nutritionService) ListFoods(ctx context.Context) ([]Food, error) {
	return s.foodRepo.List(ctx)
}

func (s *nutritionService) UpdateFood(ctx context.Context, req *UpdateFoodRequest) (*Food, error) {
	return s.foodRepo.Update(ctx, req)
}

func (s *nutritionService) DeleteFood(ctx context.Context, id string) error {
	return s.foodRepo.Delete(ctx, id)
}
