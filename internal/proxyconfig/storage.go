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
