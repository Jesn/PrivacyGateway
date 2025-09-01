// Package errors 提供统一的错误处理和错误类型定义
package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode 错误代码类型
type ErrorCode string

// 预定义的错误代码
const (
	// 认证相关错误
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenDisabled    ErrorCode = "TOKEN_DISABLED"
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// 资源相关错误
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeConfigNotFound    ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeTokenNotFound     ErrorCode = "TOKEN_NOT_FOUND"
	ErrCodeDuplicateResource ErrorCode = "DUPLICATE_RESOURCE"

	// 验证相关错误
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// 系统相关错误
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeResourceExhausted  ErrorCode = "RESOURCE_EXHAUSTED"

	// 业务相关错误
	ErrCodeMaxTokensExceeded ErrorCode = "MAX_TOKENS_EXCEEDED"
	ErrCodeInvalidTarget     ErrorCode = "INVALID_TARGET"
	ErrCodeProxyFailed       ErrorCode = "PROXY_FAILED"
)

// AppError 应用程序错误类型
type AppError struct {
	Code       ErrorCode              `json:"error_code"`
	Message    string                 `json:"error"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Cause      error                  `json:"-"`
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError 创建新的应用程序错误
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// NewAppErrorWithCause 创建带原因的应用程序错误
func NewAppErrorWithCause(code ErrorCode, message string, statusCode int, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
		Cause:      cause,
	}
}

// WithDetail 添加错误详情
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails 批量添加错误详情
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// 预定义的常用错误

// ErrUnauthorized 未授权错误
func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// ErrTokenExpired 令牌过期错误
func ErrTokenExpired() *AppError {
	return NewAppError(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
}

// ErrTokenDisabled 令牌禁用错误
func ErrTokenDisabled() *AppError {
	return NewAppError(ErrCodeTokenDisabled, "Token is disabled", http.StatusUnauthorized)
}

// ErrInvalidToken 无效令牌错误
func ErrInvalidToken() *AppError {
	return NewAppError(ErrCodeInvalidToken, "Invalid token", http.StatusUnauthorized)
}

// ErrPermissionDenied 权限拒绝错误
func ErrPermissionDenied(resource string) *AppError {
	return NewAppError(ErrCodePermissionDenied, "Permission denied", http.StatusForbidden).
		WithDetail("resource", resource)
}

// ErrNotFound 资源不存在错误
func ErrNotFound(resource string) *AppError {
	return NewAppError(ErrCodeNotFound, "Resource not found", http.StatusNotFound).
		WithDetail("resource", resource)
}

// ErrConfigNotFound 配置不存在错误
func ErrConfigNotFound(configID string) *AppError {
	return NewAppError(ErrCodeConfigNotFound, "Configuration not found", http.StatusNotFound).
		WithDetail("config_id", configID)
}

// ErrTokenNotFound 令牌不存在错误
func ErrTokenNotFound(tokenID string) *AppError {
	return NewAppError(ErrCodeTokenNotFound, "Token not found", http.StatusNotFound).
		WithDetail("token_id", tokenID)
}

// ErrValidationFailed 验证失败错误
func ErrValidationFailed(field string, reason string) *AppError {
	return NewAppError(ErrCodeValidationFailed, "Validation failed", http.StatusBadRequest).
		WithDetails(map[string]interface{}{
			"field":  field,
			"reason": reason,
		})
}

// ErrInvalidInput 无效输入错误
func ErrInvalidInput(message string) *AppError {
	if message == "" {
		message = "Invalid input"
	}
	return NewAppError(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

// ErrMissingField 缺少字段错误
func ErrMissingField(field string) *AppError {
	return NewAppError(ErrCodeMissingField, "Required field is missing", http.StatusBadRequest).
		WithDetail("field", field)
}

// ErrInternalError 内部错误
func ErrInternalError(message string, cause error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppErrorWithCause(ErrCodeInternalError, message, http.StatusInternalServerError, cause)
}

// ErrServiceUnavailable 服务不可用错误
func ErrServiceUnavailable(service string) *AppError {
	return NewAppError(ErrCodeServiceUnavailable, "Service temporarily unavailable", http.StatusServiceUnavailable).
		WithDetail("service", service)
}

// ErrRateLimitExceeded 速率限制错误
func ErrRateLimitExceeded() *AppError {
	return NewAppError(ErrCodeRateLimitExceeded, "Rate limit exceeded", http.StatusTooManyRequests)
}

// ErrMaxTokensExceeded 超过最大令牌数量错误
func ErrMaxTokensExceeded(maxTokens int) *AppError {
	return NewAppError(ErrCodeMaxTokensExceeded, "Maximum number of tokens exceeded", http.StatusBadRequest).
		WithDetail("max_tokens", maxTokens)
}

// ErrProxyFailed 代理失败错误
func ErrProxyFailed(target string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeProxyFailed, "Proxy request failed", http.StatusBadGateway, cause).
		WithDetail("target", target)
}

// ErrorResponse API错误响应结构
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	ErrorCode ErrorCode              `json:"error_code"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ToErrorResponse 将AppError转换为ErrorResponse
func (e *AppError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     e.Message,
		ErrorCode: e.Code,
		Details:   e.Details,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}
}

// IsAppError 检查错误是否为AppError类型
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// GetStatusCode 获取错误对应的HTTP状态码
func GetStatusCode(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
