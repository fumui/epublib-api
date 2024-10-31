package epublib

import (
	"context"
	"time"
)

type AuthLevel string

const (
	AdminLevel AuthLevel = "Admin"
	UserLevel  AuthLevel = "User"
)

// IsValid checks if an AuthLevel is valid
func (a AuthLevel) IsValid() bool {
	switch a {
	case AdminLevel, UserLevel:
		return true
	default:
		return false
	}
}

// String returns the string representation of the AuthLevel
func (a AuthLevel) String() string {
	return string(a)
}

// Values returns all possible AuthLevel values
func Values() []AuthLevel {
	return []AuthLevel{AdminLevel, UserLevel}
}

type Auth struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	Level     AuthLevel `json:"level"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

// AuthService represents a service for managing auths.
type AuthService interface {
	// Looks up an authentication object by username and password.
	// Returns ENOTFOUND if no matching username and password does not exist.
	FindAuthByEmailPass(ctx context.Context, username, password string) (*Auth, error)

	// Looks up an authentication object by ID.
	// Returns ENOTFOUND if ID does not exist.
	FindAuthByID(ctx context.Context, id string) (*Auth, error)

	// Looks up an authentication object by UserID.
	// Returns ENOTFOUND if ID does not exist.
	FindAuthByUserID(ctx context.Context, id string) (*Auth, error)

	// Creates a new authentication object
	// On success, the auth.ID is set to the new authentication ID.
	CreateAuth(ctx context.Context, auth *Auth) error

	// Permanently deletes an authentication object from the system by ID.
	// The parent user object is not removed.
	DeleteAuth(ctx context.Context, id string) error

	// FindAuthByEmail looks up an authentication object by email.
	FindAuthByEmail(ctx context.Context, email string) (*Auth, error)

	// Resets Auth password
	ResetAuthPassword(ctx context.Context, id, email, password string) error
}
