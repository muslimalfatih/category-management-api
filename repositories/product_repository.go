package repositories

import (
	"database/sql"
	"retail-core-api/models"
	"time"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	GetAll(name string) ([]models.Product, error)
	GetByID(id int) (*models.Product, error)
	GetByCategoryID(categoryID int) ([]models.Product, error)
	Create(product models.Product) (*models.Product, error)
	Update(id int, product models.Product) (*models.Product, error)
	Delete(id int) error
}

// productRepository implements ProductRepository interface with PostgreSQL
type productRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository instance
func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

// GetAll returns all products from database with category names (LEFT JOIN)
// Supports optional name filter for search functionality
func (r *productRepository) GetAll(nameFilter string) ([]models.Product, error) {
	query := `
		SELECT 
			p.id, 
			p.name, 
			p.price, 
			p.stock, 
			p.category_id,
			COALESCE(c.name, '') as category_name,
			p.created_at, 
			p.updated_at 
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id`

	args := []interface{}{}
	if nameFilter != "" {
		query += " WHERE p.name ILIKE $1"
		args = append(args, "%"+nameFilter+"%")
	}

	query += " ORDER BY p.id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var prod models.Product
		err := rows.Scan(
			&prod.ID,
			&prod.Name,
			&prod.Price,
			&prod.Stock,
			&prod.CategoryID,
			&prod.CategoryName,
			&prod.CreatedAt,
			&prod.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, prod)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// GetByID returns a product by its ID with category name (LEFT JOIN)
func (r *productRepository) GetByID(id int) (*models.Product, error) {
	query := `
		SELECT 
			p.id, 
			p.name, 
			p.price, 
			p.stock, 
			p.category_id,
			COALESCE(c.name, '') as category_name,
			p.created_at, 
			p.updated_at 
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	var prod models.Product
	err := r.db.QueryRow(query, id).Scan(
		&prod.ID,
		&prod.Name,
		&prod.Price,
		&prod.Stock,
		&prod.CategoryID,
		&prod.CategoryName,
		&prod.CreatedAt,
		&prod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &prod, nil
}

// Create adds a new product and returns it
func (r *productRepository) Create(product models.Product) (*models.Product, error) {
	query := `
		INSERT INTO products (name, price, stock, category_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, price, stock, category_id, created_at, updated_at
	`
	var prod models.Product
	err := r.db.QueryRow(query, product.Name, product.Price, product.Stock, product.CategoryID).Scan(
		&prod.ID,
		&prod.Name,
		&prod.Price,
		&prod.Stock,
		&prod.CategoryID,
		&prod.CreatedAt,
		&prod.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// If product has category_id, fetch the category name
	if prod.CategoryID != nil {
		var categoryName string
		categoryQuery := `SELECT name FROM categories WHERE id = $1`
		err = r.db.QueryRow(categoryQuery, *prod.CategoryID).Scan(&categoryName)
		if err == nil {
			prod.CategoryName = categoryName
		}
	}

	return &prod, nil
}

// Update modifies an existing product
func (r *productRepository) Update(id int, product models.Product) (*models.Product, error) {
	query := `
		UPDATE products 
		SET name = $1, price = $2, stock = $3, category_id = $4, updated_at = $5 
		WHERE id = $6 
		RETURNING id, name, price, stock, category_id, created_at, updated_at
	`
	var prod models.Product
	err := r.db.QueryRow(
		query,
		product.Name,
		product.Price,
		product.Stock,
		product.CategoryID,
		time.Now(),
		id,
	).Scan(
		&prod.ID,
		&prod.Name,
		&prod.Price,
		&prod.Stock,
		&prod.CategoryID,
		&prod.CreatedAt,
		&prod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// If product has category_id, fetch the category name
	if prod.CategoryID != nil {
		var categoryName string
		categoryQuery := `SELECT name FROM categories WHERE id = $1`
		err = r.db.QueryRow(categoryQuery, *prod.CategoryID).Scan(&categoryName)
		if err == nil {
			prod.CategoryName = categoryName
		}
	}

	return &prod, nil
}

// Delete removes a product by its ID
func (r *productRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
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

// GetByCategoryID returns all products belonging to a specific category
func (r *productRepository) GetByCategoryID(categoryID int) ([]models.Product, error) {
	query := `
		SELECT
			p.id,
			p.name,
			p.price,
			p.stock,
			p.category_id,
			COALESCE(c.name, '') as category_name,
			p.created_at,
			p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1
		ORDER BY p.id
	`
	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var prod models.Product
		err := rows.Scan(
			&prod.ID,
			&prod.Name,
			&prod.Price,
			&prod.Stock,
			&prod.CategoryID,
			&prod.CategoryName,
			&prod.CreatedAt,
			&prod.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, prod)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
