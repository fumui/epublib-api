package http

import (
	"encoding/json"
	epublib "epublib"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type SendResetTokenRequest struct {
	Email string `json:"email"`
}
type ValidateResetTokenRequest struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}
type ResetPasswordRequest struct {
	ID       string `json:"id"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (api *API) handleResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	var payload SendResetTokenRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, err.Error(), nil, w)
		return
	}
	ctx := r.Context()
	auth, err := api.AuthService.FindAuthByEmail(ctx, payload.Email)
	if err != nil {
		if err == epublib.ErrNotFound {
			api.httpGeneralWrite(http.StatusNotFound, "email not registered", nil, w)
			return
		}
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	token, err := api.ResetTokenService.GenerateResetToken(ctx, auth)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	mail, err := buildMailForToken(auth, token)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	err = api.MailerService.SendMail(ctx, *mail)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	api.httpGeneralWrite(http.StatusOK, "Success", nil, w)
}

func (api *API) handleValidateResetToken(w http.ResponseWriter, r *http.Request) {
	var payload ValidateResetTokenRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, err.Error(), nil, w)
		return
	}
	ctx := r.Context()
	valid, err := api.ResetTokenService.ValidateResetToken(ctx, payload.ID, payload.Token)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	if !valid {
		api.httpGeneralWrite(http.StatusForbidden, "Invalid reset token", nil, w)
	}
	api.httpGeneralWrite(http.StatusOK, "Success", nil, w)
}

func (api *API) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload ResetPasswordRequest
	bodyBytes, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(bodyBytes, &payload)
	if err != nil {
		api.httpGeneralWrite(http.StatusBadRequest, err.Error(), nil, w)
		return
	}
	ctx := r.Context()
	valid, err := api.ResetTokenService.ValidateResetToken(ctx, payload.ID, payload.Token)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}
	if !valid {
		api.httpGeneralWrite(http.StatusForbidden, "Invalid reset token", nil, w)
		return
	}

	auth, err := api.ResetTokenService.UseResetToken(ctx, payload.ID, payload.Token)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	err = api.AuthService.ResetAuthPassword(ctx, auth.ID, auth.Email, payload.Password)
	if err != nil {
		api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
		return
	}

	api.httpGeneralWrite(http.StatusOK, "Success", nil, w)
}

func buildMailForToken(auth *epublib.Auth, token *epublib.ResetToken) (*epublib.Mail, error) {
	templateFilePath := os.Getenv("TEMPLATE_FILE_PATH")
	if templateFilePath == "" {
		err := fmt.Errorf("TEMPLATE_FILE_PATH environment variable not set")
		log.Println(err)
		return nil, err
	}
	senderAddr := os.Getenv("SMTP_SENDER_ADDR")
	if senderAddr == "" {
		err := fmt.Errorf("SMTP_SENDER_ADDR environment variable not set")
		log.Println(err)
		return nil, err
	}
	content, err := os.ReadFile(templateFilePath)
	if err != nil {
		fmt.Println("Error loading template:", err)
		return nil, err
	}
	variablesMap := map[string]interface{}{
		"RESET_TOKEN_ID": token.ID,
		"TOKEN":          token.Token,
		"USERNAME":       auth.Username,
	}
	emailContent := generateEmailContent(string(content), variablesMap)
	return &epublib.Mail{
		Channel:     "email",
		From:        senderAddr,
		To:          auth.Email,
		Subject:     "Reset Password",
		ContentType: "text/html",
		Body:        emailContent,
	}, nil

}

func generateEmailContent(template string, variablesMap map[string]interface{}) string {
	for key, value := range variablesMap {
		placeholder := fmt.Sprintf("$$%s$$", key)
		template = strings.ReplaceAll(template, placeholder, fmt.Sprint(value))
	}

	return template
}
