package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrCodeUnauthorized, "test message", http.StatusUnauthorized)
	
	if err.Code != ErrCodeUnauthorized {
		t.Errorf("Expected code %s, got %s", ErrCodeUnauthorized, err.Code)
	}
	
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}
	
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, err.StatusCode)
	}
}

func TestAppErrorWithDetails(t *testing.T) {
	err := NewAppError(ErrCodeValidationFailed, "validation failed", http.StatusBadRequest).
		WithDetail("field", "email").
		WithDetail("reason", "invalid format")
	
	if err.Details["field"] != "email" {
		t.Errorf("Expected field detail 'email', got '%v'", err.Details["field"])
	}
	
	if err.Details["reason"] != "invalid format" {
		t.Errorf("Expected reason detail 'invalid format', got '%v'", err.Details["reason"])
	}
}

func TestAppErrorWithCause(t *testing.T) {
	cause := errors.New("original error")
	err := NewAppErrorWithCause(ErrCodeInternalError, "internal error", http.StatusInternalServerError, cause)
	
	if err.Cause != cause {
		t.Errorf("Expected cause to be set")
	}
	
	if err.Unwrap() != cause {
		t.Errorf("Expected Unwrap to return cause")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		code     ErrorCode
		status   int
	}{
		{
			name:   "ErrUnauthorized",
			err:    ErrUnauthorized("test"),
			code:   ErrCodeUnauthorized,
			status: http.StatusUnauthorized,
		},
		{
			name:   "ErrTokenExpired",
			err:    ErrTokenExpired(),
			code:   ErrCodeTokenExpired,
			status: http.StatusUnauthorized,
		},
		{
			name:   "ErrNotFound",
			err:    ErrNotFound("user"),
			code:   ErrCodeNotFound,
			status: http.StatusNotFound,
		},
		{
			name:   "ErrValidationFailed",
			err:    ErrValidationFailed("email", "invalid format"),
			code:   ErrCodeValidationFailed,
			status: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, tt.err.Code)
			}
			
			if tt.err.StatusCode != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, tt.err.StatusCode)
			}
		})
	}
}

func TestIsAppError(t *testing.T) {
	appErr := NewAppError(ErrCodeUnauthorized, "test", http.StatusUnauthorized)
	stdErr := errors.New("standard error")
	
	// 测试AppError
	if err, ok := IsAppError(appErr); !ok || err != appErr {
		t.Errorf("Expected IsAppError to return true for AppError")
	}
	
	// 测试标准错误
	if _, ok := IsAppError(stdErr); ok {
		t.Errorf("Expected IsAppError to return false for standard error")
	}
}

func TestGetStatusCode(t *testing.T) {
	appErr := NewAppError(ErrCodeUnauthorized, "test", http.StatusUnauthorized)
	stdErr := errors.New("standard error")
	
	// 测试AppError
	if status := GetStatusCode(appErr); status != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, status)
	}
	
	// 测试标准错误
	if status := GetStatusCode(stdErr); status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestErrorResponse(t *testing.T) {
	err := NewAppError(ErrCodeValidationFailed, "validation failed", http.StatusBadRequest).
		WithDetail("field", "email")
	
	resp := err.ToErrorResponse()
	
	if resp.Success != false {
		t.Errorf("Expected success to be false")
	}
	
	if resp.Error != "validation failed" {
		t.Errorf("Expected error message 'validation failed', got '%s'", resp.Error)
	}
	
	if resp.ErrorCode != ErrCodeValidationFailed {
		t.Errorf("Expected error code %s, got %s", ErrCodeValidationFailed, resp.ErrorCode)
	}
	
	if resp.Details["field"] != "email" {
		t.Errorf("Expected field detail 'email', got '%v'", resp.Details["field"])
	}
	
	if resp.Timestamp == "" {
		t.Errorf("Expected timestamp to be set")
	}
}

func TestErrorString(t *testing.T) {
	// 测试无原因的错误
	err1 := NewAppError(ErrCodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	expected1 := "UNAUTHORIZED: unauthorized"
	if err1.Error() != expected1 {
		t.Errorf("Expected error string '%s', got '%s'", expected1, err1.Error())
	}
	
	// 测试有原因的错误
	cause := errors.New("connection failed")
	err2 := NewAppErrorWithCause(ErrCodeInternalError, "internal error", http.StatusInternalServerError, cause)
	expected2 := "INTERNAL_ERROR: internal error (caused by: connection failed)"
	if err2.Error() != expected2 {
		t.Errorf("Expected error string '%s', got '%s'", expected2, err2.Error())
	}
}
