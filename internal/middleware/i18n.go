package middleware

import (
	"context"
	"net/http"
	"strings"
)

type i18nContextKey string

const LanguageKey i18nContextKey = "language"

// I18nMiddleware extracts language preference
func I18nMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "en" // Default language

		// Check query parameter
		if queryLang := r.URL.Query().Get("lang"); queryLang != "" {
			lang = queryLang
		} else if cookie, err := r.Cookie("language"); err == nil {
			lang = cookie.Value
		} else {
			// Check Accept-Language header
			acceptLang := r.Header.Get("Accept-Language")
			if acceptLang != "" {
				parts := strings.Split(acceptLang, ",")
				if len(parts) > 0 {
					langPart := strings.Split(parts[0], ";")[0]
					langCode := strings.Split(langPart, "-")[0]
					lang = strings.ToLower(strings.TrimSpace(langCode))
				}
			}
		}

		ctx := context.WithValue(r.Context(), LanguageKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


