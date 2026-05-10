// internal/repository/product_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Bughay/go-backend-layout/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepository defines the contract for Product data access.
type ProductRepository interface {
	Create(ctx context.Context, userID int64, req *model.CreateProductRequest) (*model.Product, error)
	FindAll(ctx context.Context, page, limit int) ([]*model.Product, int64, error)
	FindByID(ctx context.Context, id int64) (*model.Product, error)
	Update(ctx context.Context, id int64, req *model.UpdateProductRequest) (*model.Product, error)
	Delete(ctx context.Context, id int64) error
}

type pgProductRepository struct {
	pool *pgxpool.Pool
}

// NewProductRepository creates a new PostgreSQL-backed ProductRepository.
func NewProductRepository(pool *pgxpool.Pool) ProductRepository {
	return &pgProductRepository{pool: pool}
}

func (r *pgProductRepository) Create(ctx context.Context, userID int64, req *model.CreateProductRequest) (*model.Product, error) {
	query := `
        INSERT INTO products (user_id, name, description, price, stock)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, user_id, name, description, price, stock, created_at, updated_at`

	var p model.Product
	err := r.pool.QueryRow(ctx, query, userID, req.Name, req.Description, req.Price, req.Stock).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("productRepo.Create: %w", err)
	}
	return &p, nil
}

func (r *pgProductRepository) FindAll(ctx context.Context, page, limit int) ([]*model.Product, int64, error) {
	// Use a Common Table Expression to get both the products and total count in a single query
	query := `
        WITH product_count AS (SELECT COUNT(*) FROM products)
        SELECT id, user_id, name, description, price, stock, created_at, updated_at,
               (SELECT * FROM product_count)
        FROM products
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2`

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("productRepo.FindAll: %w", err)
	}
	defer rows.Close()

	var products []*model.Product
	var total int64
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("productRepo.FindAll scan: %w", err)
		}
		products = append(products, &p)
	}
	return products, total, rows.Err()
}

func (r *pgProductRepository) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	query := `SELECT id, user_id, name, description, price, stock, created_at, updated_at FROM products WHERE id = $1`

	var p model.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepo.FindByID: %w", err)
	}
	return &p, nil
}

func (r *pgProductRepository) Update(ctx context.Context, id int64, req *model.UpdateProductRequest) (*model.Product, error) {
	// Build the SET clause dynamically — only include fields the client actually sent.
	// nil pointer = "field not provided by client" = skip it in SQL.
	setClauses := []string{}
	args := []any{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, *req.Price)
		argIdx++
	}
	if req.Stock != nil {
		setClauses = append(setClauses, fmt.Sprintf("stock = $%d", argIdx))
		args = append(args, *req.Stock)
		argIdx++
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
        UPDATE products
        SET %s
        WHERE id = $%d
        RETURNING id, user_id, name, description, price, stock, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var p model.Product
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("productRepo.Update: %w", err)
	}
	return &p, nil
}

func (r *pgProductRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("productRepo.Delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("productRepo.Delete: no product found with id %d", id)
	}
	return nil
}
