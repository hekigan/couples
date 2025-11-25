package services

import (
	"fmt"
	"log"
	"os"
)

// ServiceLogger provides consistent logging across all services
// Standardizes log format and level
type ServiceLogger struct {
	serviceName string
}

// NewServiceLogger creates a new logger for a service
// Example usage:
//   logger := NewServiceLogger("RoomService")
func NewServiceLogger(serviceName string) *ServiceLogger {
	return &ServiceLogger{
		serviceName: serviceName,
	}
}

// Error logs an error message with ❌ emoji
// Use for critical errors that prevent operation completion
func (l *ServiceLogger) Error(format string, args ...interface{}) {
	log.Printf("❌ [%s] %s", l.serviceName, fmt.Sprintf(format, args...))
}

// Warn logs a warning message with ⚠️  emoji
// Use for recoverable issues or unexpected situations
func (l *ServiceLogger) Warn(format string, args ...interface{}) {
	log.Printf("⚠️  [%s] %s", l.serviceName, fmt.Sprintf(format, args...))
}

// Info logs an informational message with ℹ️  emoji
// Use for important state changes or key operations
func (l *ServiceLogger) Info(format string, args ...interface{}) {
	log.Printf("ℹ️  [%s] %s", l.serviceName, fmt.Sprintf(format, args...))
}

// Debug logs a debug message with 🐛 emoji
// Only logs when DEBUG environment variable is set to "true"
// Use for detailed debugging information
func (l *ServiceLogger) Debug(format string, args ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("🐛 [%s] %s", l.serviceName, fmt.Sprintf(format, args...))
	}
}

// Success logs a success message with ✅ emoji
// Use for successful completion of important operations
func (l *ServiceLogger) Success(format string, args ...interface{}) {
	log.Printf("✅ [%s] %s", l.serviceName, fmt.Sprintf(format, args...))
}
