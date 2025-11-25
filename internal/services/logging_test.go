package services

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestNewServiceLogger(t *testing.T) {
	logger := NewServiceLogger("TestService")

	if logger == nil {
		t.Fatal("NewServiceLogger() returned nil")
	}

	if logger.serviceName != "TestService" {
		t.Errorf("serviceName = %s, want TestService", logger.serviceName)
	}
}

func TestServiceLogger_Error(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewServiceLogger("TestService")
	logger.Error("test error: %s", "failure")

	output := buf.String()

	if !strings.Contains(output, "❌") {
		t.Error("Error log should contain ❌ emoji")
	}

	if !strings.Contains(output, "[TestService]") {
		t.Error("Error log should contain service name")
	}

	if !strings.Contains(output, "test error: failure") {
		t.Error("Error log should contain formatted message")
	}
}

func TestServiceLogger_Warn(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewServiceLogger("TestService")
	logger.Warn("test warning: %d items", 5)

	output := buf.String()

	if !strings.Contains(output, "⚠️") {
		t.Error("Warn log should contain ⚠️ emoji")
	}

	if !strings.Contains(output, "[TestService]") {
		t.Error("Warn log should contain service name")
	}

	if !strings.Contains(output, "test warning: 5 items") {
		t.Error("Warn log should contain formatted message")
	}
}

func TestServiceLogger_Info(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewServiceLogger("TestService")
	logger.Info("test info: %s", "information")

	output := buf.String()

	if !strings.Contains(output, "ℹ️") {
		t.Error("Info log should contain ℹ️ emoji")
	}

	if !strings.Contains(output, "[TestService]") {
		t.Error("Info log should contain service name")
	}

	if !strings.Contains(output, "test info: information") {
		t.Error("Info log should contain formatted message")
	}
}

func TestServiceLogger_Success(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewServiceLogger("TestService")
	logger.Success("operation completed: %s", "success")

	output := buf.String()

	if !strings.Contains(output, "✅") {
		t.Error("Success log should contain ✅ emoji")
	}

	if !strings.Contains(output, "[TestService]") {
		t.Error("Success log should contain service name")
	}

	if !strings.Contains(output, "operation completed: success") {
		t.Error("Success log should contain formatted message")
	}
}

func TestServiceLogger_Debug_WithDebugEnabled(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Enable debug mode
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")

	logger := NewServiceLogger("TestService")
	logger.Debug("debug message: %s", "details")

	output := buf.String()

	if !strings.Contains(output, "🐛") {
		t.Error("Debug log should contain 🐛 emoji when DEBUG=true")
	}

	if !strings.Contains(output, "[TestService]") {
		t.Error("Debug log should contain service name")
	}

	if !strings.Contains(output, "debug message: details") {
		t.Error("Debug log should contain formatted message")
	}
}

func TestServiceLogger_Debug_WithDebugDisabled(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Ensure debug mode is disabled
	os.Unsetenv("DEBUG")

	logger := NewServiceLogger("TestService")
	logger.Debug("debug message: %s", "details")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug log should not output when DEBUG is not set")
	}

	if len(output) > 0 {
		t.Errorf("Expected no output when DEBUG is disabled, got: %s", output)
	}
}

func TestServiceLogger_MultipleMessages(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logger := NewServiceLogger("MultiService")

	logger.Info("Starting operation")
	logger.Warn("Operation slow")
	logger.Error("Operation failed")

	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "ℹ️") || !strings.Contains(lines[0], "Starting operation") {
		t.Error("First line should be Info log")
	}

	if !strings.Contains(lines[1], "⚠️") || !strings.Contains(lines[1], "Operation slow") {
		t.Error("Second line should be Warn log")
	}

	if !strings.Contains(lines[2], "❌") || !strings.Contains(lines[2], "Operation failed") {
		t.Error("Third line should be Error log")
	}
}
