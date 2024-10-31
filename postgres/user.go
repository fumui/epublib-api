package postgres

import (
	"context"
	"database/sql"
	epublib "epublib"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type User struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Address     string       `json:"address"`
	PhoneNumber string       `json:"phone_number"`
	Gender      string       `json:"gender"`
	BirthDate   sql.NullTime `json:"birth_date"`
	ImgProfile  string       `json:"img_profile"`
	CreatedAt   sql.NullTime `json:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at"`
}

func (user *User) toEpublibUser() *epublib.User {
	return &epublib.User{
		ID:          user.ID,
		Name:        user.Name,
		Address:     user.Address,
		PhoneNumber: user.PhoneNumber,
		Gender:      epublib.Gender(user.Gender),
		BirthDate:   user.BirthDate.Time,
		ImgProfile:  user.ImgProfile,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
		DeletedAt:   user.DeletedAt.Time,
	}
}

// UserService represents a service for managing users.
type UserService struct {
	db epublib.Conn
}

// NewUserService returns a new instance of UserService attached to DB.
func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

func (svc *UserService) FindUserByID(ctx context.Context, id string) (*epublib.User, error) {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	user := &User{}
	err := db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", id).Scan(
		&user.ID,
		&user.Name,
		&user.Address,
		&user.PhoneNumber,
		&user.Gender,
		&user.BirthDate,
		&user.ImgProfile,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return user.toEpublibUser(), nil
}

func (svc *UserService) FindUserByUsername(ctx context.Context, username string) (*epublib.User, error) {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	user := &User{}
	err := db.QueryRow(ctx, "SELECT * FROM users WHERE username = $1", username).Scan(
		&user.ID,
		&user.Name,
		&user.Address,
		&user.PhoneNumber,
		&user.Gender,
		&user.BirthDate,
		&user.ImgProfile,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, epublib.ErrNotFound
		}
		log.Println(err)
		return nil, err
	}
	return user.toEpublibUser(), nil
}

func (svc *UserService) FindUsers(ctx context.Context, filter epublib.UserFilter) ([]*epublib.User, int, error) {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	var users []*epublib.User
	var totalCount int

	// Build the SQL query based on the filter criteria.
	query := "SELECT * FROM users WHERE true"
	filterQuery := ""
	args := []interface{}{}

	if filter.Name != "" {
		args = append(args, "%"+filter.Name+"%")
		filterQuery += fmt.Sprintf(" AND name LIKE $%d", len(args))
	}
	if !filter.IncludeDeleted {
		filterQuery += " AND deleted_at IS NULL"
	}

	// Count the total number of matching users.
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE true"+filterQuery, args...).Scan(&totalCount)
	if err != nil {
		log.Println(err)
		if err == pgx.ErrNoRows {
			return nil, 0, epublib.ErrNotFound
		}
		return nil, 0, err
	}

	// Apply filter & pagination using OFFSET and LIMIT.
	query += filterQuery + fmt.Sprintf(" OFFSET %d LIMIT %d", filter.Offset, filter.Limit)

	rows, err := svc.db.Query(ctx, query, args...)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Address,
			&user.PhoneNumber,
			&user.Gender,
			&user.BirthDate,
			&user.ImgProfile,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			log.Println(err)
			return nil, 0, err
		}
		users = append(users, user.toEpublibUser())
	}

	return users, totalCount, nil
}

func (svc *UserService) CreateUser(ctx context.Context, user *epublib.User) error {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	err := db.QueryRow(
		ctx,
		"INSERT INTO users (name, address, phone_number, gender, birth_date, img_profile) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		user.Name,
		user.Address,
		user.PhoneNumber,
		user.Gender,
		user.BirthDate,
		user.ImgProfile,
	).Scan(&user.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return epublib.ErrNotFound
		}
		log.Println(err)
		return err
	}
	return nil
}

func (svc *UserService) UpdateUser(ctx context.Context, id string, upd epublib.UserUpdate) (*epublib.User, error) {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	_, err := db.Exec(
		ctx,
		"UPDATE users SET name=$1, address=$2, phone_number=$3, gender=$4, birth_date=$5, img_profile=$6, updated_at = current_timestamp WHERE id=$7",
		upd.Name,
		upd.Address,
		upd.PhoneNumber,
		upd.Gender,
		upd.BirthDate,
		upd.ImgProfile,
		id,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Retrieve the updated user for response.
	return svc.FindUserByID(ctx, id)
}

func (svc *UserService) DeleteUser(ctx context.Context, id string) error {
	db := svc.db
	if tx := epublib.TxFromContext(ctx); tx != nil {
		db = tx
	}
	_, err := db.Exec(ctx, "UPDATE users SET deleted_at=current_timestamp WHERE id=$1", id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
