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
// @Summary Get all products
// @Description Retrieve a list of all products with their category names. Supports optional search by name.
// @Tags Products
// @Produce json
// @Param name query string false "Search product by name (case-insensitive partial match)"
// @Success 200 {object} helpers.Response{data=[]models.Product} "Successfully retrieved all products"
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	name := c.Query("name")

	products, err := h.service.GetAllProducts(name)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve products", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved all products", products)
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

	product := models.Product{
		Name:       input.Name,
		Price:      input.Price,
		Stock:      input.Stock,
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
		CategoryID: input.CategoryID,
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
