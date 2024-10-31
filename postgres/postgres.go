package postgres

import (
	"context"
	epublib "epublib"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDb() (*pgxpool.Pool, error) {

	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println(err)
	// }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	port, err := strconv.Atoi(dbPort)
	if err != nil {
		panic(err)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbUser, dbPass, dbHost, port, dbName)
	dbConnPool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		panic(err)
	}
	return dbConnPool, nil
}

func BeginTx(ctx context.Context, db *pgxpool.Pool) (context.Context, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	ctx = epublib.NewContextWithTx(ctx, tx)
	return ctx, nil
}

func Rollback(ctx context.Context) error {
	conn := epublib.TxFromContext(ctx)
	if tx, ok := conn.(pgx.Tx); ok {
		return tx.Rollback(ctx)
	}
	return nil
}

func Commit(ctx context.Context) error {
	conn := epublib.TxFromContext(ctx)
	if tx, ok := conn.(pgx.Tx); ok {
		return tx.Commit(ctx)
	}
	return nil
}
