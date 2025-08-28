# 第二阶段：功能增强 (v1.2.0)

## 🎯 阶段目标

在第一阶段基础上，增强配置管理功能，添加持久化存储和高级特性，提升系统的实用性和稳定性。

## 📦 功能范围

### 核心增强功能
- ✅ **配置文件持久化存储** - 支持JSON文件持久化，服务重启后配置不丢失
- ✅ **配置导入/导出功能** - 支持批量配置的导入导出，便于备份和迁移
- ✅ **批量操作支持** - 支持批量启用/禁用、批量删除等操作
- ✅ **配置模板功能** - 预定义常用配置模板，快速创建配置
- ✅ **访问统计和监控** - 统计各配置的访问次数和响应时间

### 高级特性
- ✅ **配置分组管理** - 支持配置分组，便于管理大量配置
- ✅ **路径重写规则** - 支持URL路径重写，更灵活的代理配置
- ✅ **请求头自定义** - 支持自定义请求头，满足特殊需求
- ✅ **负载均衡支持** - 支持多个目标服务器的负载均衡
- ✅ **健康检查功能** - 定期检查目标服务器健康状态

## 🛠️ 技术实现

### 1. 持久化存储实现

#### 数据结构扩展
```go
// internal/proxyconfig/types.go (扩展)
type ProxyConfig struct {
    // ... 现有字段
    Group       string            `json:"group"`        // 配置分组
    Tags        []string          `json:"tags"`         // 标签
    PathRewrite []RewriteRule     `json:"path_rewrite"` // 路径重写规则
    Headers     map[string]string `json:"headers"`      // 自定义请求头
    LoadBalance *LoadBalanceConfig `json:"load_balance"` // 负载均衡配置
    HealthCheck *HealthCheckConfig `json:"health_check"` // 健康检查配置
    Stats       *ConfigStats      `json:"stats"`        // 访问统计
}

type RewriteRule struct {
    From  string `json:"from"`  // 源路径模式
    To    string `json:"to"`    // 目标路径模式
    Regex bool   `json:"regex"` // 是否使用正则表达式
}

type LoadBalanceConfig struct {
    Targets   []string `json:"targets"`   // 目标服务器列表
    Algorithm string   `json:"algorithm"` // 负载均衡算法 (round_robin, random, weighted)
    Weights   []int    `json:"weights"`   // 权重配置
}

type HealthCheckConfig struct {
    Enabled  bool   `json:"enabled"`  // 是否启用健康检查
    Interval int    `json:"interval"` // 检查间隔(秒)
    Timeout  int    `json:"timeout"`  // 超时时间(秒)
    Path     string `json:"path"`     // 健康检查路径
}

type ConfigStats struct {
    RequestCount    int64   `json:"request_count"`    // 请求总数
    ErrorCount      int64   `json:"error_count"`      // 错误总数
    AvgResponseTime float64 `json:"avg_response_time"` // 平均响应时间
    LastAccessed    time.Time `json:"last_accessed"`  // 最后访问时间
}
```

#### 持久化存储实现
```go
// internal/proxyconfig/persistent.go
type PersistentStorage struct {
    *MemoryStorage
    filePath     string
    saveInterval time.Duration
    autoSave     bool
    mutex        sync.RWMutex
}

func NewPersistentStorage(filePath string, maxEntries int, autoSave bool) *PersistentStorage {
    ps := &PersistentStorage{
        MemoryStorage: NewMemoryStorage(maxEntries),
        filePath:      filePath,
        saveInterval:  30 * time.Second,
        autoSave:      autoSave,
    }
    
    // 启动时加载配置
    ps.LoadFromFile()
    
    // 启动自动保存
    if autoSave {
        ps.StartAutoSave()
    }
    
    return ps
}

func (ps *PersistentStorage) SaveToFile() error {
    ps.mutex.RLock()
    defer ps.mutex.RUnlock()
    
    data, err := json.MarshalIndent(ps.configs, "", "  ")
    if err != nil {
        return err
    }
    
    return ioutil.WriteFile(ps.filePath, data, 0644)
}

func (ps *PersistentStorage) LoadFromFile() error {
    if _, err := os.Stat(ps.filePath); os.IsNotExist(err) {
        return nil // 文件不存在，跳过加载
    }
    
    data, err := ioutil.ReadFile(ps.filePath)
    if err != nil {
        return err
    }
    
    var configs map[string]*ProxyConfig
    if err := json.Unmarshal(data, &configs); err != nil {
        return err
    }
    
    ps.mutex.Lock()
    defer ps.mutex.Unlock()
    
    ps.configs = configs
    ps.rebuildSubdomainIndex()
    
    return nil
}

func (ps *PersistentStorage) StartAutoSave() {
    go func() {
        ticker := time.NewTicker(ps.saveInterval)
        defer ticker.Stop()
        
        for range ticker.C {
            if err := ps.SaveToFile(); err != nil {
                // 记录错误日志
                log.Printf("Failed to auto-save configs: %v", err)
            }
        }
    }()
}
```

### 2. 配置导入导出功能

#### API接口扩展
```go
// internal/handler/proxyconfig.go (扩展)

// 导出配置
func handleExportConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage) {
    configs, err := storage.ExportAll()
    if err != nil {
        http.Error(w, "Export failed", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=proxy-configs.json")
    json.NewEncoder(w).Encode(configs)
}

// 导入配置
func handleImportConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage) {
    var configs []proxyconfig.ProxyConfig
    if err := json.NewDecoder(r.Body).Decode(&configs); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    result, err := storage.ImportConfigs(configs, r.URL.Query().Get("mode"))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// 批量操作
func handleBatchOperation(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage) {
    var req struct {
        Operation string   `json:"operation"` // enable, disable, delete
        ConfigIDs []string `json:"config_ids"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    result, err := storage.BatchOperation(req.Operation, req.ConfigIDs)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 3. 配置模板功能

#### 模板定义
```go
// internal/proxyconfig/templates.go
type ConfigTemplate struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Category    string            `json:"category"`
    Template    ProxyConfig       `json:"template"`
    Variables   []TemplateVar     `json:"variables"`
}

type TemplateVar struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Type        string `json:"type"` // string, number, boolean
    Required    bool   `json:"required"`
    Default     string `json:"default"`
}

// 预定义模板
var DefaultTemplates = []ConfigTemplate{
    {
        ID:          "github-api",
        Name:        "GitHub API",
        Description: "GitHub API代理配置",
        Category:    "开发工具",
        Template: ProxyConfig{
            Name:      "GitHub API",
            Subdomain: "github-api",
            TargetURL: "https://api.github.com",
            Protocol:  "https",
            Enabled:   true,
        },
        Variables: []TemplateVar{
            {Name: "subdomain", Description: "子域名", Type: "string", Required: true},
        },
    },
    // 更多模板...
}
```

### 4. 访问统计和监控

#### 统计数据收集
```go
// internal/proxyconfig/stats.go
type StatsCollector struct {
    storage Storage
    mutex   sync.RWMutex
}

func (sc *StatsCollector) RecordRequest(configID string, responseTime time.Duration, success bool) {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()
    
    config, err := sc.storage.GetByID(configID)
    if err != nil {
        return
    }
    
    if config.Stats == nil {
        config.Stats = &ConfigStats{}
    }
    
    config.Stats.RequestCount++
    if !success {
        config.Stats.ErrorCount++
    }
    
    // 更新平均响应时间
    config.Stats.AvgResponseTime = (config.Stats.AvgResponseTime + responseTime.Seconds()) / 2
    config.Stats.LastAccessed = time.Now()
    
    sc.storage.Update(configID, config)
}
```

## 📊 新增API接口

### 配置导入导出
```http
GET /config/proxy/export - 导出所有配置
POST /config/proxy/import - 导入配置
POST /config/proxy/batch - 批量操作
```

### 配置模板
```http
GET /config/templates - 获取模板列表
POST /config/proxy/from-template - 从模板创建配置
```

### 统计信息
```http
GET /config/proxy/stats - 获取统计信息
GET /config/proxy/{id}/stats - 获取单个配置统计
```

## 🔧 环境变量配置

```bash
# 持久化配置
PROXY_CONFIG_PERSIST=true
PROXY_CONFIG_FILE=/data/proxy-configs.json
PROXY_CONFIG_AUTO_SAVE=true
PROXY_CONFIG_SAVE_INTERVAL=30s

# 高级功能
PROXY_CONFIG_MAX_ENTRIES=5000
PROXY_CONFIG_ENABLE_STATS=true
PROXY_CONFIG_HEALTH_CHECK=true
PROXY_CONFIG_LOAD_BALANCE=true

# 模板配置
PROXY_CONFIG_TEMPLATES_DIR=/data/templates
PROXY_CONFIG_CUSTOM_TEMPLATES=true
```

## 🧪 测试计划

### 功能测试
- ✅ 持久化存储功能测试
- ✅ 导入导出功能测试
- ✅ 批量操作功能测试
- ✅ 配置模板功能测试
- ✅ 统计功能测试

### 性能测试
- ✅ 大量配置的加载性能
- ✅ 文件读写性能测试
- ✅ 统计数据收集性能
- ✅ 内存使用情况测试

### 稳定性测试
- ✅ 长时间运行稳定性
- ✅ 文件损坏恢复测试
- ✅ 并发操作安全性测试

## 📈 验收标准

### 功能验收
- ✅ 配置持久化正常工作，重启后数据不丢失
- ✅ 导入导出功能完整，支持批量操作
- ✅ 配置模板功能可用，提高配置效率
- ✅ 统计功能准确，提供有用的监控信息

### 性能验收
- ✅ 支持至少5000个配置项
- ✅ 配置加载时间 < 2秒
- ✅ 文件保存时间 < 1秒
- ✅ 内存使用增长 < 100MB

### 兼容性验收
- ✅ 与v1.1.0完全兼容
- ✅ 配置文件格式向后兼容
- ✅ API接口向后兼容

## 🚀 部署升级

### 升级步骤
1. 备份现有配置数据
2. 停止v1.1.0服务
3. 部署v1.2.0版本
4. 启动服务并验证功能
5. 启用新功能特性

### 回滚方案
1. 停止v1.2.0服务
2. 恢复v1.1.0版本
3. 恢复配置数据
4. 验证功能正常

## 📝 文档更新

- ✅ 更新API文档，添加新接口说明
- ✅ 更新用户手册，介绍新功能使用方法
- ✅ 更新配置文档，说明新的环境变量
- ✅ 编写升级指南，指导用户升级

## 📈 下一步计划

完成第二阶段后，进入[第三阶段：企业级功能](PROXY_CONFIG_PHASE3.md)，主要包括：
- 多租户支持和RBAC权限管理
- Prometheus指标导出和监控告警
- Webhook通知和外部系统集成
- 服务发现和动态配置

---

**阶段状态**: 规划中  
**预计开始**: v1.1.0完成后  
**预计完成**: 10个工作日  
**负责人**: 开发团队
