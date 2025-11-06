package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var Store *sessions.CookieStore

// InitSessionStore initializes the session store
func InitSessionStore() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	}
}

