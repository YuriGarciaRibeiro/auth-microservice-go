package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation_error"
	ErrorTypeAuthentication ErrorType = "authentication_error"
	ErrorTypeAuthorization  ErrorType = "authorization_error"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeConflict       ErrorType = "conflict_error"
	ErrorTypeInternal       ErrorType = "internal_error"
	ErrorTypeBadRequest     ErrorType = "bad_request"
)

type APIError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    string    `json:"code,omitempty"`
	Details any       `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}

func WriteError(w http.ResponseWriter, statusCode int, errorType ErrorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := APIError{
		Type:    errorType,
		Message: message,
	}

	json.NewEncoder(w).Encode(err)
}

func WriteErrorWithDetails(w http.ResponseWriter, statusCode int, errorType ErrorType, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := APIError{
		Type:    errorType,
		Message: message,
		Details: details,
	}

	json.NewEncoder(w).Encode(err)
}

// Common error responses
func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, ErrorTypeBadRequest, message)
}

func ValidationError(w http.ResponseWriter, message string, details any) {
	WriteErrorWithDetails(w, http.StatusUnprocessableEntity, ErrorTypeValidation, message, details)
}

func Unauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, ErrorTypeAuthentication, message)
}

func Forbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, ErrorTypeAuthorization, message)
}

func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, ErrorTypeNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, ErrorTypeConflict, message)
}

func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, ErrorTypeInternal, message)
}
