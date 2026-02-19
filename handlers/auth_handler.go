package handlers

import (
	"retail-core-api/helpers"
	"retail-core-api/models"
	"retail-core-api/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related endpoints
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.LoginInput true "Login credentials"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Failure 401 {object} helpers.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if input.Email == "" || input.Password == "" {
		helpers.BadRequest(c, "Email and password are required")
		return
	}

	result, err := h.authService.Login(input.Email, input.Password)
	if err != nil {
		helpers.Unauthorized(c, err.Error())
		return
	}

	helpers.OK(c, "Login successful", result)
}

// Register godoc
// @Summary Register new user
// @Description Create a new user account (owner-only)
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.UserInput true "User registration data"
// @Success 201 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Failure 409 {object} helpers.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if input.Name == "" || input.Email == "" || input.Password == "" {
		helpers.BadRequest(c, "Name, email, and password are required")
		return
	}

	role := input.Role
	if role == "" {
		role = "cashier"
	}

	user, err := h.authService.Register(input.Name, input.Email, input.Password, role)
	if err != nil {
		if err.Error() == "email already registered" {
			helpers.Error(c, 409, err.Error())
			return
		}
		helpers.BadRequest(c, err.Error())
		return
	}

	helpers.Created(c, "User registered successfully", user)
}
