package proxyconfig

import (
	"errors"
	"time"
)

// ProxyConfig 代理配置结构
type ProxyConfig struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Subdomain string       `json:"subdomain"`
	TargetURL string       `json:"target_url"`
	Protocol  string       `json:"protocol"`
	Enabled   bool         `json:"enabled"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Stats     *ConfigStats `json:"stats,omitempty"`
}

// ConfigStats 配置访问统计
type ConfigStats struct {
	RequestCount    int64     `json:"request_count"`     // 请求总数
	ErrorCount      int64     `json:"error_count"`       // 错误总数
	AvgResponseTime float64   `json:"avg_response_time"` // 平均响应时间(毫秒)
	LastAccessed    time.Time `json:"last_accessed"`     // 最后访问时间
	TotalBytes      int64     `json:"total_bytes"`       // 总传输字节数
}

// ConfigFilter 配置筛选条件
type ConfigFilter struct {
	Search  string `json:"search"`
	Enabled *bool  `json:"enabled"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

// ConfigResponse 配置列表响应
type ConfigResponse struct {
	Configs    []ProxyConfig `json:"configs"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// StorageStats 存储统计信息
type StorageStats struct {
	TotalConfigs   int `json:"total_configs"`
	EnabledConfigs int `json:"enabled_configs"`
	MemoryUsage    int `json:"memory_usage"`
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	Operation string   `json:"operation"` // enable, disable, delete
	ConfigIDs []string `json:"config_ids"`
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	Success     []string `json:"success"`      // 成功的配置ID
	Failed      []string `json:"failed"`       // 失败的配置ID
	FailedCount int      `json:"failed_count"` // 失败数量
	TotalCount  int      `json:"total_count"`  // 总数量
}

// ExportData 导出数据结构
type ExportData struct {
	Version    string        `json:"version"`
	ExportAt   time.Time     `json:"export_at"`
	Configs    []ProxyConfig `json:"configs"`
	TotalCount int           `json:"total_count"`
}

// ImportResult 导入结果
type ImportResult struct {
	ImportedCount int      `json:"imported_count"` // 导入成功数量
	SkippedCount  int      `json:"skipped_count"`  // 跳过数量
	ErrorCount    int      `json:"error_count"`    // 错误数量
	Errors        []string `json:"errors"`         // 错误信息列表
}

// 错误定义
var (
	ErrConfigNotFound     = errors.New("config not found")
	ErrInvalidConfigID    = errors.New("invalid config id")
	ErrDuplicateSubdomain = errors.New("subdomain already exists")
	ErrInvalidSubdomain   = errors.New("invalid subdomain format")
	ErrInvalidTargetURL   = errors.New("invalid target url")
	ErrMaxEntriesExceeded = errors.New("maximum entries exceeded")
)
