package services

import (
	"errors"
	"retail-core-api/models"
	"retail-core-api/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	Login(email, password string) (*models.LoginResponse, error)
	Register(name, email, password, role string) (*models.User, error)
}

// authService implements AuthService interface
type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

// NewAuthService creates a new auth service instance
func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(email, password string) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("failed to find user")
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"name":    user.Name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Clear password before returning
	user.Password = ""

	return &models.LoginResponse{
		Token: tokenString,
		User:  *user,
	}, nil
}

// Register creates a new user account
func (s *authService) Register(name, email, password, role string) (*models.User, error) {
	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("failed to check existing user")
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Validate role
	if role != "owner" && role != "cashier" {
		return nil, errors.New("role must be 'owner' or 'cashier'")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     role,
	}

	return s.userRepo.Create(user)
}
