package services

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
)

// TemplateService handles rendering of HTML fragments for SSE
type TemplateService struct {
	templates *template.Template
	mu        sync.RWMutex
}

// NewTemplateService creates a new template service
func NewTemplateService(templatesDir string) (*TemplateService, error) {
	// Parse all partial templates
	partialsPattern := filepath.Join(templatesDir, "partials/**/*.html")
	templates, err := template.ParseGlob(partialsPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &TemplateService{
		templates: templates,
	}, nil
}

// RenderFragment renders a template fragment with the given data
func (s *TemplateService) RenderFragment(templateName string, data interface{}) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// ReloadTemplates reloads all templates (useful for development)
func (s *TemplateService) ReloadTemplates(templatesDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partialsPattern := filepath.Join(templatesDir, "partials/**/*.html")
	templates, err := template.ParseGlob(partialsPattern)
	if err != nil {
		return fmt.Errorf("failed to reload templates: %w", err)
	}

	s.templates = templates
	return nil
}

// Common data structures for templates

// JoinRequestData represents data for join request partial
type JoinRequestData struct {
	ID        string
	RoomID    string
	Username  string
	CreatedAt string
}

// PlayerJoinedData represents data for player joined partial
type PlayerJoinedData struct {
	Username string
	UserID   string
}

// RequestAcceptedData represents data for request accepted partial
type RequestAcceptedData struct {
	GuestUsername string
}

// QuestionDrawnData represents data for question drawn partial
type QuestionDrawnData struct {
	RoomID                string
	QuestionNumber        int
	MaxQuestions          int
	Category              string
	CategoryLabel         string
	QuestionText          string
	IsMyTurn              bool
	CurrentPlayerUsername string
}

// AnswerSubmittedData represents data for answer submitted partial
type AnswerSubmittedData struct {
	RoomID                string
	Username              string
	AnswerText            string
	ActionType            string // "answered" or "passed"
	IsMyTurn              bool   // Is it now my turn to draw next question?
	CurrentPlayerUsername string
}

// GameStartedData represents data for game started partial
type GameStartedData struct {
	RoomID string
}

// NotificationData represents data for notification partial
type NotificationData struct {
	ID      string
	Type    string
	Title   string
	Message string
	Link    string
}
