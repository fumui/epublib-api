package epublib

import (
	"context"
)

// contextKey represents an internal key for adding context fields.
// This is considered best practice as it prevents other packages from
// interfering with our context keys.
type contextKey int

// List of context keys.
// These are used to store request-scoped information.
const (
	// Stores the current logged in user in the context.
	userContextKey = contextKey(iota + 1)
	// Stores database tx instance if any
	txContextKey = contextKey(iota + 1)
)

// NewContextWithUser returns a new context with the given user.
func NewContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext returns the current logged in user.
func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userContextKey).(*User)
	return user
}

// UserIDFromContext is a helper function that returns the ID of the current
// logged in user. Returns empty if no user is logged in.
func UserIDFromContext(ctx context.Context) string {
	if user := UserFromContext(ctx); user != nil {
		return user.ID
	}
	return ""
}

// NewContextWithTx returns a new context with the given tx.
func NewContextWithTx(ctx context.Context, tx Conn) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}

// TxFromContext returns the current logged in user.
func TxFromContext(ctx context.Context) Conn {
	tx, _ := ctx.Value(txContextKey).(Conn)
	return tx
}
