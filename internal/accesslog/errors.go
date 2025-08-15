package accesslog

import "errors"

// 日志记录相关错误定义
var (
	// 数据验证错误
	ErrInvalidLogID      = errors.New("invalid log ID")
	ErrInvalidTimestamp  = errors.New("invalid timestamp")
	ErrInvalidMethod     = errors.New("invalid HTTP method")
	ErrInvalidTargetHost = errors.New("invalid target host")
	ErrInvalidStatusCode = errors.New("invalid status code")
	ErrInvalidTimeRange  = errors.New("invalid time range: from time must be before to time")

	// 存储相关错误
	ErrStorageFull       = errors.New("storage is full")
	ErrLogNotFound       = errors.New("log not found")
	ErrStorageNotReady   = errors.New("storage is not ready")
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")

	// 认证相关错误
	ErrInvalidSecret     = errors.New("invalid secret")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrSecretRequired    = errors.New("secret is required")

	// 筛选相关错误
	ErrInvalidFilter     = errors.New("invalid filter parameters")
	ErrInvalidPageNumber = errors.New("invalid page number")
	ErrInvalidLimit      = errors.New("invalid limit value")
)
