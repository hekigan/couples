package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

// NewServer creates a new Echo instance with default configuration
func NewServer(port string) *echo.Echo {
	e := echo.New()

	// Hide banner in production
	if os.Getenv("ENV") == "production" {
		e.HideBanner = true
	}

	return e
}

// StartWithGracefulShutdown starts the Echo server with graceful shutdown support
// It listens for SIGINT and SIGTERM signals and shuts down the server gracefully
// with a 30-second timeout
func StartWithGracefulShutdown(e *echo.Echo, port string) {
	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", port)

		// Configure server with WriteTimeout=0 for SSE support
		e.Server.Addr = addr
		e.Server.ReadTimeout = 15 * time.Second
		e.Server.WriteTimeout = 0 // Disabled for SSE long-lived connections
		e.Server.IdleTimeout = 60 * time.Second

		log.Printf("🚀 Server running on http://localhost:%s", port)

		if err := e.Start(addr); err != nil {
			// Only log if it's not a graceful shutdown
			if err.Error() != "http: Server closed" {
				log.Printf("Server error: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Graceful shutdown with 30-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server exited properly")
}
