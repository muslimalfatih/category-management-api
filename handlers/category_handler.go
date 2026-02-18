package handlers

import (
	"database/sql"
	"retail-core-api/helpers"
	"retail-core-api/models"
	"retail-core-api/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	service        services.CategoryService
	productService services.ProductService
}

// NewCategoryHandler creates a new category handler instance
func NewCategoryHandler(service services.CategoryService, productService services.ProductService) *CategoryHandler {
	return &CategoryHandler{service: service, productService: productService}
}

// List godoc
// @Summary Get all categories
// @Description Retrieve a list of all categories
// @Tags Categories
// @Produce json
// @Success 200 {object} helpers.Response{data=[]models.Category} "Successfully retrieved all categories"
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve categories", err.Error())
		return
	}
	helpers.OK(c, "Successfully retrieved all categories", categories)
}

// GetByID godoc
// @Summary Get a category by ID
// @Description Retrieve details of a specific category by its ID
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} helpers.Response{data=models.Category} "Category retrieved successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid category ID"
// @Failure 404 {object} helpers.ErrorResponse "Category not found"
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid category ID")
		return
	}

	category, err := h.service.GetCategoryByID(id)
	if err != nil {
		helpers.InternalError(c, "Failed to retrieve category", err.Error())
		return
	}
	if category == nil {
		helpers.NotFound(c, "Category not found")
		return
	}
	helpers.OK(c, "Category retrieved successfully", category)
}

// Create godoc
// @Summary Create a new category
// @Description Add a new category to the database
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body models.CategoryInput true "Category object that needs to be added"
// @Success 201 {object} helpers.Response{data=models.Category} "Category created successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request body or validation error"
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
	}

	created, err := h.service.CreateCategory(category)
	if err != nil {
		helpers.BadRequest(c, err.Error())
		return
	}
	helpers.Created(c, "Category created successfully", created)
}

// Update godoc
// @Summary Update a category
// @Description Update an existing category by its ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body models.CategoryInput true "Updated category object"
// @Success 200 {object} helpers.Response{data=models.Category} "Category updated successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid request body or validation error"
// @Failure 404 {object} helpers.ErrorResponse "Category not found"
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid category ID")
		return
	}

	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
	}

	updated, err := h.service.UpdateCategory(id, category)
	if err != nil {
		if helpers.IsNotFound(err) || err.Error() == "category not found" {
			helpers.NotFound(c, "Category not found")
		} else {
			helpers.BadRequest(c, err.Error())
		}
		return
	}
	helpers.OK(c, "Category updated successfully", updated)
}

// Delete godoc
// @Summary Delete a category
// @Description Delete a category by its ID
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} helpers.Response "Category deleted successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid category ID"
// @Failure 404 {object} helpers.ErrorResponse "Category not found"
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid category ID")
		return
	}

	err = h.service.DeleteCategory(id)
	if err != nil {
		if err == sql.ErrNoRows {
			helpers.NotFound(c, "Category not found")
			return
		}
		helpers.InternalError(c, "Failed to delete category", err.Error())
		return
	}
	helpers.OK(c, "Category deleted successfully", nil)
}

// GetProducts godoc
// @Summary Get products by category
// @Description Retrieve all products belonging to a specific category
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} helpers.Response{data=[]models.Product} "Products retrieved successfully"
// @Failure 400 {object} helpers.ErrorResponse "Invalid category ID"
// @Router /categories/{id}/products [get]
func (h *CategoryHandler) GetProducts(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		helpers.BadRequest(c, "Invalid category ID")
		return
	}

	products, err := h.productService.GetProductsByCategoryID(id)
	if err != nil {
		helpers.InternalError(c, "Failed to get products", err.Error())
		return
	}
	helpers.OK(c, "Products retrieved successfully", products)
}
