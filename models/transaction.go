package models

import "time"

// Transaction represents a completed transaction
// @Description Transaction information with details of purchased items
type Transaction struct {
	ID            int                 `json:"id" example:"1"`
	TotalAmount   int                 `json:"total_amount" example:"45000"`
	PaymentMethod string              `json:"payment_method" example:"cash"`
	Discount      int                 `json:"discount" example:"0"`
	Notes         string              `json:"notes" example:""`
	Status        string              `json:"status" example:"active"`
	CreatedAt     time.Time           `json:"created_at" example:"2026-02-08T12:00:00Z"`
	Details       []TransactionDetail `json:"details"`
}

// TransactionDetail represents a single item in a transaction
// @Description Detail of a single item within a transaction
type TransactionDetail struct {
	ID            int    `json:"id" example:"1"`
	TransactionID int    `json:"transaction_id" example:"1"`
	ProductID     int    `json:"product_id" example:"3"`
	ProductName   string `json:"product_name,omitempty" example:"Indomie Goreng"`
	Quantity      int    `json:"quantity" example:"5"`
	UnitPrice     int    `json:"unit_price" example:"3000"`
	Subtotal      int    `json:"subtotal" example:"15000"`
}

// CheckoutItem represents a single item in a checkout request
// @Description Single item to be checked out
type CheckoutItem struct {
	ProductID int `json:"product_id" example:"3"`
	Quantity  int `json:"quantity" example:"5"`
}

// CheckoutRequest represents the request body for checkout
// @Description Request body for processing a checkout
type CheckoutRequest struct {
	Items         []CheckoutItem `json:"items"`
	PaymentMethod string         `json:"payment_method" example:"cash"`
	Discount      int            `json:"discount" example:"0"`
	Notes         string         `json:"notes" example:""`
}

// SalesReport represents the sales summary response
// @Description Sales summary report with revenue, transaction count, and best seller
type SalesReport struct {
	TotalRevenue       int                 `json:"total_revenue" example:"45000"`
	TotalTransactions  int                 `json:"total_transactions" example:"5"`
	BestSellingProduct *BestSellingProduct `json:"best_selling_product"`
}

// BestSellingProduct represents the best selling product in a report
// @Description Best selling product information
type BestSellingProduct struct {
	Name    string `json:"name" example:"Indomie Goreng"`
	QtySold int    `json:"qty_sold" example:"12"`
}

// DashboardStats represents the summary statistics for the dashboard
// @Description Dashboard summary statistics
type DashboardStats struct {
	TotalRevenueToday int                 `json:"total_revenue_today" example:"450000"`
	TransactionsToday int                 `json:"transactions_today" example:"10"`
	TotalProducts     int                 `json:"total_products" example:"50"`
	TotalCategories   int                 `json:"total_categories" example:"8"`
	LowStockCount     int                 `json:"low_stock_count" example:"3"`
	BestSellingToday  *BestSellingProduct `json:"best_selling_today"`
}

// TransactionListItem represents a transaction in the list view
// @Description Transaction summary for list display
type TransactionListItem struct {
	ID            int       `json:"id" example:"1"`
	TotalAmount   int       `json:"total_amount" example:"45000"`
	PaymentMethod string    `json:"payment_method" example:"cash"`
	Discount      int       `json:"discount" example:"0"`
	Status        string    `json:"status" example:"active"`
	ItemCount     int       `json:"item_count" example:"3"`
	CreatedAt     time.Time `json:"created_at" example:"2026-02-08T12:00:00Z"`
}

// PaginatedTransactions represents a paginated list of transactions
// @Description Paginated list of transactions
type PaginatedTransactions struct {
	Data       []TransactionListItem `json:"data"`
	Total      int                   `json:"total" example:"100"`
	Page       int                   `json:"page" example:"1"`
	Limit      int                   `json:"limit" example:"10"`
	TotalPages int                   `json:"total_pages" example:"10"`
}

// CategoryRevenue represents revenue breakdown per category
// @Description Revenue breakdown per category
type CategoryRevenue struct {
	CategoryID   int    `json:"category_id" example:"1"`
	CategoryName string `json:"category_name" example:"Electronics"`
	Revenue      int    `json:"revenue" example:"5000000"`
	Transactions int    `json:"transactions" example:"25"`
}

// ReportSummary represents the aggregated report summary
// @Description Aggregated report summary with category breakdown
type ReportSummary struct {
	TotalRevenue       int                `json:"total_revenue" example:"15000000"`
	TotalTransactions  int                `json:"total_transactions" example:"100"`
	BestSellingProduct *BestSellingProduct `json:"best_selling_product"`
	CategoryBreakdown  []CategoryRevenue  `json:"category_breakdown"`
}
