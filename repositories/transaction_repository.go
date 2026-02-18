package repositories

import (
	"database/sql"
	"fmt"
	"retail-core-api/models"
	"time"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error)
	GetAllTransactions(page, limit int) (*models.PaginatedTransactions, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	GetDashboardStats() (*models.DashboardStats, error)
	GetDailySalesReport() (*models.SalesReport, error)
	GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error)
}

// transactionRepository implements TransactionRepository interface
type transactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository instance
func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// CreateTransaction processes a checkout: validates products, deducts stock,
// creates transaction record and detail rows inside a single DB transaction.
func (repo *transactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0, len(items))

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow(
			"SELECT name, price, stock FROM products WHERE id = $1",
			item.ProductID,
		).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		if stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product '%s' (available: %d, requested: %d)",
				productName, stock, item.Quantity)
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec(
			"UPDATE products SET stock = stock - $1 WHERE id = $2",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// Insert transaction header — also retrieve created_at
	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id, created_at",
		totalAmount,
	).Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, err
	}

	// Insert transaction details — use RETURNING id to capture the generated ID
	for i := range details {
		details[i].TransactionID = transactionID

		var detailID int
		err = tx.QueryRow(
			"INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4) RETURNING id",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal,
		).Scan(&detailID)
		if err != nil {
			return nil, err
		}
		details[i].ID = detailID
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		CreatedAt:   createdAt,
		Details:     details,
	}, nil
}

// GetDailySalesReport returns the sales summary for today
func (repo *transactionRepository) GetDailySalesReport() (*models.SalesReport, error) {
	report := &models.SalesReport{}

	// Get total revenue and transaction count for today
	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM transactions
		WHERE created_at::date = CURRENT_DATE
	`).Scan(&report.TotalRevenue, &report.TotalTransactions)
	if err != nil {
		return nil, err
	}

	// Get best selling product for today
	var best models.BestSellingProduct
	err = repo.db.QueryRow(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) AS qty_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at::date = CURRENT_DATE
		GROUP BY p.id, p.name
		ORDER BY qty_sold DESC
		LIMIT 1
	`).Scan(&best.Name, &best.QtySold)
	if err == sql.ErrNoRows {
		report.BestSellingProduct = nil
	} else if err != nil {
		return nil, err
	} else {
		report.BestSellingProduct = &best
	}

	return report, nil
}

// GetSalesReportByDateRange returns the sales summary for a given date range
func (repo *transactionRepository) GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error) {
	report := &models.SalesReport{}

	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM transactions
		WHERE created_at::date >= $1::date AND created_at::date <= $2::date
	`, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransactions)
	if err != nil {
		return nil, err
	}

	var best models.BestSellingProduct
	err = repo.db.QueryRow(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) AS qty_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at::date >= $1::date AND t.created_at::date <= $2::date
		GROUP BY p.id, p.name
		ORDER BY qty_sold DESC
		LIMIT 1
	`, startDate, endDate).Scan(&best.Name, &best.QtySold)
	if err == sql.ErrNoRows {
		report.BestSellingProduct = nil
	} else if err != nil {
		return nil, err
	} else {
		report.BestSellingProduct = &best
	}

	return report, nil
}

// GetAllTransactions returns a paginated list of transactions with item counts
func (repo *transactionRepository) GetAllTransactions(page, limit int) (*models.PaginatedTransactions, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int
	err := repo.db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := repo.db.Query(`
		SELECT t.id, t.total_amount, t.created_at,
		       COUNT(td.id) AS item_count
		FROM transactions t
		LEFT JOIN transaction_details td ON td.transaction_id = t.id
		GROUP BY t.id, t.total_amount, t.created_at
		ORDER BY t.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TransactionListItem, 0)
	for rows.Next() {
		var item models.TransactionListItem
		if err := rows.Scan(&item.ID, &item.TotalAmount, &item.CreatedAt, &item.ItemCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	totalPages := (total + limit - 1) / limit

	return &models.PaginatedTransactions{
		Data:       items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetTransactionByID returns a single transaction with all its details
func (repo *transactionRepository) GetTransactionByID(id int) (*models.Transaction, error) {
	var t models.Transaction
	err := repo.db.QueryRow(`
		SELECT id, total_amount, created_at FROM transactions WHERE id = $1
	`, id).Scan(&t.ID, &t.TotalAmount, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction id %d not found", id)
	}
	if err != nil {
		return nil, err
	}

	rows, err := repo.db.Query(`
		SELECT td.id, td.transaction_id, td.product_id,
		       COALESCE(p.name, 'Deleted Product') AS product_name,
		       td.quantity, td.subtotal
		FROM transaction_details td
		LEFT JOIN products p ON p.id = td.product_id
		WHERE td.transaction_id = $1
		ORDER BY td.id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	details := make([]models.TransactionDetail, 0)
	for rows.Next() {
		var d models.TransactionDetail
		if err := rows.Scan(&d.ID, &d.TransactionID, &d.ProductID, &d.ProductName, &d.Quantity, &d.Subtotal); err != nil {
			return nil, err
		}
		details = append(details, d)
	}
	t.Details = details
	return &t, nil
}

// GetDashboardStats returns summary statistics for the admin dashboard
func (repo *transactionRepository) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM transactions
		WHERE created_at::date = CURRENT_DATE
	`).Scan(&stats.TotalRevenueToday, &stats.TransactionsToday)
	if err != nil {
		return nil, err
	}

	err = repo.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&stats.TotalProducts)
	if err != nil {
		return nil, err
	}

	err = repo.db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&stats.TotalCategories)
	if err != nil {
		return nil, err
	}

	err = repo.db.QueryRow(`SELECT COUNT(*) FROM products WHERE stock < 10`).Scan(&stats.LowStockCount)
	if err != nil {
		return nil, err
	}

	var best models.BestSellingProduct
	err = repo.db.QueryRow(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) AS qty_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at::date = CURRENT_DATE
		GROUP BY p.id, p.name
		ORDER BY qty_sold DESC
		LIMIT 1
	`).Scan(&best.Name, &best.QtySold)
	if err == sql.ErrNoRows {
		stats.BestSellingToday = nil
	} else if err != nil {
		return nil, err
	} else {
		stats.BestSellingToday = &best
	}

	return stats, nil
}
