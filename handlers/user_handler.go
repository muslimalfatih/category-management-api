package handlers

import (
	"retail-core-api/helpers"
	"retail-core-api/models"
	"retail-core-api/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler instance
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetAll godoc
// @Summary List all users
// @Description Get all users (owner only)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} helpers.Response
// @Router /api/users [get]
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll()
	if err != nil {
		helpers.InternalError(c, "Failed to fetch users", err.Error())
		return
	}
	helpers.OK(c, "Users retrieved successfully", users)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get a single user by ID (owner only)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} helpers.Response
// @Failure 404 {object} helpers.Response
// @Router /api/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		helpers.NotFound(c, err.Error())
		return
	}

	helpers.OK(c, "User retrieved successfully", user)
}

// Update godoc
// @Summary Update a user
// @Description Update user details (owner only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param body body models.UserInput true "User data"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Failure 404 {object} helpers.Response
// @Router /api/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "Invalid user ID")
		return
	}

	var input models.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	user, err := h.userService.Update(id, input)
	if err != nil {
		helpers.BadRequest(c, err.Error())
		return
	}

	helpers.OK(c, "User updated successfully", user)
}

// Delete godoc
// @Summary Delete a user
// @Description Soft delete a user (owner only)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} helpers.Response
// @Failure 404 {object} helpers.Response
// @Router /api/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helpers.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.Delete(id); err != nil {
		helpers.NotFound(c, err.Error())
		return
	}

	helpers.OK(c, "User deleted successfully", nil)
}
