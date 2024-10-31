package http

import (
	"encoding/json"
	"epublib"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponseData struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Level    string `json:"level"`
}

func (api *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, err.Error(), nil, w)
		return
	}
	ctx := r.Context()
	auth, err := api.AuthService.FindAuthByEmailPass(ctx, payload.Email, payload.Password)
	if err != nil {
		if err == epublib.ErrNotFound {
			api.httpGeneralWrite(http.StatusForbidden, "Incorrect email or password", nil, w)
			return
		}
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	user, err := api.UserService.FindUserByID(ctx, auth.UserID)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	response := &LoginResponseData{
		UserID:   user.ID,
		Username: auth.Username,
		Level:    auth.Level.String(),
	}
	response.Token, err = createJWT(auth.UserID, 24*time.Hour)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	api.httpGeneralWrite(http.StatusOK, "Success", response, w)
}

func getJWTKey(token *jwt.Token) (interface{}, error) {
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		return nil, fmt.Errorf("JWT_KEY is empty")
	}
	return []byte(jwtKey), nil
}
func createJWT(subject string, duration time.Duration) (string, error) {
	jwtKey, err := getJWTKey(nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Audience:  []string{"epublib"},
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
	})
	tokenString, err := token.SignedString(jwtKey.([]byte))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return tokenString, nil
}
func decodeJWT(token string) (*jwt.Token, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, getJWTKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return parsed, nil
}
