package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	Client        *resend.Client // Exported for test handler
	from          string         // Configured sender email
	baseURL       string         // App base URL for links
	isDevelopment bool           // Whether running in development mode
}

func NewEmailService(apiKey, fromEmail, baseURL string) *EmailService {
	isDev := os.Getenv("ENV") == "development"
	if isDev {
		log.Println("📧 Email service running in DEVELOPMENT mode - using test emails (delivered@resend.dev)")
	}
	return &EmailService{
		Client:        resend.NewClient(apiKey),
		from:          fromEmail,
		baseURL:       baseURL,
		isDevelopment: isDev,
	}
}

// SendFriendInvitationToExistingUser sends email to existing user + in-app notification
func (s *EmailService) SendFriendInvitationToExistingUser(ctx context.Context, toEmail, senderUsername string) error {
	// Use test email in development mode
	recipient := toEmail
	tags := []resend.Tag{}

	if s.isDevelopment {
		recipient = "delivered@resend.dev"
		tags = []resend.Tag{
			{Name: "environment", Value: "development"},
			{Name: "email_type", Value: "friend_request_existing"},
			{Name: "original_recipient", Value: toEmail},
		}
		log.Printf("📧 [DEV] Sending friend request email to TEST address (original: %s, sender: %s)", toEmail, senderUsername)
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{recipient},
		Subject: fmt.Sprintf("%s sent you a friend request", senderUsername),
		Html:    s.buildExistingUserEmailHTML(senderUsername),
		Tags:    tags,
	}

	_, err := s.Client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("❌ Failed to send friend request email: %v", err)
	}
	return err
}

// SendFriendInvitationToNewUser sends join invitation with token link
func (s *EmailService) SendFriendInvitationToNewUser(ctx context.Context, toEmail, senderUsername, token string) error {
	signupLink := fmt.Sprintf("%s/auth/signup?friend_invitation=%s", s.baseURL, token)

	// Use test email in development mode
	recipient := toEmail
	tags := []resend.Tag{}

	if s.isDevelopment {
		recipient = "delivered@resend.dev"
		tags = []resend.Tag{
			{Name: "environment", Value: "development"},
			{Name: "email_type", Value: "friend_invite_new_user"},
			{Name: "original_recipient", Value: toEmail},
			{Name: "signup_token", Value: token[:16] + "..."}, // First 16 chars of token
		}
		log.Printf("📧 [DEV] Sending join invitation email to TEST address (original: %s, sender: %s, token: %s...)", toEmail, senderUsername, token[:16])
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{recipient},
		Subject: fmt.Sprintf("%s invited you to join Couple Card Game", senderUsername),
		Html:    s.buildNewUserEmailHTML(senderUsername, signupLink),
		Tags:    tags,
	}

	_, err := s.Client.Emails.SendWithContext(ctx, params)
	if err != nil {
		log.Printf("❌ Failed to send join invitation email: %v", err)
	}
	return err
}

// Getter methods for private fields
func (s *EmailService) From() string {
	return s.from
}

func (s *EmailService) IsDevelopment() bool {
	return s.isDevelopment
}

// Email HTML templates
func (s *EmailService) buildExistingUserEmailHTML(senderUsername string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background: #667eea; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0;">
			<h1>🎴 Couple Card Game</h1>
		</div>
		<div style="background: #f7fafc; padding: 30px; border-radius: 0 0 8px 8px;">
			<h2>New Friend Request!</h2>
			<p>Hi there,</p>
			<p><strong>%s</strong> sent you a friend request on Couple Card Game.</p>
			<p>Connect with them to play together and strengthen your relationship through meaningful conversations.</p>
			<a href="%s/friends" style="display: inline-block; padding: 12px 24px; background: #667eea; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0;">View Friend Requests</a>
			<p style="color: #718096; font-size: 14px;">
				Or copy and paste this link into your browser:<br>
				%s/friends
			</p>
		</div>
		<div style="text-align: center; margin-top: 30px; color: #718096; font-size: 14px;">
			<p>Couple Card Game - Play together, learn together</p>
		</div>
	</div>
</body>
</html>
`, senderUsername, s.baseURL, s.baseURL)
}

func (s *EmailService) buildNewUserEmailHTML(senderUsername, signupLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<div style="background: #667eea; color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0;">
			<h1>🎴 You're Invited!</h1>
		</div>
		<div style="background: #f7fafc; padding: 30px; border-radius: 0 0 8px 8px;">
			<h2>Join Couple Card Game</h2>
			<p>Hi there,</p>
			<p><strong>%s</strong> invited you to join Couple Card Game!</p>
			<p>Couple Card Game is a fun way to strengthen relationships through meaningful question-based conversations. Play with your partner, friends, or family.</p>
			<h3>What you'll get:</h3>
			<ul>
				<li>Hundreds of thoughtful questions</li>
				<li>Multiple categories (romance, dreams, past, etc.)</li>
				<li>Real-time multiplayer gameplay</li>
				<li>Multi-language support</li>
			</ul>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background: #667eea; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0;">Accept Invitation & Sign Up</a>
			<p style="color: #718096; font-size: 14px;">
				This invitation expires in 7 days.<br>
				Or copy and paste this link into your browser:<br>
				%s
			</p>
		</div>
		<div style="text-align: center; margin-top: 30px; color: #718096; font-size: 14px;">
			<p>Couple Card Game - Play together, learn together</p>
		</div>
	</div>
</body>
</html>
`, senderUsername, signupLink, signupLink)
}
