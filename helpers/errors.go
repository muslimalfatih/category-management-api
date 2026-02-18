package helpers

import "errors"

// Sentinel errors for common application error conditions.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

// AppError wraps an error with an application-specific message so callers can
// provide user-facing context while preserving the underlying sentinel for
// programmatic checks.
type AppError struct {
	Err     error
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates an AppError wrapping ErrNotFound.
func NewNotFoundError(message string) *AppError {
	return &AppError{Err: ErrNotFound, Message: message}
}

// NewValidationError creates an AppError wrapping ErrValidation.
func NewValidationError(message string) *AppError {
	return &AppError{Err: ErrValidation, Message: message}
}

// IsNotFound reports whether err (or any error in its chain) is ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsValidation reports whether err (or any error in its chain) is ErrValidation.
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}
