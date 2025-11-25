package services

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"
)

// TemplateFuncMap contains custom template functions for partials
var TemplateFuncMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b int) int { return a * b },
	"gte": func(a, b int) bool { return a >= b },
	"lte": func(a, b int) bool { return a <= b },
	"until": func(count int) []int {
		result := make([]int, count)
		for i := range result {
			result[i] = i
		}
		return result
	},
}

// TemplateService handles rendering of HTML fragments for SSE
type TemplateService struct {
	templates    *template.Template
	templatesDir string
	isProduction bool
	mu           sync.RWMutex
}

// NewTemplateService creates a new template service
func NewTemplateService(templatesDir string) (*TemplateService, error) {
	isProduction := os.Getenv("ENV") == "production"

	var templates *template.Template
	var err error

	// Production mode: Parse and cache all templates at startup
	if isProduction {
		partialsPattern := filepath.Join(templatesDir, "partials/**/*.html")
		templates, err = template.New("").Funcs(TemplateFuncMap).ParseGlob(partialsPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse templates: %w", err)
		}
	}
	// Development mode: Skip caching, parse on-demand for hot-reload

	return &TemplateService{
		templates:    templates,
		templatesDir: templatesDir,
		isProduction: isProduction,
	}, nil
}

// RenderFragment renders a template fragment with the given data
func (s *TemplateService) RenderFragment(templateName string, data interface{}) (string, error) {
	var templates *template.Template

	// Development mode: Parse templates on-demand for hot-reload
	if !s.isProduction {
		partialsPattern := filepath.Join(s.templatesDir, "partials/**/*.html")
		var err error
		templates, err = template.New("").Funcs(TemplateFuncMap).ParseGlob(partialsPattern)
		if err != nil {
			return "", fmt.Errorf("failed to parse templates: %w", err)
		}
	} else {
		// Production mode: Use cached templates
		s.mu.RLock()
		defer s.mu.RUnlock()
		templates = s.templates
	}

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, templateName, data); err != nil {
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

// NOTE: All template data structures have been moved to template_models.go
// This keeps this file focused on the rendering logic while centralizing
// all data structure definitions in a single location.
