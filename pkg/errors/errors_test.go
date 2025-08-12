package errors

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("test message", errors.New("inner error"))

	if err.Error() != "validation: test message (inner error)" {
		t.Errorf("Expected 'validation: test message (inner error)', got '%s'", err.Error())
	}

	if err.Unwrap().Error() != "inner error" {
		t.Errorf("Expected 'inner error', got '%s'", err.Unwrap().Error())
	}
}

func TestConfigError(t *testing.T) {
	err := NewConfigError("config issue", nil)

	if err.Error() != "config: config issue" {
		t.Errorf("Expected 'config: config issue', got '%s'", err.Error())
	}
}

func TestNetworkError(t *testing.T) {
	err := NewNetworkError("network issue", errors.New("connection failed"))

	if err.Error() != "network: network issue (connection failed)" {
		t.Errorf("Expected 'network: network issue (connection failed)', got '%s'", err.Error())
	}
}

func TestProcessingError(t *testing.T) {
	err := NewProcessingError("processing failed", nil)

	if err.Error() != "processing: processing failed" {
		t.Errorf("Expected 'processing: processing failed', got '%s'", err.Error())
	}
}
