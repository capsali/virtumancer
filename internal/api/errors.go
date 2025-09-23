package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/capsali/virtumancer/internal/logging"
	"github.com/go-chi/chi/v5"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeConflict     ErrorCode = "CONFLICT"
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrorCodeRateLimit    ErrorCode = "RATE_LIMIT"

	// Server errors (5xx)
	ErrorCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavail ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeTimeout        ErrorCode = "TIMEOUT"
	ErrorCodeDependency     ErrorCode = "DEPENDENCY_ERROR"

	// Business logic errors
	ErrorCodeHostNotFound     ErrorCode = "HOST_NOT_FOUND"
	ErrorCodeHostDisconnected ErrorCode = "HOST_DISCONNECTED"
	ErrorCodeVMNotFound       ErrorCode = "VM_NOT_FOUND"
	ErrorCodeVMBusy           ErrorCode = "VM_BUSY"
	ErrorCodeVMStateError     ErrorCode = "VM_STATE_ERROR"
	ErrorCodeLibvirtError     ErrorCode = "LIBVIRT_ERROR"
	ErrorCodeDatabaseError    ErrorCode = "DATABASE_ERROR"
	ErrorCodeConfigError      ErrorCode = "CONFIG_ERROR"
)

// APIError represents a structured API error response
type APIError struct {
	Code      ErrorCode   `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new structured API error
func NewAPIError(code ErrorCode, message string, details interface{}) *APIError {
	return &APIError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// WriteError writes a structured error response
func WriteError(w http.ResponseWriter, err *APIError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
		log.Verbosef("Failed to encode error response: %v", encodeErr)
	}

	// Log the error for monitoring
	log.Verbosef("API Error [%d]: %s - %s", statusCode, err.Code, err.Message)
}

// HandleError provides centralized error handling with proper status codes and structured responses
func (h *APIHandler) HandleError(w http.ResponseWriter, err error, operation string) {
	if err == nil {
		return
	}

	var apiErr *APIError
	var statusCode int

	// Convert different error types to structured API errors
	switch {
	case strings.Contains(strings.ToLower(err.Error()), "not found"):
		if strings.Contains(strings.ToLower(err.Error()), "host") {
			apiErr = NewAPIError(ErrorCodeHostNotFound, "Host not found", err.Error())
			statusCode = http.StatusNotFound
		} else if strings.Contains(strings.ToLower(err.Error()), "vm") || strings.Contains(strings.ToLower(err.Error()), "domain") {
			apiErr = NewAPIError(ErrorCodeVMNotFound, "Virtual machine not found", err.Error())
			statusCode = http.StatusNotFound
		} else {
			apiErr = NewAPIError(ErrorCodeNotFound, "Resource not found", err.Error())
			statusCode = http.StatusNotFound
		}

	case strings.Contains(strings.ToLower(err.Error()), "disconnected") || strings.Contains(strings.ToLower(err.Error()), "connection"):
		apiErr = NewAPIError(ErrorCodeHostDisconnected, "Host is disconnected or unreachable", err.Error())
		statusCode = http.StatusServiceUnavailable

	case strings.Contains(strings.ToLower(err.Error()), "timeout"):
		apiErr = NewAPIError(ErrorCodeTimeout, "Operation timed out", err.Error())
		statusCode = http.StatusGatewayTimeout

	case strings.Contains(strings.ToLower(err.Error()), "libvirt") || strings.Contains(strings.ToLower(err.Error()), "domain"):
		apiErr = NewAPIError(ErrorCodeLibvirtError, "Libvirt operation failed", err.Error())
		statusCode = http.StatusInternalServerError

	case strings.Contains(strings.ToLower(err.Error()), "database") || strings.Contains(strings.ToLower(err.Error()), "sql"):
		apiErr = NewAPIError(ErrorCodeDatabaseError, "Database operation failed", "An internal database error occurred")
		statusCode = http.StatusInternalServerError

	case strings.Contains(strings.ToLower(err.Error()), "busy") || strings.Contains(strings.ToLower(err.Error()), "in use"):
		apiErr = NewAPIError(ErrorCodeVMBusy, "Resource is busy", err.Error())
		statusCode = http.StatusConflict

	case strings.Contains(strings.ToLower(err.Error()), "invalid") || strings.Contains(strings.ToLower(err.Error()), "bad"):
		apiErr = NewAPIError(ErrorCodeBadRequest, "Invalid request", err.Error())
		statusCode = http.StatusBadRequest

	default:
		// Generic internal server error
		apiErr = NewAPIError(ErrorCodeInternal, "Internal server error", "An unexpected error occurred")
		statusCode = http.StatusInternalServerError
	}

	// Add operation context
	if operation != "" {
		apiErr.Details = map[string]interface{}{
			"operation": operation,
			"error":     err.Error(),
		}
	}

	WriteError(w, apiErr, statusCode)
}

// ValidateRequest provides common request validation
func (h *APIHandler) ValidateRequest(w http.ResponseWriter, r *http.Request, requiredParams ...string) bool {
	// Check required URL parameters
	for _, param := range requiredParams {
		if value := chi.URLParam(r, param); value == "" {
			err := NewAPIError(ErrorCodeValidation, fmt.Sprintf("Missing required parameter: %s", param), nil)
			WriteError(w, err, http.StatusBadRequest)
			return false
		}
	}
	return true
}

// Recovery middleware for panic recovery
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Verbosef("Panic recovered: %v", err)

				apiErr := NewAPIError(ErrorCodeInternal, "Internal server error", "A panic occurred while processing the request")
				WriteError(w, apiErr, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
