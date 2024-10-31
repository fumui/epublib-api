package http

import (
	"encoding/json"
	epublib "epublib"
	"epublib/postgres"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (api *API) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	offset, _ := strconv.Atoi(queryParams.Get("offset"))
	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	includeDeleted := queryParams.Get("include_deleted") == "true"

	// Construct filter based on query parameters
	filter := epublib.UserFilter{
		Name:           name,
		Offset:         offset,
		Limit:          limit,
		IncludeDeleted: includeDeleted,
	}

	// Retrieve users from the service
	users, totalCount, err := api.UserService.FindUsers(r.Context(), filter)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"users":       users,
		"total_count": totalCount,
	}

	// Send the response
	api.httpGeneralWrite(http.StatusOK, "Success", response, w)
}

func (api *API) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path
	vars := mux.Vars(r)
	id := vars["id"]

	// Retrieve user by ID from the service
	user, err := api.UserService.FindUserByID(r.Context(), id)
	if err != nil {
		if err == epublib.ErrNotFound {
			api.httpGeneralWrite(http.StatusNotFound, "User not found", nil, w)
			return
		}
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"user": user,
	}

	// Send the response
	api.httpGeneralWrite(http.StatusOK, "Success", response, w)
}

type CreateUserRequest struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Email    string            `json:"email"`
	Level    epublib.AuthLevel `json:"level"`
}

func (api *API) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Determine authorization
	ctx := r.Context()
	currentUser := epublib.UserFromContext(ctx)
	currentUserAuth, err := api.AuthService.FindAuthByUserID(ctx, currentUser.ID)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	if currentUserAuth.Level != epublib.AdminLevel {
		api.httpGeneralWrite(http.StatusForbidden, "Only admin can create user", nil, w)
		return
	}

	// Parse the JSON request body into a CreateUserRequest struct
	var payload CreateUserRequest
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, "Invalid JSON payload", nil, w)
		return
	}

	// Validate the required fields
	if payload.Username == "" {
		api.httpGeneralWrite(http.StatusBadRequest, "username is required field", nil, w)
		return
	}
	if payload.Password == "" {
		api.httpGeneralWrite(http.StatusBadRequest, "password is required field", nil, w)
		return
	}
	if payload.Email == "" {
		api.httpGeneralWrite(http.StatusBadRequest, "email is required field", nil, w)
		return
	}
	if !payload.Level.IsValid() {
		api.httpGeneralWrite(http.StatusBadRequest, "invalid user level", nil, w)
		return
	}

	// Check if the email is already registered
	_, err = api.AuthService.FindAuthByEmail(ctx, payload.Email)
	if err == nil {
		api.httpGeneralWrite(http.StatusBadRequest, "email is already registered", nil, w)
		return
	}

	//Begin transaction
	ctx, err = postgres.BeginTx(ctx, api.db)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	defer postgres.Rollback(ctx)

	// Create the user using the service

	user := epublib.User{
		Name:        payload.Username,
		Address:     "",
		PhoneNumber: "",
		Gender:      epublib.UnidentifiedGender,
		BirthDate:   time.Now(),
		ImgProfile:  "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = api.UserService.CreateUser(ctx, &user)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Create the auth using the service
	auth := epublib.Auth{
		UserID:   user.ID,
		Username: payload.Username,
		Password: payload.Password,
		Email:    payload.Email,
		Level:    epublib.AuthLevel(payload.Level),
	}
	err = api.AuthService.CreateAuth(ctx, &auth)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Commit transaction
	err = postgres.Commit(ctx)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"user": user,
	}

	// Send the response
	api.httpGeneralWrite(http.StatusCreated, "User created successfully", response, w)
}

func (api *API) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path
	vars := mux.Vars(r)
	id := vars["id"]

	// Parse the JSON request body into a UserUpdate struct
	var update epublib.UserUpdate
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, "Invalid JSON payload", nil, w)
		return
	}

	// Validate the required fields
	if update.Name == "" {
		api.httpGeneralWrite(http.StatusBadRequest, "Name is required fields", nil, w)
		return
	}

	// Update the user using the service
	user, err := api.UserService.UpdateUser(r.Context(), id, update)
	if err != nil {
		if err == epublib.ErrNotFound {
			api.httpGeneralWrite(http.StatusNotFound, "User not found", nil, w)
			return
		}
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"user": user,
	}

	// Send the response
	api.httpGeneralWrite(http.StatusOK, "User updated successfully", response, w)
}

func (api *API) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the URL path
	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	//Begin transaction
	ctx, err := postgres.BeginTx(ctx, api.db)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	defer postgres.Rollback(ctx)
	// Soft delete the user using the service
	err = api.UserService.DeleteUser(r.Context(), id)
	if err != nil {
		if err == epublib.ErrNotFound {
			api.httpGeneralWrite(http.StatusNotFound, "User not found", nil, w)
			return
		}
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	auth, err := api.AuthService.FindAuthByUserID(ctx, id)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	err = api.AuthService.DeleteAuth(r.Context(), auth.ID)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Commit transaction
	err = postgres.Commit(ctx)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	// Send the response
	api.httpGeneralWrite(http.StatusOK, "User deleted successfully", nil, w)
}
