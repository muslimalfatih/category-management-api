package handlers

import (
	"retail-core-api/helpers"
	"retail-core-api/models"
	"retail-core-api/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service services.ProductService
}

// NewProductHandler creates a new product handler instance
func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// List godoc
// @Summary Get all products (paginated)
// @Description Retrieve a paginated list of products. Supports search by name and filter by category_id.
// @Tags Products
// @Produce json
// @Param search query string false "Search product by name (case-insensitive partial match)"
// @Param category_id query int false "Filter by category ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} helpers.PaginatedResponse
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	params := models.ProductListParams{
		Search: c.Query("search"),
	}

	// Also support legacy "name" query param
	if params.Search == "" {
		params.Search = c.Query("name")
	}

	if catID := c.Query("category_id"); catID != "" {
		if id, err := strconv.Atoi(catID); err == nil {
			params.CategoryID = &id
		}
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			params.Limit = l
		}
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	result, err := h.service.GetAllProducts(params)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve products", err.Error())
		return
	}

	helpers.Paginated(c, "Successfully retrieved products", result.Data, helpers.PaginationMeta{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// GetByID godoc
// @Summary Get a product by ID
// @Description Retrieve details of a specific product by its ID with category name
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} helpers.Response{data=models.Product} "Product retrieved successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid product ID"
// @Failure 404 {object} helpers.ErrorResponse "Product not found"
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid product ID")
		return
	}

	product, err := h.service.GetProductByID(id)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve product", err.Error())
		return
	}
	if product == nil {
		helpers.NotFound(c, "Product not found")
		return
	}
	helpers.OK(c, "Product retrieved successfully", product)
}

// Create godoc
// @Summary Create a new product
// @Description Add a new product to the database
// @Tags Products
// @Accept json
// @Produce json
// @Param product body models.ProductInput true "Product object that needs to be added"
// @Success 201 {object} helpers.Response{data=models.Product} "Product created successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request body or validation error"
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var input models.ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		SKU:        input.SKU,
		ImageURL:   input.ImageURL,
		Unit:       input.Unit,
		IsActive:   isActive,
		CategoryID: input.CategoryID,
	}

	created, err := h.service.CreateProduct(product)
	if err != nil {
		helpers.BadRequest(c, err.Error())
		return
	}
	helpers.Created(c, "Product created successfully", created)
}

// Update godoc
// @Summary Update a product
// @Description Update an existing product by its ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body models.ProductInput true "Updated product object"
// @Success 200 {object} helpers.Response{data=models.Product} "Product updated successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request body or validation error"
// @Failure 404 {object} helpers.ErrorResponse "Product not found"
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid product ID")
		return
	}

	var input models.ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
		SKU:        input.SKU,
		ImageURL:   input.ImageURL,
		Unit:       input.Unit,
		CategoryID: input.CategoryID,
	}

	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	} else {
		product.IsActive = true
	}

	updated, err := h.service.UpdateProduct(id, product)
	if err != nil {
		if helpers.IsNotFound(err) || err.Error() == "product not found" {
			helpers.NotFound(c, "Product not found")
		} else {
			helpers.BadRequest(c, err.Error())
		}
		return
	}
	helpers.OK(c, "Product updated successfully", updated)
}

// Delete godoc
// @Summary Delete a product
// @Description Delete a product by its ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} helpers.Response "Product deleted successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid product ID"
// @Failure 404 {object} helpers.ErrorResponse "Product not found"
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid product ID")
		return
	}

	err = h.service.DeleteProduct(id)
	if err != nil {
		if helpers.IsNotFound(err) || err.Error() == "product not found" {
			helpers.NotFound(c, "Product not found")
			return
		}
		helpers.InternalError(c, "Failed to delete product", err.Error())
		return
	}
	helpers.OK(c, "Product deleted successfully", nil)
}
