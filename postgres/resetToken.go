package postgres

import (
	"context"
	"database/sql"
	epublib "epublib"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ResetToken struct {
	ID        string       `json:"id"`
	Token     string       `json:"token"`
	Used      bool         `json:"used"`
	CreatedAt sql.NullTime `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

func (r *ResetToken) toEpublibResetToken() *epublib.ResetToken {
	return &epublib.ResetToken{
		ID:        r.ID,
		Token:     r.Token,
		Used:      r.Used,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
		DeletedAt: r.DeletedAt.Time,
	}
}

// ResetTokenService represents a service for managing reset tokens.
type ResetTokenService struct {
	db epublib.Conn
}

// NewResetTokenService returns a new instance of ResetTokenService attached to DB.
func NewResetTokenService(db *pgxpool.Pool) *ResetTokenService {
	return &ResetTokenService{db: db}
}

// Generates new Reset Token based on auth.
func (r *ResetTokenService) GenerateResetToken(ctx context.Context, auth *epublib.Auth) (*epublib.ResetToken, error) {
	var err error
	token := &ResetToken{}
	token.Token, err = createJWT(auth, 24*time.Hour)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db := r.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	err = db.QueryRow(
		ctx,
		"INSERT INTO reset_token (token) VALUES ($1) RETURNING id",
		token.Token,
	).Scan(&token.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return token.toEpublibResetToken(), nil
}

// Validates Reset Token, checks if it was registered, unexpired, and has the correct format.
func (r *ResetTokenService) ValidateResetToken(ctx context.Context, id, token string) (bool, error) {
	db := r.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	used := false
	err := db.QueryRow(
		ctx,
		"SELECT used FROM reset_token WHERE id = $1 AND token = $2",
		id,
		token,
	).Scan(&used)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, epublib.ErrNotFound
		}
		log.Println(err)
		return false, err
	}
	return !used, nil
}

// Marks Reset Token as used and returns decoded auth in token.
func (r *ResetTokenService) UseResetToken(ctx context.Context, id string, token string) (*epublib.Auth, error) {
	db := r.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	_, err := db.Exec(
		ctx,
		"UPDATE reset_token SET used = true WHERE id = $1",
		id,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	decoded, err := decodeJWT(token)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	claims, ok := decoded.Claims.(*ResetTokenClaims)
	if !ok {
		err := fmt.Errorf("failed to parse claimed token: %v", decoded)
		log.Println(err)
		return nil, err
	}
	return &claims.Auth, nil
}

type ResetTokenClaims struct {
	jwt.RegisteredClaims
	Auth epublib.Auth `json:"auth"`
}

func (c ResetTokenClaims) Valid() error {
	return c.RegisteredClaims.Valid()
}
func getJWTKey(token *jwt.Token) (interface{}, error) {
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		return nil, fmt.Errorf("JWT_KEY is empty")
	}
	return []byte(jwtKey), nil
}
func createJWT(auth *epublib.Auth, duration time.Duration) (string, error) {
	jwtKey, err := getJWTKey(nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, ResetTokenClaims{
		Auth: *auth,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"epublib-erp"},
			Subject:   auth.Email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	})
	tokenString, err := token.SignedString(jwtKey.([]byte))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return tokenString, nil
}
func decodeJWT(token string) (*jwt.Token, error) {
	parsed, err := jwt.ParseWithClaims(token, &ResetTokenClaims{}, getJWTKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return parsed, nil
}
