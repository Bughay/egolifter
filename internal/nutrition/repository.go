package nutrition

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bughay/egolifter/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FoodRepository defines the contract for Food data access.
type FoodRepository interface {
	Create(ctx context.Context, req *CreateFoodRequest) (*Food, error)
	FindByID(ctx context.Context, id string) (*Food, error)
	List(ctx context.Context) ([]Food, error)
	Update(ctx context.Context, req *UpdateFoodRequest) (*Food, error)
	Delete(ctx context.Context, id string) error
}

type pgFoodRepository struct {
	queries *db.Queries
}

// NewFoodRepository creates a new PostgreSQL-backed FoodRepository wrapping the sqlc-generated queries.
func NewFoodRepository(pool *pgxpool.Pool) FoodRepository {
	return &pgFoodRepository{queries: db.New(pool)}
}

func (r *pgFoodRepository) Create(ctx context.Context, req *CreateFoodRequest) (*Food, error) {
	row, err := r.queries.CreateFood(ctx, db.CreateFoodParams{
		Name:             req.Name,
		Calories100:      req.Calories100,
		Protein100:       req.Protein100,
		Carbohydrates100: req.Carbohydrates100,
		Fat100:           req.Fat100,
	})
	if err != nil {
		return nil, fmt.Errorf("foodRepo.Create: %w", err)
	}
	return toFood(row), nil
}

func (r *pgFoodRepository) FindByID(ctx context.Context, id string) (*Food, error) {
	row, err := r.queries.GetFoodByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is not an error at this layer
		}
		return nil, fmt.Errorf("foodRepo.FindByID: %w", err)
	}
	return toFood(row), nil
}

func (r *pgFoodRepository) List(ctx context.Context) ([]Food, error) {
	rows, err := r.queries.ListFoods(ctx)
	if err != nil {
		return nil, fmt.Errorf("foodRepo.List: %w", err)
	}
	foods := make([]Food, 0, len(rows))
	for _, row := range rows {
		foods = append(foods, *toFood(row))
	}
	return foods, nil
}

func (r *pgFoodRepository) Update(ctx context.Context, req *UpdateFoodRequest) (*Food, error) {
	row, err := r.queries.UpdateFood(ctx, db.UpdateFoodParams{
		ID:               req.ID,
		Name:             req.Name,
		Calories100:      req.Calories100,
		Protein100:       req.Protein100,
		Carbohydrates100: req.Carbohydrates100,
		Fat100:           req.Fat100,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("foodRepo.Update: %w", err)
	}
	return toFood(row), nil
}

func (r *pgFoodRepository) Delete(ctx context.Context, id string) error {
	if err := r.queries.DeleteFood(ctx, id); err != nil {
		return fmt.Errorf("foodRepo.Delete: %w", err)
	}
	return nil
}

// toFood maps a sqlc-generated row to the domain model.
func toFood(row db.Food) *Food {
	return &Food{
		ID:               row.ID,
		Name:             row.Name,
		Calories100:      row.Calories100,
		Protein100:       row.Protein100,
		Carbohydrates100: row.Carbohydrates100,
		Fat100:           row.Fat100,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
