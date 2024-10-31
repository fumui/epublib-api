package http

import (
	epublib "epublib"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4"
)

// authenticate is middleware for loading session data from a cookie or API key header.
func (api *API) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Login via API key, if available.
		if v := r.Header.Get("Authorization"); strings.HasPrefix(v, "Bearer ") {
			token := strings.TrimPrefix(v, "Bearer ")
			decoded, err := decodeJWT(token)
			// Lookup user by API key. Display error if not found.
			// Otherwise set
			if err != nil {
				if errors.Is(err, jwt.ErrTokenMalformed) {
					api.httpGeneralWrite(http.StatusBadRequest, err.Error(), nil, w)
					return
				} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
					api.httpGeneralWrite(http.StatusForbidden, err.Error(), nil, w)
					return
				} else {
					log.Println(err)
					api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
					return
				}
			}
			claims, ok := decoded.Claims.(*jwt.RegisteredClaims)
			if !ok {
				log.Println(err)
			}
			// Find authenticated user data
			user, err := api.UserService.FindUserByID(r.Context(), claims.Subject)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					api.httpGeneralWrite(http.StatusForbidden, err.Error(), nil, w)
					return
				} else {
					log.Println(err)
					api.httpGeneralWrite(http.StatusInternalServerError, err.Error(), nil, w)
					return
				}
			}

			// Update request context to include authenticated user.
			r = r.WithContext(epublib.NewContextWithUser(r.Context(), user))

			// Delegate to next HTTP handler.
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (api *API) handleCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("CORS_ALLOW_ALL") == "true" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
			w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAuth is middleware for requiring authentication. This is used by
// nearly every page except for the login & oauth pages.
func (api *API) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If user is logged in, delegate to next HTTP handler.
		if userID := epublib.UserIDFromContext(r.Context()); userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise save the current URL (without scheme/host).
		// redirectURL := r.URL
		// redirectURL.Scheme, redirectURL.Host = "", ""

		// // Save the URL to the session and redirect to the log in page.
		// // On successful login, the user will be redirected to their original location.
		// session, _ := api.session(r)
		// session.RedirectURL = redirectURL.String()
		// if err := api.setSession(w, session); err != nil {
		// 	log.Printf("http: cannot set session: %s", err)
		// }
		// http.Redirect(w, r, "/login", http.StatusFound)
		api.httpGeneralWrite(401, "Unauthorized", "please login first", w)
	})
}

// requireNoAuth is middleware for requiring no authentication.
// This is used if a user goes to log in but is already logged in.
func (api *API) requireNoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If user is logged in, redirect to the home page.
		if userID := epublib.UserIDFromContext(r.Context()); userID != "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Delegate to next HTTP handler.
		next.ServeHTTP(w, r)
	})
}
