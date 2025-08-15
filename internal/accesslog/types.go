package accesslog

import (
	"time"
)

// AccessLog 访问日志记录结构
type AccessLog struct {
	ID             string            `json:"id"`                        // 唯一标识符
	Timestamp      time.Time         `json:"timestamp"`                 // 请求时间戳
	Method         string            `json:"method"`                    // HTTP 方法
	TargetHost     string            `json:"target_host"`               // 目标主机
	TargetPath     string            `json:"target_path"`               // 目标路径
	StatusCode     int               `json:"status_code"`               // HTTP 状态码
	ResponseBody   string            `json:"response_body,omitempty"`   // 响应内容（仅非200状态码）
	UserAgent      string            `json:"user_agent,omitempty"`      // 发送给目标服务器的User-Agent
	ProxyInfo      string            `json:"proxy_info,omitempty"`      // 代理服务器信息
	ClientIP       string            `json:"client_ip,omitempty"`       // 客户端IP
	Duration       int64             `json:"duration_ms"`               // 请求处理时长（毫秒）
	RequestSize    int64             `json:"request_size,omitempty"`    // 请求大小（字节）
	ResponseSize   int64             `json:"response_size,omitempty"`   // 响应大小（字节）
	RequestHeaders map[string]string `json:"request_headers,omitempty"` // 请求头信息
	RequestBody    string            `json:"request_body,omitempty"`    // 请求体内容
}

// LogFilter 日志筛选条件
type LogFilter struct {
	Domain     string    `json:"domain,omitempty"`      // 域名筛选
	StatusCode []int     `json:"status_code,omitempty"` // 状态码筛选
	FromTime   time.Time `json:"from_time,omitempty"`   // 开始时间
	ToTime     time.Time `json:"to_time,omitempty"`     // 结束时间
	Page       int       `json:"page"`                  // 页码（从1开始）
	Limit      int       `json:"limit"`                 // 每页条数
}

// LogResponse 日志查询响应
type LogResponse struct {
	Logs       []AccessLog `json:"logs"`        // 日志列表
	Total      int         `json:"total"`       // 总条数
	Page       int         `json:"page"`        // 当前页码
	Limit      int         `json:"limit"`       // 每页条数
	TotalPages int         `json:"total_pages"` // 总页数
}

// StorageStats 存储统计信息
type StorageStats struct {
	CurrentEntries int     `json:"current_entries"` // 当前日志条数
	MaxEntries     int     `json:"max_entries"`     // 最大日志条数
	MemoryUsageMB  float64 `json:"memory_usage_mb"` // 内存使用量（MB）
	CleanupCount   int64   `json:"cleanup_count"`   // 清理次数
	LastCleanup    string  `json:"last_cleanup"`    // 最后清理时间
	OldestEntry    string  `json:"oldest_entry"`    // 最老日志时间
	NewestEntry    string  `json:"newest_entry"`    // 最新日志时间
}

// LogLevel 日志级别枚举
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// IsErrorStatus 判断是否为错误状态码
func (log *AccessLog) IsErrorStatus() bool {
	return log.StatusCode >= 400
}

// IsSuccessStatus 判断是否为成功状态码
func (log *AccessLog) IsSuccessStatus() bool {
	return log.StatusCode >= 200 && log.StatusCode < 300
}

// ShouldRecordBody 判断是否应该记录响应体
func (log *AccessLog) ShouldRecordBody() bool {
	return log.StatusCode != 200
}

// GetStatusCategory 获取状态码分类
func (log *AccessLog) GetStatusCategory() string {
	switch {
	case log.StatusCode >= 200 && log.StatusCode < 300:
		return "2xx"
	case log.StatusCode >= 300 && log.StatusCode < 400:
		return "3xx"
	case log.StatusCode >= 400 && log.StatusCode < 500:
		return "4xx"
	case log.StatusCode >= 500:
		return "5xx"
	default:
		return "other"
	}
}

// Validate 验证日志记录的有效性
func (log *AccessLog) Validate() error {
	if log.ID == "" {
		return ErrInvalidLogID
	}
	if log.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}
	if log.Method == "" {
		return ErrInvalidMethod
	}
	if log.TargetHost == "" {
		return ErrInvalidTargetHost
	}
	if log.StatusCode < 100 || log.StatusCode > 599 {
		return ErrInvalidStatusCode
	}
	return nil
}

// Validate 验证筛选条件的有效性
func (filter *LogFilter) Validate() error {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}
	if !filter.FromTime.IsZero() && !filter.ToTime.IsZero() && filter.FromTime.After(filter.ToTime) {
		return ErrInvalidTimeRange
	}
	return nil
}

// SetDefaults 设置筛选条件的默认值
func (filter *LogFilter) SetDefaults() {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}
}
