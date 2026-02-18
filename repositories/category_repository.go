package repositories

import (
	"database/sql"
	"retail-core-api/models"
	"time"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	GetAll() ([]models.Category, error)
	GetByID(id int) (*models.Category, error)
	Create(category models.Category) (*models.Category, error)
	Update(id int, category models.Category) (*models.Category, error)
	Delete(id int) error
}

// categoryRepository implements CategoryRepository interface with PostgreSQL
type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository instance
func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// GetAll returns all categories from database
func (r *categoryRepository) GetAll() ([]models.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetByID returns a category by its ID
func (r *categoryRepository) GetByID(id int) (*models.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`
	var cat models.Category
	err := r.db.QueryRow(query, id).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

// Create adds a new category and returns it
func (r *categoryRepository) Create(category models.Category) (*models.Category, error) {
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id, name, description, created_at, updated_at`
	var cat models.Category
	err := r.db.QueryRow(query, category.Name, category.Description).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// Update modifies an existing category
func (r *categoryRepository) Update(id int, category models.Category) (*models.Category, error) {
	query := `UPDATE categories SET name = $1, description = $2, updated_at = $3 WHERE id = $4 RETURNING id, name, description, created_at, updated_at`
	var cat models.Category
	err := r.db.QueryRow(query, category.Name, category.Description, time.Now(), id).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

// Delete removes a category by its ID
func (r *categoryRepository) Delete(id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}
