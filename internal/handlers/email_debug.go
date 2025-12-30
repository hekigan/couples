package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/resend/resend-go/v2"
)

// TestEmailHandler sends a test email to verify Resend configuration
// Access via: GET /api/test-email?to=your@email.com
func (h *Handler) TestEmailHandler(c echo.Context) error {
	toEmail := c.QueryParam("to")
	if toEmail == "" {
		toEmail = "delivered@resend.dev" // Default to test email
	}

	// Check if EmailService is configured
	if h.EmailService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error":   "Email service not configured",
			"message": "Please set RESEND_API_KEY, EMAIL_FROM, and APP_BASE_URL in .env file",
		})
	}

	// Prepare test email
	from := h.EmailService.From()
	subject := fmt.Sprintf("Test Email from Couple Card Game - %s", time.Now().Format("15:04:05"))
	html := `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
	<div style="max-width: 600px; margin: 0 auto; border: 2px solid #667eea; border-radius: 8px; padding: 30px;">
		<h1 style="color: #667eea;">✅ Email Test Successful!</h1>
		<p>This is a test email from your Couple Card Game application.</p>
		<p><strong>Timestamp:</strong> ` + time.Now().Format("2006-01-02 15:04:05 MST") + `</p>
		<p><strong>Environment:</strong> ` + c.Request().Host + `</p>
		<hr style="border: 1px solid #e5e7eb; margin: 20px 0;">
		<p style="color: #6b7280; font-size: 14px;">
			If you received this email, your Resend integration is working correctly! 🎉
		</p>
	</div>
</body>
</html>`

	// Add tags for development mode
	tags := []resend.Tag{
		{Name: "test", Value: "true"},
		{Name: "timestamp", Value: time.Now().Format("2006-01-02_15-04-05")},
	}

	// Send email using Resend client directly
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: subject,
		Html:    html,
		Tags:    tags,
	}

	// Use background context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sent, err := h.EmailService.Client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to send test email. Check your RESEND_API_KEY and configuration.",
			"config": map[string]interface{}{
				"from":           from,
				"to":             toEmail,
				"api_key_set":    h.EmailService.Client != nil,
				"is_development": h.EmailService.IsDevelopment(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Test email sent successfully! Check your inbox (or Resend dashboard).",
		"details": map[string]interface{}{
			"email_id":       sent.Id,
			"from":           from,
			"to":             toEmail,
			"subject":        subject,
			"is_development": h.EmailService.IsDevelopment(),
			"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
		},
		"next_steps": []string{
			"Check Resend dashboard: https://resend.com/emails",
			fmt.Sprintf("Look for email ID: %s", sent.Id),
			fmt.Sprintf("Recipient: %s", toEmail),
			"Email should arrive within seconds",
		},
	})
}
