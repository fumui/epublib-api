package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	epublib "epublib"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Auth struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	Username  string       `json:"username"`
	Password  string       `json:"password"`
	Email     string       `json:"email"`
	Level     string       `json:"level"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

func (auth *Auth) toEpublibAuth() *epublib.Auth {
	return &epublib.Auth{
		ID:        auth.ID,
		UserID:    auth.UserID,
		Username:  auth.Username,
		Password:  auth.Password,
		Email:     auth.Email,
		Level:     epublib.AuthLevel(auth.Level),
		CreatedAt: auth.CreatedAt.Time,
		UpdatedAt: auth.UpdatedAt.Time,
		DeletedAt: auth.DeletedAt.Time,
	}
}

// AuthService represents a service for managing OAuth authentication.
type AuthService struct {
	db epublib.Conn
}

// NewAuthService returns a new instance of AuthService attached to DB.
func NewAuthService(db *pgxpool.Pool) *AuthService {
	return &AuthService{db: db}
}

func (svc *AuthService) FindAuthByEmail(ctx context.Context, email string) (*epublib.Auth, error) {
	auth := &Auth{}
	err := svc.db.QueryRow(ctx, "SELECT * FROM auth WHERE email = $1 AND deleted_at IS NULL", email).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.Username,
		&auth.Password,
		&auth.Email,
		&auth.Level,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&auth.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return auth.toEpublibAuth(), nil
}

func (svc *AuthService) FindAuthByEmailPass(ctx context.Context, email, password string) (*epublib.Auth, error) {
	auth := &Auth{}
	encrypted := encryptPass(password, email)
	err := svc.db.QueryRow(ctx, "SELECT * FROM auth WHERE email = $1 AND password = $2", email, encrypted).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.Username,
		&auth.Password,
		&auth.Email,
		&auth.Level,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&auth.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return auth.toEpublibAuth(), nil
}

// Looks up an authentication object by ID.
// Returns ENOTFOUND if ID does not exist.
func (svc *AuthService) FindAuthByID(ctx context.Context, id string) (*epublib.Auth, error) {
	auth := &Auth{}
	err := svc.db.QueryRow(ctx, "SELECT * FROM auth WHERE id = $1", id).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.Username,
		&auth.Password,
		&auth.Email,
		&auth.Level,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&auth.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return auth.toEpublibAuth(), nil
}

// Looks up an authentication object by UserID.
// Returns ENOTFOUND if ID does not exist.
func (svc *AuthService) FindAuthByUserID(ctx context.Context, id string) (*epublib.Auth, error) {
	auth := &Auth{}
	err := svc.db.QueryRow(ctx, "SELECT * FROM auth WHERE user_id = $1", id).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.Username,
		&auth.Password,
		&auth.Email,
		&auth.Level,
		&auth.CreatedAt,
		&auth.UpdatedAt,
		&auth.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return auth.toEpublibAuth(), nil
}

// Creates a new authentication object
// On success, the auth.ID is set to the new authentication ID.
func (svc *AuthService) CreateAuth(ctx context.Context, auth *epublib.Auth) error {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	encrypted := encryptPass(auth.Password, auth.Email)
	err := db.QueryRow(
		ctx,
		"INSERT INTO auth (user_id, username, password, email, level, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		auth.UserID,
		auth.Username,
		encrypted,
		auth.Email,
		auth.Level,
		auth.CreatedAt,
		auth.UpdatedAt,
	).Scan(&auth.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return epublib.ErrNotFound
		}
		log.Println(err)
		return err
	}
	return nil
}

func (svc *AuthService) ResetAuthPassword(ctx context.Context, id, email, password string) error {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	encrypted := encryptPass(password, email)
	_, err := db.Exec(
		ctx,
		"UPDATE auth SET password = $1, updated_at = current_timestamp WHERE id = $2",
		encrypted,
		id,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Soft deletes an authentication object from the system by ID.
// The parent user object is not removed.
func (svc *AuthService) DeleteAuth(ctx context.Context, id string) error {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	_, err := db.Exec(ctx, "UPDATE auth SET deleted_at = current_timestamp WHERE id = $1", id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func encryptPass(pass string, salt string) string {
	h := sha256.New()
	h.Write([]byte(pass + "|" + salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}
