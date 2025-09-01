package proxyconfig

import (
	"errors"
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

	// 令牌管理
	AddToken(configID string, token *AccessToken) error
	UpdateToken(configID, tokenID string, token *AccessToken) error
	DeleteToken(configID, tokenID string) error
	GetTokens(configID string) ([]AccessToken, error)
	GetTokenByID(configID, tokenID string) (*AccessToken, error)
	ValidateToken(configID, tokenValue string) (*TokenValidationResult, error)
	UpdateTokenUsage(configID, tokenValue string) error
	GetTokenStats(configID string) (*TokenStats, error)
	FindConfigByToken(tokenValue string) (string, error)
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	configs    map[string]*ProxyConfig
	mutex      sync.RWMutex
	maxEntries int
}

// NewMemoryStorage 创建内存存储实例
func NewMemoryStorage(maxEntries int) *MemoryStorage {
	return &MemoryStorage{
		configs:    make(map[string]*ProxyConfig),
		maxEntries: maxEntries,
	}
}

// Add 添加配置
func (s *MemoryStorage) Add(config *ProxyConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查是否超过最大条目数
	if len(s.configs) >= s.maxEntries {
		return fmt.Errorf("maximum entries (%d) exceeded", s.maxEntries)
	}

	// 生成ID和时间戳
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// 初始化令牌相关字段
	if config.AccessTokens == nil {
		config.AccessTokens = make([]AccessToken, 0)
	}
	if config.TokenStats == nil {
		config.TokenStats = &TokenStats{}
	}

	// 存储配置
	s.configs[config.ID] = config

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

	// 更新配置，保留令牌数据
	config.ID = id
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()

	// 保留原有的令牌数据和统计信息
	config.AccessTokens = existing.AccessTokens
	config.TokenStats = existing.TokenStats

	s.configs[id] = config

	return nil
}

// Delete 删除配置
func (s *MemoryStorage) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.configs[id]
	if !exists {
		return ErrConfigNotFound
	}

	delete(s.configs, id)

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

// ==================== 令牌管理方法 ====================

// AddToken 添加令牌到指定配置
func (s *MemoryStorage) AddToken(configID string, token *AccessToken) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[configID]
	if !exists {
		return ErrConfigNotFound
	}

	// 验证令牌数据
	if err := token.Validate(); err != nil {
		return err
	}

	// 检查令牌数量限制
	if err := ValidateTokenLimit(len(config.AccessTokens)); err != nil {
		return err
	}

	// 检查令牌名称是否重复
	for _, existingToken := range config.AccessTokens {
		if existingToken.Name == token.Name {
			return errors.New("token name already exists")
		}
	}

	// 添加令牌
	config.AccessTokens = append(config.AccessTokens, *token)
	config.UpdatedAt = time.Now()

	// 更新令牌统计
	s.updateTokenStatsLocked(config)

	return nil
}

// UpdateToken 更新指定令牌
func (s *MemoryStorage) UpdateToken(configID, tokenID string, token *AccessToken) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[configID]
	if !exists {
		return ErrConfigNotFound
	}

	// 查找令牌
	tokenIndex := -1
	for i, existingToken := range config.AccessTokens {
		if existingToken.ID == tokenID {
			tokenIndex = i
			break
		}
	}

	if tokenIndex == -1 {
		return ErrTokenNotFound
	}

	// 验证令牌数据
	if err := token.Validate(); err != nil {
		return err
	}

	// 检查名称冲突（排除自己）
	for i, existingToken := range config.AccessTokens {
		if i != tokenIndex && existingToken.Name == token.Name {
			return errors.New("token name already exists")
		}
	}

	// 更新令牌
	config.AccessTokens[tokenIndex] = *token
	config.UpdatedAt = time.Now()

	// 更新令牌统计
	s.updateTokenStatsLocked(config)

	return nil
}

// DeleteToken 删除指定令牌
func (s *MemoryStorage) DeleteToken(configID, tokenID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[configID]
	if !exists {
		return ErrConfigNotFound
	}

	// 查找令牌
	tokenIndex := -1
	for i, token := range config.AccessTokens {
		if token.ID == tokenID {
			tokenIndex = i
			break
		}
	}

	if tokenIndex == -1 {
		return ErrTokenNotFound
	}

	// 删除令牌
	config.AccessTokens = append(config.AccessTokens[:tokenIndex], config.AccessTokens[tokenIndex+1:]...)
	config.UpdatedAt = time.Now()

	// 更新令牌统计
	s.updateTokenStatsLocked(config)

	return nil
}

// GetTokens 获取指定配置的所有令牌
func (s *MemoryStorage) GetTokens(configID string) ([]AccessToken, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[configID]
	if !exists {
		return nil, ErrConfigNotFound
	}

	// 返回副本
	tokens := make([]AccessToken, len(config.AccessTokens))
	copy(tokens, config.AccessTokens)

	return tokens, nil
}

// GetTokenByID 根据ID获取指定令牌
func (s *MemoryStorage) GetTokenByID(configID, tokenID string) (*AccessToken, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[configID]
	if !exists {
		return nil, ErrConfigNotFound
	}

	// 查找令牌
	for _, token := range config.AccessTokens {
		if token.ID == tokenID {
			// 返回副本
			tokenCopy := token
			return &tokenCopy, nil
		}
	}

	return nil, ErrTokenNotFound
}

// ValidateToken 验证令牌并返回验证结果
func (s *MemoryStorage) ValidateToken(configID, tokenValue string) (*TokenValidationResult, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[configID]
	if !exists {
		return &TokenValidationResult{
			Valid:     false,
			ErrorCode: "CONFIG_NOT_FOUND",
			ErrorMsg:  "configuration not found",
		}, nil
	}

	// 计算令牌哈希
	tokenHash := HashToken(tokenValue)

	// 查找匹配的令牌
	for _, token := range config.AccessTokens {
		if token.TokenHash == tokenHash {
			// 验证令牌访问权限
			tokenCopy := token // 创建副本避免指针问题
			if err := ValidateTokenAccess(&tokenCopy); err != nil {
				return &TokenValidationResult{
					Valid:     false,
					Token:     &tokenCopy,
					ConfigID:  configID,
					ErrorCode: getErrorCode(err),
					ErrorMsg:  err.Error(),
				}, nil
			}

			// 令牌有效
			return &TokenValidationResult{
				Valid:    true,
				Token:    &tokenCopy,
				ConfigID: configID,
			}, nil
		}
	}

	// 令牌未找到
	return &TokenValidationResult{
		Valid:     false,
		ErrorCode: "TOKEN_NOT_FOUND",
		ErrorMsg:  "token not found",
	}, nil
}

// FindConfigByToken 通过令牌值查找对应的配置ID
func (s *MemoryStorage) FindConfigByToken(tokenValue string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 计算令牌哈希
	tokenHash := HashToken(tokenValue)

	// 遍历所有配置查找匹配的令牌
	for configID, config := range s.configs {
		for _, token := range config.AccessTokens {
			if token.TokenHash == tokenHash {
				// 验证令牌是否有效
				tokenCopy := token
				if err := ValidateTokenAccess(&tokenCopy); err != nil {
					continue // 跳过无效令牌
				}
				return configID, nil
			}
		}
	}

	return "", ErrTokenNotFound
}

// UpdateTokenUsage 更新令牌使用统计
func (s *MemoryStorage) UpdateTokenUsage(configID, tokenValue string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	config, exists := s.configs[configID]
	if !exists {
		return ErrConfigNotFound
	}

	// 计算令牌哈希
	tokenHash := HashToken(tokenValue)

	// 查找并更新令牌
	for i, token := range config.AccessTokens {
		if token.TokenHash == tokenHash {
			config.AccessTokens[i].UpdateUsage()
			config.UpdatedAt = time.Now()

			// 更新令牌统计
			s.updateTokenStatsLocked(config)

			return nil
		}
	}

	return ErrTokenNotFound
}

// GetTokenStats 获取令牌统计信息
func (s *MemoryStorage) GetTokenStats(configID string) (*TokenStats, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	config, exists := s.configs[configID]
	if !exists {
		return nil, ErrConfigNotFound
	}

	if config.TokenStats == nil {
		return &TokenStats{}, nil
	}

	// 返回副本
	statsCopy := *config.TokenStats
	return &statsCopy, nil
}

// updateTokenStatsLocked 更新令牌统计信息（需要持有锁）
func (s *MemoryStorage) updateTokenStatsLocked(config *ProxyConfig) {
	stats := CalculateTokenStats(config.AccessTokens)
	config.TokenStats = stats
}

// getErrorCode 根据错误类型返回错误代码
func getErrorCode(err error) string {
	switch err {
	case ErrTokenNotFound:
		return "TOKEN_NOT_FOUND"
	case ErrTokenExpired:
		return "TOKEN_EXPIRED"
	case ErrTokenDisabled:
		return "TOKEN_DISABLED"
	case ErrTokenInvalid:
		return "TOKEN_INVALID"
	default:
		return "UNKNOWN_ERROR"
	}
}
