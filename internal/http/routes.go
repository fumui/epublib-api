package http

import (
	"encoding/json"
	epublib "epublib"
	"net/http"
	"regexp"
)

func (api *API) Register() {
	api.Router.Use(api.handleCors)
	router := api.Router.PathPrefix("/api/v1").Subrouter()
	router.HandleFunc("/swagger-spec", byteHandler(epublib.SwaggerSpec)).Methods("GET")
	router.Use(api.authenticate)
	router.Use(api.handleCors)

	// Register unauthenticated routes.
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(api.handleCors)
		r.Use(api.requireNoAuth)
		r.HandleFunc("/login", api.handleLogin).Methods("POST")
		r.HandleFunc("/reset-password/request", api.handleResetPasswordRequest).Methods("POST")
		r.HandleFunc("/reset-password/validate", api.handleValidateResetToken).Methods("POST")
		r.HandleFunc("/reset-password", api.handleResetPassword).Methods("POST")
	}

	// Register authenticated routes.
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(api.handleCors)
		r.Use(api.requireAuth)

		r.HandleFunc("/users", api.handleGetUsers).Methods("GET")
		r.HandleFunc("/users/{id}", api.handleGetUserByID).Methods("GET")
		r.HandleFunc("/users", api.handleCreateUser).Methods("POST")
		r.HandleFunc("/users/{id}", api.handleUpdateUser).Methods("PUT")
		r.HandleFunc("/users/{id}", api.handleDeleteUser).Methods("DELETE")
	}
}

func byteHandler(b []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Write(b)
	}
}

func Sanitize(str string) string {
	sanitize, _ := regexp.Compile(`['"]`)
	return sanitize.ReplaceAllString(str, "")
}

type GeneralResult struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (api *API) httpGeneralWrite(status int, message string, data interface{}, resp http.ResponseWriter) {
	resp.Header().Set("Content-Type", "application/json")
	byte, err := json.Marshal(GeneralResult{
		Status:  status,
		Message: message,
		Data:    data,
	})
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	if status != 0 {
		resp.WriteHeader(status)
	}
	resp.Write(byte)
}
