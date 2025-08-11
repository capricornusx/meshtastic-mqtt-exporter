package errors

import "fmt"

type ErrorType string

const (
	ValidationError ErrorType = "validation"
	ConfigError     ErrorType = "config"
	NetworkError    ErrorType = "network"
	ProcessingError ErrorType = "processing"
)

type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewValidationError(message string, cause error) *AppError {
	return &AppError{Type: ValidationError, Message: message, Cause: cause}
}

func NewConfigError(message string, cause error) *AppError {
	return &AppError{Type: ConfigError, Message: message, Cause: cause}
}

func NewNetworkError(message string, cause error) *AppError {
	return &AppError{Type: NetworkError, Message: message, Cause: cause}
}

func NewProcessingError(message string, cause error) *AppError {
	return &AppError{Type: ProcessingError, Message: message, Cause: cause}
}
