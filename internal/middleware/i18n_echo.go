package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Supported languages
var supportedLanguages = map[string]bool{
	"en": true,
	"fr": true,
	"ja": true,
}

// EchoI18n detects and sets the user's language preference
// Priority: query param ?lang= → cookie → Accept-Language header → default "en"
func EchoI18n() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var language string

			// 1. Check query parameter
			if lang := c.QueryParam("lang"); lang != "" && supportedLanguages[lang] {
				language = lang
			}

			// 2. Check cookie if query param not present
			if language == "" {
				if cookie, err := c.Cookie("language"); err == nil {
					if supportedLanguages[cookie.Value] {
						language = cookie.Value
					}
				}
			}

			// 3. Check Accept-Language header
			if language == "" {
				acceptLanguage := c.Request().Header.Get("Accept-Language")
				if acceptLanguage != "" {
					// Parse Accept-Language header (e.g., "en-US,en;q=0.9,fr;q=0.8")
					languages := strings.Split(acceptLanguage, ",")
					for _, lang := range languages {
						// Extract language code (before - or ;)
						langCode := strings.Split(strings.TrimSpace(lang), "-")[0]
						langCode = strings.Split(langCode, ";")[0]

						if supportedLanguages[langCode] {
							language = langCode
							break
						}
					}
				}
			}

			// 4. Default to "en"
			if language == "" {
				language = "en"
			}

			// Store in context
			SetLanguage(c, language)

			// Set cookie if different from current
			if cookie, err := c.Cookie("language"); err != nil || cookie.Value != language {
				c.SetCookie(&http.Cookie{
					Name:     "language",
					Value:    language,
					Path:     "/",
					MaxAge:   86400 * 365, // 1 year
					HttpOnly: false,        // Allow JavaScript access for i18n
					Secure:   c.IsTLS(),
					SameSite: http.SameSiteLaxMode,
				})
			}

			return next(c)
		}
	}
}
