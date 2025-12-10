package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/hekigan/couples/internal/build"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Determine build mode from ENV environment variable
	env := os.Getenv("ENV")
	mode := build.ModeDevelopment
	if env == "production" {
		mode = build.ModeProduction
	}

	// Create build configuration
	config := build.BuildConfig{
		Mode:      mode,
		SourceDir: "./static/js",
		OutputDir: "./static/dist",
		Minify:    mode == build.ModeProduction,
	}

	service := build.NewEsbuildService(config)

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down gracefully...")
		cancel()
	}()

	// Execute command
	switch command {
	case "build":
		log.Printf("🚀 Starting %s build...\n", mode)
		if err := service.BuildAll(ctx); err != nil {
			log.Fatalf("❌ Build failed: %v", err)
		}
		log.Println("🎉 Build complete!")

	case "watch":
		log.Printf("👀 Starting %s watch mode (Ctrl+C to stop)...\n", mode)
		if err := service.Watch(ctx); err != nil && err != context.Canceled {
			log.Fatalf("❌ Watch failed: %v", err)
		}
		log.Println("👋 Watch mode stopped")

	case "clean":
		log.Println("🗑️  Cleaning bundles...")
		if err := service.Clean(); err != nil {
			log.Fatalf("❌ Clean failed: %v", err)
		}
		log.Println("✨ Clean complete!")

	default:
		fmt.Printf("❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("esbuild - JavaScript bundler for Couple Card Game")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/esbuild/main.go <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  build   - Build JavaScript bundles")
	fmt.Println("  watch   - Watch and rebuild bundles on changes")
	fmt.Println("  clean   - Remove generated bundles")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  ENV=production   - Minified production build with external source maps")
	fmt.Println("  ENV=development  - Development build with inline source maps (default)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  ENV=production go run ./cmd/esbuild/main.go build")
	fmt.Println("  ENV=development go run ./cmd/esbuild/main.go watch")
	fmt.Println("  go run ./cmd/esbuild/main.go clean")
}

func getSourcemapMode(mode build.BuildMode) api.SourceMap {
	if mode == build.ModeProduction {
		return api.SourceMapExternal
	}
	return api.SourceMapInline
}
