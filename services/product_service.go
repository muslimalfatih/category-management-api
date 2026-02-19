package services

import (
	"errors"
	"retail-core-api/models"
	"retail-core-api/repositories"
)

// ProductService defines the interface for product business logic
type ProductService interface {
	GetAllProducts(params models.ProductListParams) (*models.PaginatedProducts, error)
	GetProductByID(id int) (*models.Product, error)
	GetProductsByCategoryID(categoryID int) ([]models.Product, error)
	CreateProduct(product models.Product) (*models.Product, error)
	UpdateProduct(id int, product models.Product) (*models.Product, error)
	DeleteProduct(id int) error
}

// productService implements ProductService interface
type productService struct {
	repo         repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

// NewProductService creates a new product service instance
func NewProductService(repo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) ProductService {
	return &productService{
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

// GetAllProducts returns paginated products with optional search/filter
func (s *productService) GetAllProducts(params models.ProductListParams) (*models.PaginatedProducts, error) {
	return s.repo.GetAll(params)
}

// GetProductByID returns a product by its ID
func (s *productService) GetProductByID(id int) (*models.Product, error) {
	return s.repo.GetByID(id)
}

// CreateProduct validates and creates a new product
func (s *productService) CreateProduct(product models.Product) (*models.Product, error) {
	// Business logic validation
	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	if product.Price < 0 {
		return nil, errors.New("product price cannot be negative")
	}

	if product.Stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	// Validate category exists if category_id is provided
	if product.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(*product.CategoryID)
		if err != nil {
			return nil, errors.New("failed to validate category")
		}
		if category == nil {
			return nil, errors.New("category not found")
		}
	}

	return s.repo.Create(product)
}

// UpdateProduct validates and updates an existing product
func (s *productService) UpdateProduct(id int, product models.Product) (*models.Product, error) {
	// Business logic validation
	if product.Name == "" {
		return nil, errors.New("product name is required")
	}

	if product.Price < 0 {
		return nil, errors.New("product price cannot be negative")
	}

	if product.Stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	// Validate category exists if category_id is provided
	if product.CategoryID != nil {
		category, err := s.categoryRepo.GetByID(*product.CategoryID)
		if err != nil {
			return nil, errors.New("failed to validate category")
		}
		if category == nil {
			return nil, errors.New("category not found")
		}
	}

	updated, err := s.repo.Update(id, product)
	if err != nil {
		return nil, err
	}

	if updated == nil {
		return nil, errors.New("product not found")
	}

	return updated, nil
}

// DeleteProduct removes a product by its ID
func (s *productService) DeleteProduct(id int) error {
	return s.repo.Delete(id)
}

// GetProductsByCategoryID returns all products belonging to a category
func (s *productService) GetProductsByCategoryID(categoryID int) ([]models.Product, error) {
	if categoryID <= 0 {
		return nil, errors.New("invalid category ID")
	}
	return s.repo.GetByCategoryID(categoryID)
}
