package handlers

import (
	"retail-core-api/helpers"
	"retail-core-api/models"
	"retail-core-api/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TransactionHandler handles HTTP requests for transactions and reports
type TransactionHandler struct {
	service services.TransactionService
}

// NewTransactionHandler creates a new transaction handler instance
func NewTransactionHandler(service services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// Checkout godoc
// @Summary Process checkout
// @Description Process a checkout with items, payment method, optional discount and notes
// @Tags Transactions
// @Accept json
// @Produce json
// @Param request body models.CheckoutRequest true "Checkout request"
// @Success 201 {object} helpers.Response{data=models.Transaction} "Checkout successful"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} helpers.ErrorResponse "Server error or insufficient stock"
// @Router /api/checkout [post]
func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	transaction, err := h.service.Checkout(req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "insufficient stock") || strings.Contains(errMsg, "cannot be empty") || strings.Contains(errMsg, "invalid") {
			helpers.BadRequest(c, errMsg)
			return
		}
		helpers.InternalError(c, errMsg)
		return
	}
	helpers.Created(c, "Checkout successful", transaction)
}

// ListTransactions godoc
// @Summary Get all transactions
// @Description Retrieve a paginated list of all transactions with optional date range filter
// @Tags Transactions
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} helpers.Response{data=models.PaginatedTransactions} "Successfully retrieved transactions"
// @Router /api/transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	page, limit := helpers.ParsePagination(c)
	startDate := strings.TrimSpace(c.Query("start_date"))
	endDate := strings.TrimSpace(c.Query("end_date"))

	result, err := h.service.GetAllTransactions(page, limit, startDate, endDate)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve transactions", err.Error())
		return
	}
	helpers.Paginated(c, "Successfully retrieved transactions", result.Data, helpers.PaginationMeta{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// GetTransactionByID godoc
// @Summary Get a transaction by ID
// @Description Retrieve details of a specific transaction including its items
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} helpers.Response{data=models.Transaction} "Transaction retrieved successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid transaction ID"
// @Failure 404 {object} helpers.ErrorResponse "Transaction not found"
// @Router /api/transactions/{id} [get]
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid transaction ID")
		return
	}

	transaction, err := h.service.GetTransactionByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			helpers.NotFound(c, err.Error())
			return
		}
		helpers.InternalError(c, "Failed to retrieve transaction", err.Error())
		return
	}
	helpers.OK(c, "Transaction retrieved successfully", transaction)
}

// VoidTransaction godoc
// @Summary Void a transaction
// @Description Void a transaction and restore product stock
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} helpers.Response "Transaction voided successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid transaction ID or already voided"
// @Failure 500 {object} helpers.ErrorResponse "Server error"
// @Router /api/transactions/{id}/void [patch]
func (h *TransactionHandler) VoidTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid transaction ID")
		return
	}

	err = h.service.VoidTransaction(id)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "already voided") {
			helpers.BadRequest(c, errMsg)
			return
		}
		helpers.InternalError(c, errMsg)
		return
	}
	helpers.OK(c, "Transaction voided successfully", nil)
}

// DailyReport godoc
// @Summary Get today's sales report
// @Description Retrieve the sales summary for today including revenue, transaction count, and best seller
// @Tags Reports
// @Produce json
// @Success 200 {object} helpers.Response{data=models.SalesReport} "Successfully retrieved today's report"
// @Router /api/report/today [get]
func (h *TransactionHandler) DailyReport(c *gin.Context) {
	report, err := h.service.GetDailySalesReport()
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve daily report", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved today's report", report)
}

// ReportByRange godoc
// @Summary Get sales report by date range
// @Description Retrieve the sales summary for a specific date range
// @Tags Reports
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} helpers.Response{data=models.SalesReport} "Successfully retrieved report"
// @Failure 400 {object} helpers.ErrorResponse "Missing start_date or end_date"
// @Router /api/report [get]
func (h *TransactionHandler) ReportByRange(c *gin.Context) {
	startDate := strings.TrimSpace(c.Query("start_date"))
	endDate := strings.TrimSpace(c.Query("end_date"))

	if startDate == "" || endDate == "" {
		helpers.BadRequest(c, "start_date and end_date are required")
		return
	}

	report, err := h.service.GetSalesReportByDateRange(startDate, endDate)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve report", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved report", report)
}

// ReportSummary godoc
// @Summary Get aggregated report summary
// @Description Retrieve aggregated report summary with category breakdown for a date range
// @Tags Reports
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} helpers.Response{data=models.ReportSummary} "Successfully retrieved report summary"
// @Failure 400 {object} helpers.ErrorResponse "Missing start_date or end_date"
// @Router /api/report/summary [get]
func (h *TransactionHandler) ReportSummary(c *gin.Context) {
	startDate := strings.TrimSpace(c.Query("start_date"))
	endDate := strings.TrimSpace(c.Query("end_date"))

	if startDate == "" || endDate == "" {
		helpers.BadRequest(c, "start_date and end_date are required")
		return
	}

	summary, err := h.service.GetReportSummary(startDate, endDate)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve report summary", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved report summary", summary)
}

// Dashboard godoc
// @Summary Get dashboard statistics
// @Description Retrieve summary statistics for the POS dashboard
// @Tags Dashboard
// @Produce json
// @Success 200 {object} helpers.Response{data=models.DashboardStats} "Successfully retrieved dashboard data"
// @Router /api/dashboard [get]
func (h *TransactionHandler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats()
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve dashboard data", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved dashboard data", stats)
}
