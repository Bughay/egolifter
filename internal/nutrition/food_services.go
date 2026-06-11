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
	if err := validateFoodInput(req.Name, req.Calories100, req.Protein100, req.Carbohydrates100, req.Fat100); err != nil {
		return nil, err
	}
	return s.foodRepo.Create(ctx, req)
}

func (s *nutritionService) GetFood(ctx context.Context, id string) (*Food, error) {
	if err := validateFoodID(id); err != nil {
		return nil, err
	}
	return s.foodRepo.FindByID(ctx, id)
}

func (s *nutritionService) ListFoods(ctx context.Context) ([]Food, error) {
	return s.foodRepo.List(ctx)
}

func (s *nutritionService) UpdateFood(ctx context.Context, req *UpdateFoodRequest) (*Food, error) {
	if err := validateFoodID(req.ID); err != nil {
		return nil, err
	}
	if err := validateFoodInput(req.Name, req.Calories100, req.Protein100, req.Carbohydrates100, req.Fat100); err != nil {
		return nil, err
	}
	return s.foodRepo.Update(ctx, req)
}

func (s *nutritionService) DeleteFood(ctx context.Context, id string) error {
	if err := validateFoodID(id); err != nil {
		return err
	}
	return s.foodRepo.Delete(ctx, id)
}
