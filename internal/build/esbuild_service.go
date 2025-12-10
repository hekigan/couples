package build

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

// BuildMode represents the build environment
type BuildMode string

const (
	ModeDevelopment BuildMode = "development"
	ModeProduction  BuildMode = "production"
)

// BuildConfig holds configuration for the esbuild service
type BuildConfig struct {
	Mode      BuildMode
	SourceDir string
	OutputDir string
	Minify    bool
}

// EsbuildService handles JavaScript bundling with esbuild
type EsbuildService struct {
	config BuildConfig
}

// NewEsbuildService creates a new esbuild service instance
func NewEsbuildService(config BuildConfig) *EsbuildService {
	return &EsbuildService{config: config}
}

// BuildApp builds the main app bundle (htmx, sse, ui-utils, modal, notifications)
func (s *EsbuildService) BuildApp(ctx context.Context) error {
	log.Println("Building app bundle...")

	entryPoint := "./esbuild/app-entry.js"
	outfile := filepath.Join(s.config.OutputDir, "app.bundle.js")

	opts := s.getBuildOptions(entryPoint, outfile)

	result := api.Build(opts)
	if len(result.Errors) > 0 {
		return fmt.Errorf("app bundle build failed: %v", result.Errors)
	}

	if len(result.Warnings) > 0 {
		log.Printf("⚠️  App bundle warnings: %v\n", result.Warnings)
	}

	// Add source map URL comment for production builds
	if s.config.Mode == ModeProduction {
		if err := s.addSourceMapComment(outfile, "app.bundle.js.map"); err != nil {
			log.Printf("⚠️  Failed to add source map comment: %v\n", err)
		}
	}

	log.Printf("✅ App bundle built: %s\n", outfile)
	return nil
}

// BuildAdmin builds the admin bundle (admin utilities)
func (s *EsbuildService) BuildAdmin(ctx context.Context) error {
	log.Println("Building admin bundle...")

	entryPoint := "./esbuild/admin-entry.js"
	outfile := filepath.Join(s.config.OutputDir, "admin.bundle.js")

	opts := s.getBuildOptions(entryPoint, outfile)

	result := api.Build(opts)
	if len(result.Errors) > 0 {
		return fmt.Errorf("admin bundle build failed: %v", result.Errors)
	}

	if len(result.Warnings) > 0 {
		log.Printf("⚠️  Admin bundle warnings: %v\n", result.Warnings)
	}

	// Add source map URL comment for production builds
	if s.config.Mode == ModeProduction {
		if err := s.addSourceMapComment(outfile, "admin.bundle.js.map"); err != nil {
			log.Printf("⚠️  Failed to add source map comment: %v\n", err)
		}
	}

	log.Printf("✅ Admin bundle built: %s\n", outfile)
	return nil
}

// BuildAll builds both app and admin bundles
func (s *EsbuildService) BuildAll(ctx context.Context) error {
	start := time.Now()

	// Ensure output directory exists
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build app bundle
	if err := s.BuildApp(ctx); err != nil {
		return err
	}

	// Build admin bundle
	if err := s.BuildAdmin(ctx); err != nil {
		return err
	}

	elapsed := time.Since(start)
	log.Printf("✅ All bundles built in %v\n", elapsed)
	return nil
}

// Watch starts watch mode and rebuilds on changes
// Note: This implementation uses periodic rebuilds as the esbuild Go API
// version doesn't have native watch mode support
func (s *EsbuildService) Watch(ctx context.Context) error {
	log.Println("Starting watch mode (polling-based)...")
	log.Println("⚠️  Note: For better performance, consider using 'npm run esbuild -- --watch' or upgrade esbuild")

	// Initial build
	if err := s.BuildAll(ctx); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	// Setup ticker for periodic rebuilds (every 2 seconds)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Track last modification times
	lastModTimes := make(map[string]time.Time)

	// Get all JS files to watch
	jsFiles, err := filepath.Glob(filepath.Join(s.config.SourceDir, "*.js"))
	if err != nil {
		return fmt.Errorf("failed to glob JS files: %w", err)
	}

	log.Println("👀 Watching JavaScript files for changes...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping watch mode...")
			return nil
		case <-ticker.C:
			// Check if any files have changed
			changed := false
			for _, file := range jsFiles {
				info, err := os.Stat(file)
				if err != nil {
					continue
				}
				modTime := info.ModTime()
				if lastMod, ok := lastModTimes[file]; ok {
					if modTime.After(lastMod) {
						changed = true
					}
				}
				lastModTimes[file] = modTime
			}

			// Rebuild if changes detected
			if changed {
				log.Println("📝 Changes detected, rebuilding...")
				if err := s.BuildAll(ctx); err != nil {
					log.Printf("❌ Rebuild failed: %v\n", err)
				} else {
					log.Println("✅ Rebuild complete")
				}
			}
		}
	}
}

// Clean removes generated bundles
func (s *EsbuildService) Clean() error {
	log.Println("Cleaning bundles...")

	patterns := []string{
		filepath.Join(s.config.OutputDir, "*.bundle.js"),
		filepath.Join(s.config.OutputDir, "*.bundle.js.map"),
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("glob pattern error: %w", err)
		}

		for _, file := range matches {
			if err := os.Remove(file); err != nil {
				log.Printf("⚠️  Failed to remove %s: %v\n", file, err)
			} else {
				log.Printf("🗑️  Removed %s\n", file)
			}
		}
	}

	log.Println("✅ Clean complete")
	return nil
}

// addSourceMapComment appends the source map URL comment to a bundle file
func (s *EsbuildService) addSourceMapComment(bundleFile, mapFile string) error {
	// Open bundle file in append mode
	f, err := os.OpenFile(bundleFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open bundle file: %w", err)
	}
	defer f.Close()

	// Append source map URL comment
	comment := fmt.Sprintf("\n//# sourceMappingURL=%s\n", mapFile)
	if _, err := f.WriteString(comment); err != nil {
		return fmt.Errorf("failed to write source map comment: %w", err)
	}

	return nil
}

// getBuildOptions creates esbuild options based on configuration
func (s *EsbuildService) getBuildOptions(entryPoint, outfile string) api.BuildOptions {
	var sourcemap api.SourceMap
	var minify bool

	if s.config.Mode == ModeProduction {
		sourcemap = api.SourceMapExternal // External .map files for production
		minify = true
	} else {
		sourcemap = api.SourceMapInline // Inline source maps for development
		minify = false
	}

	return api.BuildOptions{
		EntryPoints:       []string{entryPoint},
		Bundle:            true,
		Outfile:           outfile,
		Sourcemap:         sourcemap,
		MinifyWhitespace:  minify,
		MinifyIdentifiers: minify,
		MinifySyntax:      minify,
		Target:            api.ES2020,
		Format:            api.FormatIIFE, // Preserve window.* exports
		LogLevel:          api.LogLevelInfo,
		Write:             true,
	}
}
