package repositories

import (
	"database/sql"
	"fmt"
	"retail-core-api/models"
	"time"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	CreateTransaction(req models.CheckoutRequest) (*models.Transaction, error)
	GetAllTransactions(page, limit int, startDate, endDate string) (*models.PaginatedTransactions, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	VoidTransaction(id int) error
	GetDashboardStats() (*models.DashboardStats, error)
	GetDailySalesReport() (*models.SalesReport, error)
	GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error)
	GetReportSummary(startDate, endDate string) (*models.ReportSummary, error)
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
func (repo *transactionRepository) CreateTransaction(req models.CheckoutRequest) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0, len(req.Items))

	for _, item := range req.Items {
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
			UnitPrice:   productPrice,
			Subtotal:    subtotal,
		})
	}

	// Apply discount
	discount := req.Discount
	if discount > totalAmount {
		discount = totalAmount
	}
	finalAmount := totalAmount - discount

	// Default payment method
	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	// Insert transaction header
	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow(
		`INSERT INTO transactions (total_amount, payment_method, discount, notes, status) 
		 VALUES ($1, $2, $3, $4, 'active') RETURNING id, created_at`,
		finalAmount, paymentMethod, discount, req.Notes,
	).Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, err
	}

	// Insert transaction details
	for i := range details {
		details[i].TransactionID = transactionID

		var detailID int
		err = tx.QueryRow(
			`INSERT INTO transaction_details (transaction_id, product_id, quantity, unit_price, subtotal) 
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			transactionID, details[i].ProductID, details[i].Quantity, details[i].UnitPrice, details[i].Subtotal,
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
		ID:            transactionID,
		TotalAmount:   finalAmount,
		PaymentMethod: paymentMethod,
		Discount:      discount,
		Notes:         req.Notes,
		Status:        "active",
		CreatedAt:     createdAt,
		Details:       details,
	}, nil
}

// VoidTransaction marks a transaction as void and restores product stock
func (repo *transactionRepository) VoidTransaction(id int) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check current status
	var status string
	err = tx.QueryRow("SELECT status FROM transactions WHERE id = $1", id).Scan(&status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction id %d not found", id)
	}
	if err != nil {
		return err
	}
	if status == "void" {
		return fmt.Errorf("transaction is already voided")
	}

	// Restore stock
	rows, err := tx.Query(
		"SELECT product_id, quantity FROM transaction_details WHERE transaction_id = $1", id,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type restoreItem struct {
		productID int
		quantity  int
	}
	var items []restoreItem
	for rows.Next() {
		var ri restoreItem
		if err := rows.Scan(&ri.productID, &ri.quantity); err != nil {
			return err
		}
		items = append(items, ri)
	}
	rows.Close()

	for _, ri := range items {
		_, err = tx.Exec("UPDATE products SET stock = stock + $1 WHERE id = $2", ri.quantity, ri.productID)
		if err != nil {
			return err
		}
	}

	// Mark as void
	_, err = tx.Exec("UPDATE transactions SET status = 'void' WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetDailySalesReport returns the sales summary for today
func (repo *transactionRepository) GetDailySalesReport() (*models.SalesReport, error) {
	report := &models.SalesReport{}

	err := repo.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0), COUNT(*)
		FROM transactions
		WHERE created_at::date = CURRENT_DATE AND status = 'active'
	`).Scan(&report.TotalRevenue, &report.TotalTransactions)
	if err != nil {
		return nil, err
	}

	var best models.BestSellingProduct
	err = repo.db.QueryRow(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) AS qty_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE t.created_at::date = CURRENT_DATE AND t.status = 'active'
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
		WHERE created_at::date >= $1::date AND created_at::date <= $2::date AND status = 'active'
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
		WHERE t.created_at::date >= $1::date AND t.created_at::date <= $2::date AND t.status = 'active'
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

// GetAllTransactions returns a paginated list of transactions with optional date filtering
func (repo *transactionRepository) GetAllTransactions(page, limit int, startDate, endDate string) (*models.PaginatedTransactions, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Build WHERE clause for date filters
	where := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if startDate != "" {
		where += fmt.Sprintf(" AND t.created_at::date >= $%d::date", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND t.created_at::date <= $%d::date", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM transactions t" + where
	var total int
	err := repo.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Fetch page
	query := fmt.Sprintf(`
		SELECT t.id, t.total_amount, t.payment_method, t.discount, t.status,
		       COUNT(td.id) AS item_count, t.created_at
		FROM transactions t
		LEFT JOIN transaction_details td ON td.transaction_id = t.id
		%s
		GROUP BY t.id, t.total_amount, t.payment_method, t.discount, t.status, t.created_at
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.TransactionListItem, 0)
	for rows.Next() {
		var item models.TransactionListItem
		if err := rows.Scan(&item.ID, &item.TotalAmount, &item.PaymentMethod, &item.Discount, &item.Status, &item.ItemCount, &item.CreatedAt); err != nil {
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
		SELECT id, total_amount, payment_method, discount, notes, status, created_at 
		FROM transactions WHERE id = $1
	`, id).Scan(&t.ID, &t.TotalAmount, &t.PaymentMethod, &t.Discount, &t.Notes, &t.Status, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction id %d not found", id)
	}
	if err != nil {
		return nil, err
	}

	rows, err := repo.db.Query(`
		SELECT td.id, td.transaction_id, td.product_id,
		       COALESCE(p.name, 'Deleted Product') AS product_name,
		       td.quantity, td.unit_price, td.subtotal
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
		if err := rows.Scan(&d.ID, &d.TransactionID, &d.ProductID, &d.ProductName, &d.Quantity, &d.UnitPrice, &d.Subtotal); err != nil {
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
		WHERE created_at::date = CURRENT_DATE AND status = 'active'
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
		WHERE t.created_at::date = CURRENT_DATE AND t.status = 'active'
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

// GetReportSummary returns an aggregated report with category breakdown
func (repo *transactionRepository) GetReportSummary(startDate, endDate string) (*models.ReportSummary, error) {
	summary := &models.ReportSummary{}

	// Build date filter
	where := " WHERE t.status = 'active'"
	args := []interface{}{}
	argIdx := 1
	if startDate != "" {
		where += fmt.Sprintf(" AND t.created_at::date >= $%d::date", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		where += fmt.Sprintf(" AND t.created_at::date <= $%d::date", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	// Total revenue and transactions
	totalQuery := "SELECT COALESCE(SUM(t.total_amount), 0), COUNT(*) FROM transactions t" + where
	err := repo.db.QueryRow(totalQuery, args...).Scan(&summary.TotalRevenue, &summary.TotalTransactions)
	if err != nil {
		return nil, err
	}

	// Best selling product
	bestQuery := fmt.Sprintf(`
		SELECT p.name, COALESCE(SUM(td.quantity), 0) AS qty_sold
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		%s
		GROUP BY p.id, p.name
		ORDER BY qty_sold DESC
		LIMIT 1
	`, where)
	var best models.BestSellingProduct
	err = repo.db.QueryRow(bestQuery, args...).Scan(&best.Name, &best.QtySold)
	if err == sql.ErrNoRows {
		summary.BestSellingProduct = nil
	} else if err != nil {
		return nil, err
	} else {
		summary.BestSellingProduct = &best
	}

	// Category breakdown
	catQuery := fmt.Sprintf(`
		SELECT COALESCE(p.category_id, 0), COALESCE(c.name, 'Uncategorized'),
		       COALESCE(SUM(td.subtotal), 0), COUNT(DISTINCT t.id)
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		GROUP BY p.category_id, c.name
		ORDER BY SUM(td.subtotal) DESC
	`, where)
	rows, err := repo.db.Query(catQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.CategoryRevenue, 0)
	for rows.Next() {
		var cr models.CategoryRevenue
		if err := rows.Scan(&cr.CategoryID, &cr.CategoryName, &cr.Revenue, &cr.Transactions); err != nil {
			return nil, err
		}
		categories = append(categories, cr)
	}
	summary.CategoryBreakdown = categories

	return summary, nil
}
