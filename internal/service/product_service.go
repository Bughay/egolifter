// internal/service/product_service.go
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bughay/go-backend-layout/internal/model"
	"github.com/Bughay/go-backend-layout/internal/repository"
)

// ProductService defines the contract for product business logic.
type ProductService interface {
	Create(ctx context.Context, userID int64, req *model.CreateProductRequest) (*model.Product, error)
	GetAll(ctx context.Context, page, limit int) (*model.PaginatedResponse[model.Product], error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	Update(ctx context.Context, requesterID, productID int64, req *model.UpdateProductRequest) (*model.Product, error)
	Delete(ctx context.Context, requesterID, productID int64) error
}

type productService struct {
	productRepo repository.ProductRepository
}

// NewProductService creates a new ProductService.
func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) Create(ctx context.Context, userID int64, req *model.CreateProductRequest) (*model.Product, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("validation: product name is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("validation: price cannot be negative")
	}
	if req.Stock < 0 {
		return nil, fmt.Errorf("validation: stock cannot be negative")
	}

	return s.productRepo.Create(ctx, userID, req)
}

func (s *productService) GetAll(ctx context.Context, page, limit int) (*model.PaginatedResponse[model.Product], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, total, err := s.productRepo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("productService.GetAll: %w", err)
	}

	// Dereference pointers for the response
	items := make([]model.Product, len(products))
	for i, p := range products {
		items[i] = *p
	}

	return &model.PaginatedResponse[model.Product]{
		Data:  items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *productService) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("productService.GetByID: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("not_found: product with id %d does not exist", id)
	}
	return product, nil
}

func (s *productService) Update(ctx context.Context, requesterID, productID int64, req *model.UpdateProductRequest) (*model.Product, error) {
	// 1. Authorization: Verify the requester owns this product (prevents IDOR)
	existing, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("not_found: product with id %d does not exist", productID)
	}
	if existing.UserID != requesterID {
		return nil, fmt.Errorf("forbidden: you do not own this product")
	}

	// 2. Validate only sent fields (nil = field was omitted from the request body)
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("validation: product name cannot be empty")
		}
		req.Name = &trimmed
	}
	if req.Price != nil && *req.Price < 0 {
		return nil, fmt.Errorf("validation: price cannot be negative")
	}
	if req.Stock != nil && *req.Stock < 0 {
		return nil, fmt.Errorf("validation: stock cannot be negative")
	}

	// 3. Reject a completely empty PATCH body
	if req.Name == nil && req.Description == nil && req.Price == nil && req.Stock == nil {
		return nil, fmt.Errorf("validation: no fields provided for update")
	}

	return s.productRepo.Update(ctx, productID, req)
}

func (s *productService) Delete(ctx context.Context, requesterID, productID int64) error {
	// Authorization check
	existing, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("not_found: product with id %d does not exist", productID)
	}
	if existing.UserID != requesterID {
		return fmt.Errorf("forbidden: you do not own this product")
	}

	return s.productRepo.Delete(ctx, productID)
}
