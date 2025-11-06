package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/supabase-community/supabase-go"
)

// I18nService handles internationalization
type I18nService struct {
	client        *supabase.Client
	translationDir string
	translations   map[string]map[string]string // language -> key -> value
}

// NewI18nService creates a new i18n service
func NewI18nService(client *supabase.Client, translationDir string) *I18nService {
	service := &I18nService{
		client:        client,
		translationDir: translationDir,
		translations:   make(map[string]map[string]string),
	}
	service.loadTranslations()
	return service
}

// loadTranslations loads translation files from disk
func (s *I18nService) loadTranslations() error {
	files, err := filepath.Glob(filepath.Join(s.translationDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to list translation files: %w", err)
	}

	for _, file := range files {
		lang := filepath.Base(file[:len(file)-5]) // Remove .json extension
		
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			continue
		}

		s.translations[lang] = translations
	}

	return nil
}

// GetTranslation retrieves a translation for a key and language
func (s *I18nService) GetTranslation(ctx context.Context, lang, key string) (string, error) {
	if translations, ok := s.translations[lang]; ok {
		if value, ok := translations[key]; ok {
			return value, nil
		}
	}
	
	// Fallback to English
	if translations, ok := s.translations["en"]; ok {
		if value, ok := translations[key]; ok {
			return value, nil
		}
	}

	return key, nil // Return key if translation not found
}

