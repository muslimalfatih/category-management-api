package repositories

import (
	"database/sql"
	"retail-core-api/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll() ([]models.User, error)
	Create(user models.User) (*models.User, error)
	Update(id int, user models.User) (*models.User, error)
	Delete(id int) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID returns a user by their ID
func (r *userRepository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, name, email, password, role, is_active, created_at FROM users WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail returns a user by their email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password, role, is_active, created_at FROM users WHERE email = $1`
	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAll returns all users
func (r *userRepository) GetAll() ([]models.User, error) {
	query := `SELECT id, name, email, password, role, is_active, created_at FROM users ORDER BY id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.Password,
			&user.Role, &user.IsActive, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		user.Password = "" // never expose password
		users = append(users, user)
	}
	return users, rows.Err()
}

// Create adds a new user
func (r *userRepository) Create(user models.User) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, email, role, is_active, created_at
	`
	var created models.User
	err := r.db.QueryRow(query, user.Name, user.Email, user.Password, user.Role, true).Scan(
		&created.ID, &created.Name, &created.Email,
		&created.Role, &created.IsActive, &created.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// Update modifies an existing user
func (r *userRepository) Update(id int, user models.User) (*models.User, error) {
	query := `
		UPDATE users SET name = $1, email = $2, role = $3, is_active = $4
		WHERE id = $5
		RETURNING id, name, email, role, is_active, created_at
	`
	var updated models.User
	err := r.db.QueryRow(query, user.Name, user.Email, user.Role, user.IsActive, id).Scan(
		&updated.ID, &updated.Name, &updated.Email,
		&updated.Role, &updated.IsActive, &updated.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// Delete deactivates a user by ID
func (r *userRepository) Delete(id int) error {
	query := `UPDATE users SET is_active = false WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
