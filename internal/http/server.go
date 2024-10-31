package http

import (
	epublib "epublib"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

type API struct {
	Router *mux.Router
	db     *pgxpool.Pool
	UseTLS bool

	AuthService       epublib.AuthService
	ResetTokenService epublib.ResetTokenService
	UserService       epublib.UserService
	MailerService     epublib.MailerService
}

func NewAPI(router *mux.Router, db *pgxpool.Pool) *API {
	api := &API{
		Router: router,
		db:     db,
	}
	return api
}
