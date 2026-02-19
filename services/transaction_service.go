package services

import (
	"errors"
	"retail-core-api/models"
	"retail-core-api/repositories"
)

// TransactionService defines the interface for transaction business logic
type TransactionService interface {
	Checkout(req models.CheckoutRequest) (*models.Transaction, error)
	GetAllTransactions(page, limit int, startDate, endDate string) (*models.PaginatedTransactions, error)
	GetTransactionByID(id int) (*models.Transaction, error)
	VoidTransaction(id int) error
	GetDashboardStats() (*models.DashboardStats, error)
	GetDailySalesReport() (*models.SalesReport, error)
	GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error)
	GetReportSummary(startDate, endDate string) (*models.ReportSummary, error)
}

// transactionService implements TransactionService interface
type transactionService struct {
	repo repositories.TransactionRepository
}

// NewTransactionService creates a new transaction service instance
func NewTransactionService(repo repositories.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

// Checkout validates the checkout request and delegates to the repository
func (s *transactionService) Checkout(req models.CheckoutRequest) (*models.Transaction, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("checkout items cannot be empty")
	}

	for _, item := range req.Items {
		if item.ProductID <= 0 {
			return nil, errors.New("invalid product ID")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("quantity must be greater than 0")
		}
	}

	return s.repo.CreateTransaction(req)
}

// VoidTransaction voids a transaction and restores stock
func (s *transactionService) VoidTransaction(id int) error {
	if id <= 0 {
		return errors.New("invalid transaction ID")
	}
	return s.repo.VoidTransaction(id)
}

// GetDailySalesReport returns the sales summary for today
func (s *transactionService) GetDailySalesReport() (*models.SalesReport, error) {
	return s.repo.GetDailySalesReport()
}

// GetSalesReportByDateRange returns the sales summary for a given date range
func (s *transactionService) GetSalesReportByDateRange(startDate, endDate string) (*models.SalesReport, error) {
	if startDate == "" || endDate == "" {
		return nil, errors.New("start_date and end_date are required")
	}
	return s.repo.GetSalesReportByDateRange(startDate, endDate)
}

// GetReportSummary returns an aggregated report with category breakdown
func (s *transactionService) GetReportSummary(startDate, endDate string) (*models.ReportSummary, error) {
	if startDate == "" || endDate == "" {
		return nil, errors.New("start_date and end_date are required")
	}
	return s.repo.GetReportSummary(startDate, endDate)
}

// GetAllTransactions returns a paginated list of transactions with optional date range
func (s *transactionService) GetAllTransactions(page, limit int, startDate, endDate string) (*models.PaginatedTransactions, error) {
	return s.repo.GetAllTransactions(page, limit, startDate, endDate)
}

// GetTransactionByID returns a single transaction with its details
func (s *transactionService) GetTransactionByID(id int) (*models.Transaction, error) {
	if id <= 0 {
		return nil, errors.New("invalid transaction ID")
	}
	return s.repo.GetTransactionByID(id)
}

// GetDashboardStats returns summary statistics for the admin dashboard
func (s *transactionService) GetDashboardStats() (*models.DashboardStats, error) {
	return s.repo.GetDashboardStats()
}
