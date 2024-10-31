package main

import (
	"epublib"
	embedServer "epublib/internal/embed"
	httpAPI "epublib/internal/http"
	"epublib/mailer"
	"epublib/postgres"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	mux := mux.NewRouter()
	db, err := postgres.InitDb()
	if err != nil {
		panic(err)
	}
	api := httpAPI.NewAPI(mux, db)
	api.Register()
	api.AuthService = postgres.NewAuthService(db)
	api.UserService = postgres.NewUserService(db)
	api.ResetTokenService = postgres.NewResetTokenService(db)
	api.MailerService = mailer.NewMailerService()
	embedServer.RegisterSwaggerUI(epublib.SwaggerUI, "docs/swaggerui", mux)

	if os.Getenv("TLS") == "true" {
		api.UseTLS = true
		log.Println("Server started at :443")
		err := http.ListenAndServeTLS(":443", os.Getenv("SSL_CRT"), os.Getenv("SSL_KEY"), api.Router)
		if err != nil {
			panic(err)
		}
	} else {
		log.Println("Server started at :80")
		err := http.ListenAndServe(":80", api.Router)
		if err != nil {
			panic(err)
		}
	}
}
