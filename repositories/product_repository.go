package repositories

import (
	"database/sql"
	"fmt"
	"math"
	"retail-core-api/models"
	"time"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	GetAll(params models.ProductListParams) (*models.PaginatedProducts, error)
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

// productColumns is the standard set of columns selected for product queries
const productColumns = `
	p.id, p.name, p.price, p.stock,
	p.sku, p.image_url, p.unit, p.is_active,
	p.category_id,
	COALESCE(c.name, '') as category_name,
	p.created_at, p.updated_at
`

// scanProduct scans a row into a Product struct
func scanProduct(scanner interface{ Scan(dest ...interface{}) error }) (*models.Product, error) {
	var prod models.Product
	err := scanner.Scan(
		&prod.ID,
		&prod.Name,
		&prod.Price,
		&prod.Stock,
		&prod.SKU,
		&prod.ImageURL,
		&prod.Unit,
		&prod.IsActive,
		&prod.CategoryID,
		&prod.CategoryName,
		&prod.CreatedAt,
		&prod.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &prod, nil
}

// GetAll returns paginated products with optional search and category filter
func (r *productRepository) GetAll(params models.ProductListParams) (*models.PaginatedProducts, error) {
	// Defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	// Build WHERE clause
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		where += fmt.Sprintf(" AND p.name ILIKE $%d", argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	if params.CategoryID != nil {
		where += fmt.Sprintf(" AND p.category_id = $%d", argIdx)
		args = append(args, *params.CategoryID)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM products p" + where
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Fetch page
	offset := (params.Page - 1) * params.Limit
	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.id DESC
		LIMIT $%d OFFSET $%d
	`, productColumns, where, argIdx, argIdx+1)
	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		prod, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *prod)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &models.PaginatedProducts{
		Data:       products,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetByID returns a product by its ID with category name (LEFT JOIN)
func (r *productRepository) GetByID(id int) (*models.Product, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`, productColumns)

	prod, err := scanProduct(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return prod, nil
}

// Create adds a new product and returns it
func (r *productRepository) Create(product models.Product) (*models.Product, error) {
	query := `
		INSERT INTO products (name, price, stock, sku, image_url, unit, is_active, category_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id, name, price, stock, sku, image_url, unit, is_active, category_id, created_at, updated_at
	`
	var prod models.Product
	err := r.db.QueryRow(
		query,
		product.Name, product.Price, product.Stock,
		product.SKU, product.ImageURL, product.Unit, product.IsActive,
		product.CategoryID,
	).Scan(
		&prod.ID, &prod.Name, &prod.Price, &prod.Stock,
		&prod.SKU, &prod.ImageURL, &prod.Unit, &prod.IsActive,
		&prod.CategoryID, &prod.CreatedAt, &prod.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch the category name
	if prod.CategoryID != nil {
		var categoryName string
		err = r.db.QueryRow(`SELECT name FROM categories WHERE id = $1`, *prod.CategoryID).Scan(&categoryName)
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
		SET name = $1, price = $2, stock = $3, sku = $4, image_url = $5, 
		    unit = $6, is_active = $7, category_id = $8, updated_at = $9
		WHERE id = $10 
		RETURNING id, name, price, stock, sku, image_url, unit, is_active, category_id, created_at, updated_at
	`
	var prod models.Product
	err := r.db.QueryRow(
		query,
		product.Name, product.Price, product.Stock,
		product.SKU, product.ImageURL, product.Unit, product.IsActive,
		product.CategoryID, time.Now(), id,
	).Scan(
		&prod.ID, &prod.Name, &prod.Price, &prod.Stock,
		&prod.SKU, &prod.ImageURL, &prod.Unit, &prod.IsActive,
		&prod.CategoryID, &prod.CreatedAt, &prod.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Fetch the category name
	if prod.CategoryID != nil {
		var categoryName string
		err = r.db.QueryRow(`SELECT name FROM categories WHERE id = $1`, *prod.CategoryID).Scan(&categoryName)
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
	query := fmt.Sprintf(`
		SELECT %s
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1
		ORDER BY p.id
	`, productColumns)

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		prod, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *prod)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
