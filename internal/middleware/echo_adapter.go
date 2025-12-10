package middleware

import (
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// Context keys for storing values in Echo context
const (
	UserIDKey   = "user_id"
	LanguageKey = "language"
	IsAdminKey  = "is_admin"
)

// GetSession retrieves the session from the Echo context
// This bridges gorilla/sessions with Echo by accessing the underlying http.Request
func GetSession(c echo.Context) (*sessions.Session, error) {
	return Store.Get(c.Request(), "couple-card-game-session")
}

// SaveSession saves the session to the response
func SaveSession(c echo.Context, session *sessions.Session) error {
	return session.Save(c.Request(), c.Response())
}

// SetUserID stores the user ID in the Echo context
func SetUserID(c echo.Context, id uuid.UUID) {
	c.Set(UserIDKey, id)
}

// GetUserID retrieves the user ID from the Echo context
func GetUserID(c echo.Context) (uuid.UUID, bool) {
	id, ok := c.Get(UserIDKey).(uuid.UUID)
	return id, ok
}

// SetLanguage stores the language in the Echo context
func SetLanguage(c echo.Context, lang string) {
	c.Set(LanguageKey, lang)
}

// GetLanguage retrieves the language from the Echo context
func GetLanguage(c echo.Context) (string, bool) {
	lang, ok := c.Get(LanguageKey).(string)
	return lang, ok
}

// SetIsAdmin stores the admin flag in the Echo context
func SetIsAdmin(c echo.Context, isAdmin bool) {
	c.Set(IsAdminKey, isAdmin)
}

// GetIsAdmin retrieves the admin flag from the Echo context
func GetIsAdmin(c echo.Context) (bool, bool) {
	isAdmin, ok := c.Get(IsAdminKey).(bool)
	return isAdmin, ok
}
