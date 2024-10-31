package epublib

import (
	"context"
	"time"
)

type Gender string

const (
	UnidentifiedGender Gender = "U"
	MaleGender         Gender = "M"
	FemaleGender       Gender = "F"
)

type User struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phone_number"`
	Gender      Gender    `json:"gender"`
	BirthDate   time.Time `json:"birth_date"`
	ImgProfile  string    `json:"img_profile"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at"`
}

// UserService represents a service for managing users.
type UserService interface {
	// Retrieves a user by ID.
	FindUserByID(ctx context.Context, id string) (*User, error)

	// Retrieves a user by username.
	FindUserByUsername(ctx context.Context, username string) (*User, error)

	// Retrieves a list of users by filter. Also returns total count of matching
	// users which may differ from returned results if filter.Limit is specified.
	FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)

	// Creates a new user.
	CreateUser(ctx context.Context, user *User) error

	// Updates a user object.
	UpdateUser(ctx context.Context, id string, upd UserUpdate) (*User, error)

	// Soft deletes a user.
	DeleteUser(ctx context.Context, id string) error
}

// UserFilter represents a filter passed to FindUsers().
type UserFilter struct {
	// Filtering fields.
	Name           string `json:"name"`
	IncludeDeleted bool   `json:"include_deleted"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// UserUpdate represents a set of fields to be updated via UpdateUser().
type UserUpdate struct {
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	PhoneNumber string    `json:"phone_number"`
	Gender      Gender    `json:"gender"`
	BirthDate   time.Time `json:"birth_date"`
	ImgProfile  string    `json:"img_profile"`
}
