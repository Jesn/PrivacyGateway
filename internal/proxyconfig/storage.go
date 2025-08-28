package proxyconfig

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage 配置存储接口
type Storage interface {
	Add(config *ProxyConfig) error
	Update(id string, config *ProxyConfig) error
	Delete(id string) error
	GetByID(id string) (*ProxyConfig, error)
	List(filter *ConfigFilter) (*ConfigResponse, error)
	GetBySubdomain(subdomain string) (*ProxyConfig, error)
	Clear()
	GetStats() *StorageStats

	// 批量操作
	BatchOperation(operation string, configIDs []string) (*BatchOperationResult, error)

	// 导入导出
	ExportAll() (*ExportData, error)
	ImportConfigs(configs []ProxyConfig, mode string) (*ImportResult, error)

	// 统计功能
	UpdateStats(configID string, responseTime time.Duration, success bool, bytes int64) error
	GetConfigStats(configID string) (*ConfigStats, error)
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	configs    map[string]*ProxyConfig
	subdomains map[string]string // subdomain -> config_id
	mutex      sync.RWMutex
	maxEntries int
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage(maxEntries int) *MemoryStorage {
	return &MemoryStorage{
		configs:    make(map[string]*ProxyConfig),
		subdomains: make(map[string]string),
		maxEntries: maxEntries,
	}
}

// Add 添加配置
func (s *MemoryStorage) Add(config *ProxyConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查子域名是否已存在
	if _, exists := s.subdomains[config.Subdomain]; exists {
		return ErrDuplicateSubdomain
	}

	// 检查是否超过最大条目数
	if len(s.configs) >= s.maxEntries {
		return fmt.Errorf("maximum entries (%d) exceeded", s.maxEntries)
	}

	// 生成ID和时间戳
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// 存储配置
	s.configs[config.ID] = config
	s.subdomains[config.Subdomain] = config.ID

	return nil
}

// Update 更新配置
func (s *MemoryStorage) Update(id string, config *ProxyConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	existing, exists := s.configs[id]
	if !exists {
		return ErrConfigNotFound
	}

	// 检查子域名冲突
	if existing.Subdomain != config.Subdomain {
		if _, exists := s.subdomains[config.Subdomain]; exists {
			return ErrDuplicateSubdomain
		}
		// 更新子域名映射
		delete(s.subdomains, existing.Subdomain)
		s.subdomains[config.Subdomain] = id
	}

	// 更新配置
	config.ID = id
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()
	s.configs[id] = config

	return nil
}

// Delete 删除配置
func (s *MemoryStorage) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[id]
	if !exists {
		return ErrConfigNotFound
	}

	delete(s.configs, id)
	delete(s.subdomains, config.Subdomain)

	return nil
}

// GetByID 根据ID获取配置
func (s *MemoryStorage) GetByID(id string) (*ProxyConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[id]
	if !exists {
		return nil, ErrConfigNotFound
	}

	// 返回副本
	configCopy := *config
	return &configCopy, nil
}

// GetBySubdomain 根据子域名获取配置
func (s *MemoryStorage) GetBySubdomain(subdomain string) (*ProxyConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configID, exists := s.subdomains[subdomain]
	if !exists {
		return nil, ErrConfigNotFound
	}

	config := s.configs[configID]
	if !config.Enabled {
		return nil, ErrConfigNotFound
	}

	// 返回副本
	configCopy := *config
	return &configCopy, nil
}

// List 获取配置列表
func (s *MemoryStorage) List(filter *ConfigFilter) (*ConfigResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var allConfigs []ProxyConfig
	for _, config := range s.configs {
		// 应用筛选条件
		if filter.Search != "" {
			searchTerm := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(config.Name), searchTerm) &&
				!strings.Contains(strings.ToLower(config.Subdomain), searchTerm) &&
				!strings.Contains(strings.ToLower(config.TargetURL), searchTerm) {
				continue
			}
		}

		if filter.Enabled != nil && config.Enabled != *filter.Enabled {
			continue
		}

		allConfigs = append(allConfigs, *config)
	}

	// 排序（按创建时间倒序）
	sort.Slice(allConfigs, func(i, j int) bool {
		return allConfigs[i].CreatedAt.After(allConfigs[j].CreatedAt)
	})

	// 分页
	total := len(allConfigs)
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	start := (filter.Page - 1) * filter.Limit
	end := start + filter.Limit

	if start >= total {
		allConfigs = []ProxyConfig{}
	} else if end > total {
		allConfigs = allConfigs[start:]
	} else {
		allConfigs = allConfigs[start:end]
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit

	return &ConfigResponse{
		Configs:    allConfigs,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// Clear 清空所有配置
func (s *MemoryStorage) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.configs = make(map[string]*ProxyConfig)
	s.subdomains = make(map[string]string)
}

// GetStats 获取统计信息
func (s *MemoryStorage) GetStats() *StorageStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	enabledCount := 0
	for _, config := range s.configs {
		if config.Enabled {
			enabledCount++
		}
	}

	return &StorageStats{
		TotalConfigs:   len(s.configs),
		EnabledConfigs: enabledCount,
		MemoryUsage:    len(s.configs) * 200, // 估算每个配置约200字节
	}
}

// BatchOperation 批量操作
func (s *MemoryStorage) BatchOperation(operation string, configIDs []string) (*BatchOperationResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := &BatchOperationResult{
		Success:    make([]string, 0),
		Failed:     make([]string, 0),
		TotalCount: len(configIDs),
	}

	for _, configID := range configIDs {
		config, exists := s.configs[configID]
		if !exists {
			result.Failed = append(result.Failed, configID)
			continue
		}

		switch operation {
		case "enable":
			config.Enabled = true
			config.UpdatedAt = time.Now()
			result.Success = append(result.Success, configID)
		case "disable":
			config.Enabled = false
			config.UpdatedAt = time.Now()
			result.Success = append(result.Success, configID)
		case "delete":
			delete(s.configs, configID)
			delete(s.subdomains, config.Subdomain)
			result.Success = append(result.Success, configID)
		default:
			result.Failed = append(result.Failed, configID)
		}
	}

	result.FailedCount = len(result.Failed)
	return result, nil
}

// ExportAll 导出所有配置
func (s *MemoryStorage) ExportAll() (*ExportData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	configs := make([]ProxyConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configs = append(configs, *config)
	}

	return &ExportData{
		Version:    "1.0",
		ExportAt:   time.Now(),
		Configs:    configs,
		TotalCount: len(configs),
	}, nil
}

// ImportConfigs 导入配置
func (s *MemoryStorage) ImportConfigs(configs []ProxyConfig, mode string) (*ImportResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := &ImportResult{
		Errors: make([]string, 0),
	}

	for _, config := range configs {
		// 验证配置
		if err := ValidateConfig(&config); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("配置 %s 验证失败: %v", config.Name, err))
			continue
		}

		// 检查子域名冲突
		if existingID, exists := s.subdomains[config.Subdomain]; exists {
			if mode == "skip" {
				result.SkippedCount++
				continue
			} else if mode == "replace" {
				// 删除现有配置
				delete(s.configs, existingID)
			} else {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("子域名 %s 已存在", config.Subdomain))
				continue
			}
		}

		// 检查是否超过最大条目数
		if len(s.configs) >= s.maxEntries {
			result.ErrorCount++
			result.Errors = append(result.Errors, fmt.Sprintf("已达到最大配置数量限制 (%d)", s.maxEntries))
			break
		}

		// 生成新的ID和时间戳
		config.ID = uuid.New().String()
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()

		// 添加配置
		s.configs[config.ID] = &config
		s.subdomains[config.Subdomain] = config.ID
		result.ImportedCount++
	}

	return result, nil
}

// UpdateStats 更新配置统计信息
func (s *MemoryStorage) UpdateStats(configID string, responseTime time.Duration, success bool, bytes int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[configID]
	if !exists {
		return ErrConfigNotFound
	}

	// 初始化统计信息
	if config.Stats == nil {
		config.Stats = &ConfigStats{}
	}

	// 更新统计数据
	config.Stats.RequestCount++
	if !success {
		config.Stats.ErrorCount++
	}

	// 更新平均响应时间（使用移动平均）
	responseTimeMs := float64(responseTime.Nanoseconds()) / 1e6
	if config.Stats.RequestCount == 1 {
		config.Stats.AvgResponseTime = responseTimeMs
	} else {
		// 使用指数移动平均，权重为0.1
		config.Stats.AvgResponseTime = config.Stats.AvgResponseTime*0.9 + responseTimeMs*0.1
	}

	config.Stats.LastAccessed = time.Now()
	config.Stats.TotalBytes += bytes

	return nil
}

// GetConfigStats 获取配置统计信息
func (s *MemoryStorage) GetConfigStats(configID string) (*ConfigStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[configID]
	if !exists {
		return nil, ErrConfigNotFound
	}

	if config.Stats == nil {
		return &ConfigStats{}, nil
	}

	// 返回副本
	statsCopy := *config.Stats
	return &statsCopy, nil
}
