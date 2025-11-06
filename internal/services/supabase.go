package services

import (
	"fmt"
	"os"

	"github.com/supabase-community/supabase-go"
)

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient() (*supabase.Client, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return client, nil
}

