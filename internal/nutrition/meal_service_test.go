package nutrition

import (
	"context"
	"strings"
	"testing"
)

// stubMealRepository records whether Create was called and returns a fixed meal.
type stubMealRepository struct {
	createCalled bool
}

func (s *stubMealRepository) Create(ctx context.Context, userID string, req *CreateMealRequest) (*Meal, error) {
	s.createCalled = true
	return &Meal{ID: "meal-1", UserID: userID, Name: req.Name, Foods: []ConsumedFood{}}, nil
}

func (s *stubMealRepository) FindByID(ctx context.Context, userID, id string) (*Meal, error) {
	return nil, nil
}

func (s *stubMealRepository) List(ctx context.Context, userID string) ([]Meal, error) {
	return nil, nil
}

func TestCreateMealValidation(t *testing.T) {
	validFoods := []MealFoodInput{
		{FoodID: "food-1", WeightG: 150},
		{FoodID: "food-2", WeightG: 60},
	}

	tests := []struct {
		name    string
		req     *CreateMealRequest
		wantErr string // substring of the expected error; empty means success
	}{
		{
			name: "valid request",
			req:  &CreateMealRequest{Name: "breakfast", Foods: validFoods},
		},
		{
			name:    "empty name",
			req:     &CreateMealRequest{Name: "", Foods: validFoods},
			wantErr: "meal name is required",
		},
		{
			name:    "whitespace name",
			req:     &CreateMealRequest{Name: "   ", Foods: validFoods},
			wantErr: "meal name is required",
		},
		{
			name:    "no foods",
			req:     &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{}},
			wantErr: "at least one food",
		},
		{
			name: "blank food_id",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{FoodID: "  ", WeightG: 100},
			}},
			wantErr: "food_id is required",
		},
		{
			name: "zero weight",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{FoodID: "food-1", WeightG: 0},
			}},
			wantErr: "weight_g must be greater than zero",
		},
		{
			name: "negative weight",
			req: &CreateMealRequest{Name: "breakfast", Foods: []MealFoodInput{
				{FoodID: "food-1", WeightG: -50},
			}},
			wantErr: "weight_g must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubMealRepository{}
			svc := NewMealService(repo)

			meal, err := svc.CreateMeal(context.Background(), "user-1", tt.req)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if !repo.createCalled {
					t.Fatal("expected repository Create to be called")
				}
				if meal == nil || meal.Name != tt.req.Name {
					t.Fatalf("unexpected meal returned: %+v", meal)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.HasPrefix(err.Error(), "validation:") {
				t.Errorf("expected validation error, got: %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
			if repo.createCalled {
				t.Error("repository Create should not be called on validation failure")
			}
		})
	}
}
