# 第一阶段：核心功能实现 (v1.1.0)

## 🎯 阶段目标

实现基础的代理配置管理功能，支持CRUD操作和基本的子域名代理。

## 📦 功能范围

### 后端功能
- ✅ 代理配置数据结构定义
- ✅ 内存存储实现
- ✅ 配置管理API接口
- ✅ 子域名代理路由处理
- ✅ 配置验证和错误处理

### 前端功能
- ✅ 标签页切换界面
- ✅ 配置列表展示
- ✅ 新增/编辑配置弹框
- ✅ 删除确认功能
- ✅ 基础搜索和筛选

## 🛠️ 技术实现

### 1. 后端开发任务

#### 1.1 数据结构设计
**文件**: `internal/proxyconfig/types.go`

```go
package proxyconfig

import (
    "time"
)

// ProxyConfig 代理配置结构
type ProxyConfig struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Subdomain   string    `json:"subdomain"`
    TargetURL   string    `json:"target_url"`
    Protocol    string    `json:"protocol"`
    Enabled     bool      `json:"enabled"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// ConfigFilter 配置筛选条件
type ConfigFilter struct {
    Search   string `json:"search"`
    Enabled  *bool  `json:"enabled"`
    Page     int    `json:"page"`
    Limit    int    `json:"limit"`
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
    ErrConfigNotFound    = errors.New("config not found")
    ErrInvalidConfigID   = errors.New("invalid config id")
    ErrDuplicateSubdomain = errors.New("subdomain already exists")
    ErrInvalidSubdomain  = errors.New("invalid subdomain format")
    ErrInvalidTargetURL  = errors.New("invalid target url")
)
```

#### 1.2 存储层实现
**文件**: `internal/proxyconfig/storage.go`

```go
package proxyconfig

import (
    "fmt"
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
```

#### 1.3 配置验证
**文件**: `internal/proxyconfig/validator.go`

```go
package proxyconfig

import (
    "net/url"
    "regexp"
    "strings"
)

var (
    // 子域名格式验证正则
    subdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// ValidateConfig 验证配置
func ValidateConfig(config *ProxyConfig) error {
    if config.Name == "" {
        return errors.New("name is required")
    }

    if len(config.Name) > 100 {
        return errors.New("name too long (max 100 characters)")
    }

    if err := ValidateSubdomain(config.Subdomain); err != nil {
        return err
    }

    if err := ValidateTargetURL(config.TargetURL); err != nil {
        return err
    }

    if config.Protocol != "http" && config.Protocol != "https" {
        return errors.New("protocol must be http or https")
    }

    return nil
}

// ValidateSubdomain 验证子域名
func ValidateSubdomain(subdomain string) error {
    if subdomain == "" {
        return errors.New("subdomain is required")
    }

    subdomain = strings.ToLower(subdomain)

    if len(subdomain) < 1 || len(subdomain) > 63 {
        return errors.New("subdomain length must be 1-63 characters")
    }

    if !subdomainRegex.MatchString(subdomain) {
        return errors.New("invalid subdomain format")
    }

    // 检查保留子域名
    reserved := []string{"www", "api", "admin", "mail", "ftp", "localhost"}
    for _, r := range reserved {
        if subdomain == r {
            return errors.New("subdomain is reserved")
        }
    }

    return nil
}

// ValidateTargetURL 验证目标URL
func ValidateTargetURL(targetURL string) error {
    if targetURL == "" {
        return errors.New("target_url is required")
    }

    u, err := url.Parse(targetURL)
    if err != nil {
        return errors.New("invalid target_url format")
    }

    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("target_url must use http or https")
    }

    if u.Host == "" {
        return errors.New("target_url must have a host")
    }

    return nil
}
```

## 📊 验收标准

### 功能验收
- ✅ 能够通过Web界面管理代理配置
- ✅ 支持新增、编辑、删除配置
- ✅ 子域名代理功能正常工作
- ✅ 配置启用/禁用功能正常
- ✅ 搜索和筛选功能正常

### 性能验收
- ✅ 配置列表加载时间 < 500ms
- ✅ 子域名代理响应时间 < 100ms
- ✅ 支持至少1000个配置项
- ✅ 内存使用增长 < 50MB

### 质量验收
- ✅ 单元测试覆盖率 > 80%
- ✅ 所有API接口正常工作
- ✅ 错误处理完善
- ✅ 数据验证严格

## 🧪 测试计划

### 单元测试
```bash
# 运行存储层测试
go test ./internal/proxyconfig -v

# 运行API处理器测试
go test ./internal/handler -v

# 运行验证器测试
go test ./internal/proxyconfig -run TestValidate -v
```

### 集成测试
```bash
# 启动测试服务器
ADMIN_SECRET=test-secret go run . &

# 运行集成测试
go test ./tests/integration -v

# 清理
pkill privacy-gateway
```

#### 1.4 API处理器实现
**文件**: `internal/handler/proxyconfig.go`

```go
package handler

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"

    "privacygateway/internal/config"
    "privacygateway/internal/logger"
    "privacygateway/internal/proxyconfig"
)

// HandleProxyConfigAPI 处理代理配置API请求
func HandleProxyConfigAPI(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger, storage proxyconfig.Storage) {
    // 认证检查
    if !isAuthorizedForConfig(r, cfg.AdminSecret) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // 设置CORS头
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Log-Secret")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    switch r.Method {
    case http.MethodGet:
        handleGetConfigs(w, r, storage, log)
    case http.MethodPost:
        handleCreateConfig(w, r, storage, log)
    case http.MethodPut:
        handleUpdateConfig(w, r, storage, log)
    case http.MethodDelete:
        handleDeleteConfig(w, r, storage, log)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

// isAuthorizedForConfig 检查配置管理权限
func isAuthorizedForConfig(r *http.Request, adminSecret string) bool {
    if adminSecret == "" {
        return false
    }

    // 检查请求头
    if secret := r.Header.Get("X-Log-Secret"); secret == adminSecret {
        return true
    }

    // 检查查询参数
    if secret := r.URL.Query().Get("secret"); secret == adminSecret {
        return true
    }

    return false
}

// handleGetConfigs 获取配置列表
func handleGetConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
    // 解析查询参数
    filter := &proxyconfig.ConfigFilter{
        Search: r.URL.Query().Get("search"),
        Page:   1,
        Limit:  20,
    }

    if pageStr := r.URL.Query().Get("page"); pageStr != "" {
        if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
            filter.Page = page
        }
    }

    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
            filter.Limit = limit
        }
    }

    if enabledStr := r.URL.Query().Get("enabled"); enabledStr != "" {
        if enabled, err := strconv.ParseBool(enabledStr); err == nil {
            filter.Enabled = &enabled
        }
    }

    // 获取配置列表
    response, err := storage.List(filter)
    if err != nil {
        log.Error("failed to get config list", "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // 返回JSON响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// handleCreateConfig 创建配置
func handleCreateConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
    var config proxyconfig.ProxyConfig
    if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 验证配置
    if err := proxyconfig.ValidateConfig(&config); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 添加配置
    if err := storage.Add(&config); err != nil {
        log.Error("failed to add config", "error", err)
        if err == proxyconfig.ErrDuplicateSubdomain {
            http.Error(w, "Subdomain already exists", http.StatusConflict)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    log.Info("config created", "id", config.ID, "name", config.Name, "subdomain", config.Subdomain)

    // 返回创建的配置
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(config)
}

// handleUpdateConfig 更新配置
func handleUpdateConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
    configID := r.URL.Query().Get("id")
    if configID == "" {
        http.Error(w, "Config ID is required", http.StatusBadRequest)
        return
    }

    var config proxyconfig.ProxyConfig
    if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // 验证配置
    if err := proxyconfig.ValidateConfig(&config); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // 更新配置
    if err := storage.Update(configID, &config); err != nil {
        log.Error("failed to update config", "id", configID, "error", err)
        if err == proxyconfig.ErrConfigNotFound {
            http.Error(w, "Config not found", http.StatusNotFound)
        } else if err == proxyconfig.ErrDuplicateSubdomain {
            http.Error(w, "Subdomain already exists", http.StatusConflict)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    log.Info("config updated", "id", configID, "name", config.Name)

    // 返回更新的配置
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(config)
}

// handleDeleteConfig 删除配置
func handleDeleteConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
    configID := r.URL.Query().Get("id")
    if configID == "" {
        http.Error(w, "Config ID is required", http.StatusBadRequest)
        return
    }

    // 删除配置
    if err := storage.Delete(configID); err != nil {
        log.Error("failed to delete config", "id", configID, "error", err)
        if err == proxyconfig.ErrConfigNotFound {
            http.Error(w, "Config not found", http.StatusNotFound)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    log.Info("config deleted", "id", configID)

    w.WriteHeader(http.StatusNoContent)
}
```

## 📈 下一步计划

完成第一阶段后，进入[第二阶段：功能增强](PROXY_CONFIG_PHASE2.md)，主要包括：
- 配置文件持久化存储
- 配置导入/导出功能
- 批量操作支持
- 访问统计和监控

---

**阶段状态**: 开发中
**预计完成**: 15个工作日
**负责人**: 开发团队
