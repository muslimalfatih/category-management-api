package services

import (
	"errors"
	"retail-core-api/models"
	"retail-core-api/repositories"

	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user business logic
type UserService interface {
	GetAll() ([]models.User, error)
	GetByID(id int) (*models.User, error)
	Update(id int, input models.UserInput) (*models.User, error)
	Delete(id int) error
}

// userService implements UserService interface
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// GetAll returns all users
func (s *userService) GetAll() ([]models.User, error) {
	return s.userRepo.GetAll()
}

// GetByID returns a user by ID
func (s *userService) GetByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	// Clear password
	user.Password = ""
	return user, nil
}

// Update updates a user
func (s *userService) Update(id int, input models.UserInput) (*models.User, error) {
	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("user not found")
	}

	// If password is provided, hash it
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		input.Password = string(hash)
	}

	// Validate role if provided
	if input.Role != "" && input.Role != "owner" && input.Role != "cashier" {
		return nil, errors.New("role must be 'owner' or 'cashier'")
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		Role:     input.Role,
	}

	return s.userRepo.Update(id, user)
}

// Delete soft-deletes a user
func (s *userService) Delete(id int) error {
	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("user not found")
	}
	return s.userRepo.Delete(id)
}
