package epublib

import (
	"context"
	"time"
)

type ResetToken struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

// AuthService represents a service for managing auths.
type ResetTokenService interface {

	// Generates new Reset Token based on auth.
	GenerateResetToken(ctx context.Context, auth *Auth) (*ResetToken, error)

	// Validates Reset Token, checks if it was registered, unexpired, and has the correct format.
	ValidateResetToken(ctx context.Context, id, token string) (bool, error)

	// Marks Reset Token as used and returns decoded auth in token.
	UseResetToken(ctx context.Context, id string, token string) (*Auth, error)
}
