package services

import (
	"errors"
	"retail-core-api/models"
	"retail-core-api/repositories"
)

// CategoryService defines the interface for category business logic
type CategoryService interface {
	GetAllCategories() ([]models.Category, error)
	GetCategoryByID(id int) (*models.Category, error)
	CreateCategory(category models.Category) (*models.Category, error)
	UpdateCategory(id int, category models.Category) (*models.Category, error)
	DeleteCategory(id int) error
}

// categoryService implements CategoryService interface
type categoryService struct {
	repo repositories.CategoryRepository
}

// NewCategoryService creates a new category service instance
func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// GetAllCategories returns all categories
func (s *categoryService) GetAllCategories() ([]models.Category, error) {
	return s.repo.GetAll()
}

// GetCategoryByID returns a category by its ID
func (s *categoryService) GetCategoryByID(id int) (*models.Category, error) {
	return s.repo.GetByID(id)
}

// CreateCategory validates and creates a new category
func (s *categoryService) CreateCategory(category models.Category) (*models.Category, error) {
	// Business logic validation
	if category.Name == "" {
		return nil, errors.New("category name is required")
	}

	return s.repo.Create(category)
}

// UpdateCategory validates and updates an existing category
func (s *categoryService) UpdateCategory(id int, category models.Category) (*models.Category, error) {
	// Business logic validation
	if category.Name == "" {
		return nil, errors.New("category name is required")
	}

	updated, err := s.repo.Update(id, category)
	if err != nil {
		return nil, err
	}
	
	if updated == nil {
		return nil, errors.New("category not found")
	}

	return updated, nil
}

// DeleteCategory removes a category by its ID
func (s *categoryService) DeleteCategory(id int) error {
	return s.repo.Delete(id)
}
