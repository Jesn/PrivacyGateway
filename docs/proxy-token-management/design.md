# 代理配置令牌管理设计文档

## 📋 文档信息

- **项目名称**: Privacy Gateway - 代理配置令牌管理
- **文档版本**: v1.0.0
- **创建日期**: 2025-08-29
- **最后更新**: 2025-08-29
- **文档状态**: 设计阶段

## 🏗️ 系统架构设计

### 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端界面      │    │   API网关       │    │   存储层        │
│                 │    │                 │    │                 │
│ • 令牌管理界面  │◄──►│ • 认证中间件    │◄──►│ • 配置存储      │
│ • 配置弹框      │    │ • 令牌验证      │    │ • 令牌存储      │
│ • 使用统计      │    │ • API路由       │    │ • 统计存储      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   代理服务      │
                       │                 │
                       │ • HTTP代理      │
                       │ • 子域名代理    │
                       │ • 请求转发      │
                       └─────────────────┘
```

### 模块划分

#### 1. 认证模块 (Authentication Module)
- **职责**: 令牌验证、权限检查、认证中间件
- **位置**: `internal/handler/auth.go`
- **接口**: `TokenAuthenticator`

#### 2. 令牌管理模块 (Token Management Module)
- **职责**: 令牌CRUD操作、令牌生成、使用统计
- **位置**: `internal/proxyconfig/token.go`
- **接口**: `TokenManager`

#### 3. 存储扩展模块 (Storage Extension Module)
- **职责**: 令牌数据持久化、配置扩展
- **位置**: `internal/proxyconfig/storage.go`
- **接口**: `Storage` (扩展)

#### 4. API控制器模块 (API Controller Module)
- **职责**: 令牌管理API端点、请求处理
- **位置**: `internal/handler/token.go`
- **接口**: HTTP API

## 📊 数据结构设计

### 核心数据结构

#### 1. AccessToken 结构
```go
// AccessToken 访问令牌结构
type AccessToken struct {
    ID          string     `json:"id"`                    // 令牌唯一标识
    Name        string     `json:"name"`                  // 令牌名称
    TokenHash   string     `json:"token_hash"`            // 令牌哈希值(不存储明文)
    Permissions []string   `json:"permissions"`           // 权限列表
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`  // 过期时间
    CreatedAt   time.Time  `json:"created_at"`            // 创建时间
    UpdatedAt   time.Time  `json:"updated_at"`            // 更新时间
    LastUsed    *time.Time `json:"last_used,omitempty"`   // 最后使用时间
    UsageCount  int64      `json:"usage_count"`           // 使用次数
    Enabled     bool       `json:"enabled"`               // 是否启用
    CreatedBy   string     `json:"created_by,omitempty"`  // 创建者
    Description string     `json:"description,omitempty"` // 描述信息
}
```

#### 2. ProxyConfig 扩展
```go
// ProxyConfig 代理配置结构 (扩展)
type ProxyConfig struct {
    // ... 现有字段
    AccessTokens []AccessToken `json:"access_tokens,omitempty"` // 访问令牌列表
    TokenStats   *TokenStats   `json:"token_stats,omitempty"`   // 令牌统计信息
}

// TokenStats 令牌统计信息
type TokenStats struct {
    TotalTokens    int       `json:"total_tokens"`     // 总令牌数
    ActiveTokens   int       `json:"active_tokens"`    // 活跃令牌数
    TotalRequests  int64     `json:"total_requests"`   // 总请求数
    LastTokenUsed  time.Time `json:"last_token_used"`  // 最后令牌使用时间
}
```

#### 3. 权限定义
```go
// Permission 权限常量
const (
    PermissionRead  = "read"   // 只读权限：可以发起代理请求
    PermissionWrite = "write"  // 读写权限：可以修改配置(预留)
    PermissionAdmin = "admin"  // 管理权限：完全控制(预留)
)

// PermissionSet 权限集合
type PermissionSet struct {
    Read  bool `json:"read"`
    Write bool `json:"write"`
    Admin bool `json:"admin"`
}
```

### 数据库设计

#### 存储结构 (JSON文件)
```json
{
  "config_id": {
    "id": "config_id",
    "name": "配置名称",
    "subdomain": "subdomain",
    "target_url": "https://example.com",
    "protocol": "https",
    "enabled": true,
    "created_at": "2025-08-29T00:00:00Z",
    "updated_at": "2025-08-29T00:00:00Z",
    "access_tokens": [
      {
        "id": "token_id",
        "name": "令牌名称",
        "token_hash": "hashed_token_value",
        "permissions": ["read"],
        "expires_at": "2025-12-31T23:59:59Z",
        "created_at": "2025-08-29T00:00:00Z",
        "updated_at": "2025-08-29T00:00:00Z",
        "last_used": "2025-08-29T12:00:00Z",
        "usage_count": 100,
        "enabled": true,
        "created_by": "admin",
        "description": "客户端A的访问令牌"
      }
    ],
    "token_stats": {
      "total_tokens": 3,
      "active_tokens": 2,
      "total_requests": 1500,
      "last_token_used": "2025-08-29T12:00:00Z"
    }
  }
}
```

## 🔧 API设计

### RESTful API 端点

#### 1. 令牌管理API
```http
# 获取配置的所有令牌
GET /config/proxy/{configId}/tokens
Authorization: X-Log-Secret: admin-secret

# 创建新令牌
POST /config/proxy/{configId}/tokens
Authorization: X-Log-Secret: admin-secret
Content-Type: application/json
{
  "name": "客户端A",
  "permissions": ["read"],
  "expires_at": "2025-12-31T23:59:59Z",
  "description": "客户端A的访问令牌"
}

# 更新令牌
PUT /config/proxy/{configId}/tokens/{tokenId}
Authorization: X-Log-Secret: admin-secret
Content-Type: application/json
{
  "name": "客户端A-更新",
  "permissions": ["read", "write"],
  "enabled": false
}

# 删除令牌
DELETE /config/proxy/{configId}/tokens/{tokenId}
Authorization: X-Log-Secret: admin-secret

# 获取令牌详情
GET /config/proxy/{configId}/tokens/{tokenId}
Authorization: X-Log-Secret: admin-secret

# 重新生成令牌
POST /config/proxy/{configId}/tokens/{tokenId}/regenerate
Authorization: X-Log-Secret: admin-secret
```

#### 2. 令牌使用API
```http
# 使用令牌访问代理 (HTTP代理)
GET /proxy?target=https://example.com
X-Proxy-Token: generated-token-value

# 使用令牌访问代理 (子域名代理)
GET http://subdomain.localhost:10805/path
X-Proxy-Token: generated-token-value
```

#### 3. 统计查询API
```http
# 获取令牌使用统计
GET /config/proxy/{configId}/tokens/{tokenId}/stats
Authorization: X-Log-Secret: admin-secret

# 获取配置的令牌统计
GET /config/proxy/{configId}/token-stats
Authorization: X-Log-Secret: admin-secret
```

### API响应格式

#### 成功响应
```json
{
  "success": true,
  "data": {
    "id": "token_id",
    "name": "客户端A",
    "token": "generated-token-value",  // 仅在创建时返回
    "permissions": ["read"],
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2025-08-29T00:00:00Z",
    "enabled": true
  },
  "message": "令牌创建成功"
}
```

#### 错误响应
```json
{
  "success": false,
  "error": {
    "code": "TOKEN_NOT_FOUND",
    "message": "指定的令牌不存在",
    "details": {
      "token_id": "invalid_token_id"
    }
  }
}
```

## 🔒 安全设计

### 令牌生成算法
```go
// 生成安全的随机令牌
func GenerateToken() string {
    // 使用加密安全的随机数生成器
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        panic(err)
    }
    
    // 使用Base64 URL编码，避免特殊字符
    return base64.URLEncoding.EncodeToString(bytes)
}

// 令牌哈希存储
func HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

### 认证流程
```
1. 客户端发送请求 + X-Proxy-Token
2. 提取配置ID (从子域名或target参数)
3. 查找配置的令牌列表
4. 验证令牌哈希匹配
5. 检查令牌是否启用
6. 检查令牌是否过期
7. 检查权限是否足够
8. 更新使用统计
9. 允许/拒绝请求
```

### 权限控制矩阵
```
操作类型          | read权限 | write权限 | admin权限
------------------|----------|-----------|----------
发起代理请求      |    ✓     |     ✓     |    ✓
查看配置信息      |    ✗     |     ✓     |    ✓
修改配置          |    ✗     |     ✓     |    ✓
管理令牌          |    ✗     |     ✗     |    ✓
删除配置          |    ✗     |     ✗     |    ✓
```

### 安全措施
1. **令牌存储**: 只存储哈希值，不存储明文
2. **传输安全**: 建议使用HTTPS传输令牌
3. **过期控制**: 支持令牌过期时间设置
4. **使用限制**: 可扩展支持使用次数限制
5. **审计日志**: 记录所有令牌操作和使用情况

## 🎨 前端设计

### 界面结构
```
代理配置弹框
├── 基本信息区域
├── 请求配置区域
├── 安全配置区域
├── 高级配置区域
├── 令牌管理区域 (新增)
│   ├── 令牌列表
│   │   ├── 令牌名称
│   │   ├── 权限标签
│   │   ├── 状态指示
│   │   ├── 使用统计
│   │   └── 操作按钮
│   ├── 新增令牌按钮
│   └── 批量操作按钮
└── 状态配置区域
```

### 令牌管理弹框
```html
<div id="token-modal" class="modal">
  <div class="modal-content">
    <div class="modal-header">
      <h3>🔑 令牌管理</h3>
      <button class="close-btn">&times;</button>
    </div>
    
    <div class="modal-body">
      <form id="token-form">
        <div class="form-group">
          <label>令牌名称</label>
          <input type="text" name="name" required>
        </div>
        
        <div class="form-group">
          <label>权限设置</label>
          <div class="permission-checkboxes">
            <label><input type="checkbox" value="read"> 只读</label>
            <label><input type="checkbox" value="write"> 读写</label>
          </div>
        </div>
        
        <div class="form-group">
          <label>过期时间 (可选)</label>
          <input type="datetime-local" name="expires_at">
        </div>
        
        <div class="form-group">
          <label>描述信息</label>
          <textarea name="description" rows="3"></textarea>
        </div>
      </form>
    </div>
    
    <div class="modal-footer">
      <button type="button" onclick="saveToken()">保存</button>
      <button type="button" onclick="closeModal()">取消</button>
    </div>
  </div>
</div>
```

### JavaScript API封装
```javascript
// 令牌管理API封装
class TokenAPI {
    constructor(baseURL, adminSecret) {
        this.baseURL = baseURL;
        this.adminSecret = adminSecret;
    }
    
    async getTokens(configId) {
        const response = await fetch(`${this.baseURL}/config/proxy/${configId}/tokens`, {
            headers: { 'X-Log-Secret': this.adminSecret }
        });
        return response.json();
    }
    
    async createToken(configId, tokenData) {
        const response = await fetch(`${this.baseURL}/config/proxy/${configId}/tokens`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Log-Secret': this.adminSecret
            },
            body: JSON.stringify(tokenData)
        });
        return response.json();
    }
    
    async updateToken(configId, tokenId, tokenData) {
        const response = await fetch(`${this.baseURL}/config/proxy/${configId}/tokens/${tokenId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-Log-Secret': this.adminSecret
            },
            body: JSON.stringify(tokenData)
        });
        return response.json();
    }
    
    async deleteToken(configId, tokenId) {
        const response = await fetch(`${this.baseURL}/config/proxy/${configId}/tokens/${tokenId}`, {
            method: 'DELETE',
            headers: { 'X-Log-Secret': this.adminSecret }
        });
        return response.json();
    }
}
```

## 🧪 测试设计

### 单元测试
```go
// 令牌生成测试
func TestGenerateToken(t *testing.T) {
    token1 := GenerateToken()
    token2 := GenerateToken()
    
    // 令牌应该不同
    assert.NotEqual(t, token1, token2)
    
    // 令牌长度应该正确
    assert.Equal(t, 44, len(token1)) // Base64编码的32字节
}

// 令牌验证测试
func TestValidateToken(t *testing.T) {
    storage := NewMemoryStorage(100)
    config := &ProxyConfig{ID: "test-config"}
    token := &AccessToken{
        ID: "test-token",
        TokenHash: HashToken("test-token-value"),
        Permissions: []string{"read"},
        Enabled: true,
    }
    
    storage.Add(config)
    storage.AddToken(config.ID, token)
    
    // 有效令牌应该验证成功
    assert.True(t, storage.ValidateToken(config.ID, "test-token-value"))
    
    // 无效令牌应该验证失败
    assert.False(t, storage.ValidateToken(config.ID, "invalid-token"))
}
```

### 集成测试
```go
// API集成测试
func TestTokenAPI(t *testing.T) {
    // 创建测试服务器
    server := httptest.NewServer(createTestHandler())
    defer server.Close()
    
    // 测试创建令牌
    tokenData := map[string]interface{}{
        "name": "测试令牌",
        "permissions": []string{"read"},
    }
    
    resp := createToken(server.URL, "test-config", tokenData)
    assert.Equal(t, 200, resp.StatusCode)
    
    // 测试使用令牌
    token := extractTokenFromResponse(resp)
    proxyResp := useToken(server.URL, token, "https://httpbin.org")
    assert.Equal(t, 200, proxyResp.StatusCode)
}
```

## 📈 性能优化

### 缓存策略
```go
// 令牌验证缓存
type TokenCache struct {
    cache map[string]*CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

type CacheEntry struct {
    Valid     bool
    ExpiresAt time.Time
}

func (tc *TokenCache) ValidateToken(configID, token string) bool {
    key := fmt.Sprintf("%s:%s", configID, HashToken(token))
    
    tc.mutex.RLock()
    entry, exists := tc.cache[key]
    tc.mutex.RUnlock()
    
    if exists && time.Now().Before(entry.ExpiresAt) {
        return entry.Valid
    }
    
    // 缓存未命中，执行实际验证
    valid := tc.storage.ValidateToken(configID, token)
    
    // 更新缓存
    tc.mutex.Lock()
    tc.cache[key] = &CacheEntry{
        Valid:     valid,
        ExpiresAt: time.Now().Add(tc.ttl),
    }
    tc.mutex.Unlock()
    
    return valid
}
```

### 性能指标
- 令牌验证延迟: < 10ms (99th percentile)
- 令牌缓存命中率: > 90%
- 并发令牌验证: > 1000 QPS
- 内存使用增长: < 10MB (1000个令牌)

## 🔄 部署策略

### 数据迁移
```go
// 配置数据迁移
func MigrateConfigData(oldData map[string]*ProxyConfig) error {
    for id, config := range oldData {
        // 为现有配置添加空的令牌列表
        if config.AccessTokens == nil {
            config.AccessTokens = []AccessToken{}
        }
        
        // 初始化令牌统计
        if config.TokenStats == nil {
            config.TokenStats = &TokenStats{
                TotalTokens:   0,
                ActiveTokens:  0,
                TotalRequests: 0,
            }
        }
    }
    return nil
}
```

### 向后兼容
- 管理员密钥 `X-Log-Secret` 继续有效
- 现有API端点保持不变
- 新增的令牌功能为可选功能
- 配置文件格式向后兼容

### 部署检查清单
- [ ] 数据备份完成
- [ ] 配置文件迁移测试
- [ ] API兼容性测试
- [ ] 性能基准测试
- [ ] 安全扫描通过
- [ ] 文档更新完成

## 🎯 设计原则

### 核心原则
1. **安全第一**: 所有设计决策优先考虑安全性
2. **向后兼容**: 不破坏现有功能和用户体验
3. **简单易用**: 提供直观的用户界面和API
4. **性能优先**: 确保不影响现有系统性能

### 技术原则
1. **模块化设计**: 功能模块独立，易于测试和维护
2. **接口抽象**: 通过接口实现松耦合
3. **错误处理**: 完善的错误处理和用户反馈
4. **可扩展性**: 为未来功能扩展预留空间

---

**文档维护**: 本设计文档将随着开发进展持续更新和完善。
