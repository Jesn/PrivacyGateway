package proxyconfig

import (
	"errors"
	"time"
)

// ProxyConfig 代理配置结构
type ProxyConfig struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subdomain string    `json:"subdomain"`
	TargetURL string    `json:"target_url"`
	Protocol  string    `json:"protocol"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

// 错误定义
var (
	ErrConfigNotFound     = errors.New("config not found")
	ErrInvalidConfigID    = errors.New("invalid config id")
	ErrDuplicateSubdomain = errors.New("subdomain already exists")
	ErrInvalidSubdomain   = errors.New("invalid subdomain format")
	ErrInvalidTargetURL   = errors.New("invalid target url")
	ErrMaxEntriesExceeded = errors.New("maximum entries exceeded")
)
