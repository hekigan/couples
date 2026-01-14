package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// customHTTPErrorHandler provides HTMX-compatible error handling
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		} else {
			message = fmt.Sprintf("%v", he.Message)
		}
	}

	// Check if HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		c.HTML(code, fmt.Sprintf("<div class='error'>%s</div>", message))
		return
	}

	// Check if JSON request
	if c.Request().Header.Get("Accept") == "application/json" || c.Request().Header.Get("Content-Type") == "application/json" {
		c.JSON(code, map[string]string{"error": message})
		return
	}

	// Default: plain text
	c.String(code, message)
}
