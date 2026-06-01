// Package httputil provides shared HTTP response helpers for the analytics API.
// Includes JSON response writing with NaN/Inf sanitization, error classification
// (validation vs internal), and a standard error response format.
package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"reflect"

	"github.com/rs/zerolog"
)

// ValidationError represents a client-side error (400).
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// NewValidationError creates a new ValidationError.
func NewValidationError(msg string) *ValidationError {
	return &ValidationError{Message: msg}
}

// HTTPError represents an error with an explicit HTTP status code.
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string { return e.Message }

func NewHTTPError(statusCode int, msg string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Message: msg}
}

func NewUnauthorizedError(msg string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, msg)
}

func NewForbiddenError(msg string) *HTTPError {
	return NewHTTPError(http.StatusForbidden, msg)
}

func NewBadRequestError(msg string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, msg)
}

func NewInternalError(msg string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, msg)
}

// ErrorResponse is the standard error JSON envelope.
type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON response with proper status code and NaN-safe encoding.
func WriteJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	SanitizeFloats(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

// WriteStatusError writes a JSON error response for the given HTTP status and message.
func WriteStatusError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, ErrorResponse{
		Status:  false,
		Message: message,
	})
}

// WriteError writes an error response with the appropriate HTTP status code.
func WriteError(w http.ResponseWriter, logger zerolog.Logger, err error) {
	var validationErr *ValidationError
	var httpErr *HTTPError

	switch {
	case errors.As(err, &validationErr):
		logger.Warn().Err(err).Int("status_code", http.StatusBadRequest).Msg("request failed")
		WriteStatusError(w, http.StatusBadRequest, validationErr.Message)
	case errors.As(err, &httpErr):
		if httpErr.StatusCode >= http.StatusInternalServerError {
			logger.Error().Err(err).Msg("request failed")
		} else {
			logger.Warn().Err(err).Int("status_code", httpErr.StatusCode).Msg("request failed")
		}
		WriteStatusError(w, httpErr.StatusCode, httpErr.Message)
	default:
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		logger.Error().Err(err).Msg("request failed")
		WriteStatusError(w, http.StatusInternalServerError, "internal server error")
	}
}

// SanitizeFloats recursively replaces NaN and Inf float64/float32 values with 0
// in struct fields. It handles structs, pointers, maps, and slices.
func SanitizeFloats(v interface{}) {
	if v == nil {
		return
	}
	sanitizeValue(reflect.ValueOf(v))
}

func sanitizeFloats(v interface{}) {
	SanitizeFloats(v)
}

func sanitizeValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			sanitizeValue(v.Elem())
		}
		return v
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			sanitizeValue(v.Field(i))
		}
		return v
	case reflect.Map:
		for _, key := range v.MapKeys() {
			elem := v.MapIndex(key)
			sanitizeValue(elem)
			v.SetMapIndex(key, sanitizeForMap(elem))
		}
		return v
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			sanitizeValue(v.Index(i))
		}
		return v
	case reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			if v.CanSet() {
				v.SetFloat(0)
			}
			return reflect.Zero(v.Type())
		}
		return v
	case reflect.Float32:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			if v.CanSet() {
				v.SetFloat(0)
			}
			return reflect.Zero(v.Type())
		}
		return v
	case reflect.Interface:
		if !v.IsNil() {
			sanitized := sanitizeValue(v.Elem())
			if v.CanSet() && sanitized.IsValid() {
				v.Set(sanitized)
			}
			return sanitized
		}
		return v
	}

	return v
}

func sanitizeForMap(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		return sanitizeValue(v.Elem())
	default:
		return sanitizeValue(v)
	}
}
