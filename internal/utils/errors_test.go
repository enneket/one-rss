package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(http.StatusBadRequest, "Bad request", nil)

	if err.Code != http.StatusBadRequest {
		t.Errorf("Expected code %d, got %d", http.StatusBadRequest, err.Code)
	}

	if err.Message != "Bad request" {
		t.Errorf("Expected message 'Bad request', got '%s'", err.Message)
	}

	if err.Err != nil {
		t.Error("Expected nil error")
	}
}

func TestNewAppErrorWithWrappedError(t *testing.T) {
	innerErr := errors.New("inner error")
	err := NewAppError(http.StatusInternalServerError, "Internal error", innerErr)

	if err.Code != http.StatusInternalServerError {
		t.Errorf("Expected code %d, got %d", http.StatusInternalServerError, err.Code)
	}

	if err.Err != innerErr {
		t.Error("Expected wrapped error")
	}
}

func TestAppErrorError(t *testing.T) {
	// Test with inner error
	innerErr := errors.New("inner error")
	err := NewAppError(http.StatusInternalServerError, "Internal error", innerErr)

	if err.Error() != "inner error" {
		t.Errorf("Expected 'inner error', got '%s'", err.Error())
	}

	// Test without inner error
	err2 := NewAppError(http.StatusBadRequest, "Bad request", nil)
	if err2.Error() != "Bad request" {
		t.Errorf("Expected 'Bad request', got '%s'", err2.Error())
	}
}

func TestHandleError(t *testing.T) {
	// Test with AppError
	w := httptest.NewRecorder()
	appErr := NewAppError(http.StatusBadRequest, "Bad request", nil)
	HandleError(w, appErr)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Test with generic error
	w2 := httptest.NewRecorder()
	genericErr := errors.New("generic error")
	HandleError(w2, genericErr)

	if w2.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w2.Code)
	}
}

func TestRespondWithJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	RespondWithJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	RespondSuccess(w, "Operation successful")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRespondBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	RespondBadRequest(w, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRespondNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	RespondNotFound(w, "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRespondInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondInternalError(w, "Server error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
