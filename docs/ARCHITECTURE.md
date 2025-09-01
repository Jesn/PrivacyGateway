# Privacy Gateway 系统架构文档

## 系统概述

Privacy Gateway 是一个高性能的代理服务系统，提供安全的API访问控制和细粒度的令牌管理功能。系统采用模块化设计，支持多种代理模式和认证方式。

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Privacy Gateway                          │
├─────────────────────────────────────────────────────────────────┤
│  Frontend (Web UI)                                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Config Mgmt   │  │   Token Mgmt    │  │   Dashboard     │ │
│  │   - Create      │  │   - Create      │  │   - Statistics  │ │
│  │   - Edit        │  │   - Edit        │  │   - Monitoring  │ │
│  │   - Delete      │  │   - Delete      │  │   - Logs        │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  HTTP Router & Middleware                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Auth Middleware│  │   CORS Middleware│  │   Log Middleware│ │
│  │   - Admin Auth  │  │   - Headers     │  │   - Request Log │ │
│  │   - Token Auth  │  │   - Methods     │  │   - Error Log   │ │
│  │   - Permission  │  │   - Origins     │  │   - Access Log  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  API Layer                                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Config API    │  │   Token API     │  │   Proxy API     │ │
│  │   - CRUD Ops    │  │   - CRUD Ops    │  │   - HTTP Proxy  │ │
│  │   - Validation  │  │   - Validation  │  │   - WS Proxy    │ │
│  │   - Export/Import│  │   - Batch Ops   │  │   - Subdomain   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  Business Logic Layer                                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Config Mgmt   │  │   Token Mgmt    │  │   Proxy Engine  │ │
│  │   - Validation  │  │   - Generation  │  │   - Request     │ │
│  │   - Storage     │  │   - Validation  │  │   - Response    │ │
│  │   - Migration   │  │   - Encryption  │  │   - Streaming   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  Data Layer                                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Memory Store  │  │   File Store    │  │   Cache Layer   │ │
│  │   - Configs     │  │   - Persistence │  │   - Token Cache │ │
│  │   - Tokens      │  │   - Backup      │  │   - Config Cache│ │
│  │   - Statistics  │  │   - Recovery    │  │   - Stats Cache │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      External Services                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Target APIs   │  │   Monitoring    │  │   Log Services  │ │
│  │   - REST APIs   │  │   - Metrics     │  │   - Centralized │ │
│  │   - GraphQL     │  │   - Alerts      │  │   - Structured  │ │
│  │   - WebSocket   │  │   - Health      │  │   - Searchable  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. 前端界面 (Frontend)

**技术栈**: HTML5, CSS3, JavaScript (ES6+)

**主要功能**:
- 代理配置管理界面
- 令牌管理界面
- 实时统计仪表板
- 系统监控和日志查看

**关键文件**:
- `static/index.html` - 主界面
- `static/css/style.css` - 样式文件
- `static/js/app.js` - 主应用逻辑
- `static/js/api.js` - API调用封装

### 2. HTTP路由器 (Router)

**技术栈**: Go net/http

**主要功能**:
- 请求路由分发
- 中间件管理
- 静态文件服务
- WebSocket升级处理

**关键文件**:
- `internal/router/router.go` - 路由器主逻辑
- `internal/middleware/` - 中间件实现

### 3. 认证中间件 (Authentication)

**认证方式**:
- 管理员密钥认证 (X-Log-Secret)
- 令牌认证 (X-Proxy-Token, Authorization Bearer)
- 查询参数认证 (secret, token)

**安全特性**:
- 常量时间比较防止时序攻击
- 令牌哈希存储
- 权限级别控制
- 请求频率限制

### 4. API层 (API Layer)

#### 配置管理API
- `GET /config/proxy` - 获取所有配置
- `POST /config/proxy` - 创建配置
- `PUT /config/proxy/{id}` - 更新配置
- `DELETE /config/proxy/{id}` - 删除配置
- `GET /config/proxy/export` - 导出配置
- `POST /config/proxy/import` - 导入配置

#### 令牌管理API
- `GET /config/proxy/{id}/tokens` - 获取令牌列表
- `POST /config/proxy/{id}/tokens` - 创建令牌
- `PUT /config/proxy/{id}/tokens/{token_id}` - 更新令牌
- `DELETE /config/proxy/{id}/tokens/{token_id}` - 删除令牌
- `POST /config/proxy/{id}/tokens/batch` - 批量操作

#### 代理API
- `GET /proxy` - HTTP代理
- `GET /ws` - WebSocket代理
- `GET /{subdomain}.*` - 子域名代理

### 5. 业务逻辑层 (Business Logic)

#### 配置管理 (ProxyConfig)
```go
type ProxyConfig struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Subdomain   string    `json:"subdomain"`
    TargetURL   string    `json:"target_url"`
    Protocol    string    `json:"protocol"`
    Enabled     bool      `json:"enabled"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    TokenStats  TokenStats `json:"token_stats"`
}
```

#### 令牌管理 (AccessToken)
```go
type AccessToken struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    TokenHash   string    `json:"-"`
    Enabled     bool      `json:"enabled"`
    Permissions []string  `json:"permissions"`
    ExpiresAt   time.Time `json:"expires_at"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CreatedBy   string    `json:"created_by"`
    LastUsedAt  time.Time `json:"last_used_at"`
    UsageCount  int64     `json:"usage_count"`
}
```

#### 代理引擎 (Proxy Engine)
- HTTP请求代理
- WebSocket连接代理
- 请求/响应修改
- 流式数据处理
- 错误处理和重试

### 6. 数据层 (Data Layer)

#### 内存存储 (Memory Storage)
```go
type MemoryStorage struct {
    configs     map[string]*ProxyConfig
    tokens      map[string]map[string]*AccessToken
    tokenHashes map[string]*TokenInfo
    mutex       sync.RWMutex
    maxConfigs  int
}
```

**特性**:
- 高性能读写
- 并发安全
- 内存限制保护
- 自动清理机制

#### 持久化存储 (Persistence)
- JSON文件存储
- 自动备份
- 数据恢复
- 版本控制

## 数据流

### 1. 配置管理流程

```
用户请求 → 认证中间件 → 配置API → 业务逻辑 → 数据存储 → 响应
    ↓
前端界面更新 ← JSON响应 ← 数据序列化 ← 存储操作
```

### 2. 令牌认证流程

```
代理请求 → 令牌提取 → 令牌验证 → 权限检查 → 目标请求 → 响应转发
    ↓           ↓           ↓           ↓           ↓
令牌解析 → 哈希比较 → 过期检查 → 权限匹配 → 统计更新
```

### 3. 代理请求流程

```
客户端请求 → 路由匹配 → 认证验证 → 目标解析 → 请求转发 → 响应处理
    ↓            ↓           ↓           ↓           ↓           ↓
协议识别 → 中间件链 → 令牌验证 → URL构建 → HTTP客户端 → 流式传输
```

## 安全架构

### 1. 认证安全
- **多层认证**: 管理员密钥 + 访问令牌
- **令牌安全**: SHA-256哈希存储，安全随机生成
- **时序攻击防护**: 常量时间比较算法
- **权限控制**: 基于角色的访问控制(RBAC)

### 2. 传输安全
- **HTTPS强制**: 生产环境强制使用HTTPS
- **安全头部**: 添加安全相关HTTP头部
- **CORS控制**: 跨域请求安全控制
- **输入验证**: 严格的输入参数验证

### 3. 存储安全
- **敏感数据加密**: 令牌哈希存储
- **访问控制**: 文件系统权限控制
- **数据备份**: 定期自动备份
- **审计日志**: 完整的操作审计

## 性能优化

### 1. 缓存策略
- **内存缓存**: 热点数据内存缓存
- **令牌缓存**: 令牌验证结果缓存
- **配置缓存**: 代理配置缓存
- **统计缓存**: 使用统计数据缓存

### 2. 并发处理
- **Goroutine池**: 限制并发数量
- **连接复用**: HTTP连接池
- **异步处理**: 非阻塞I/O操作
- **流式处理**: 大文件流式传输

### 3. 资源管理
- **内存限制**: 防止内存泄漏
- **连接限制**: 限制并发连接数
- **超时控制**: 请求超时保护
- **资源清理**: 自动资源回收

## 监控和日志

### 1. 系统监控
- **性能指标**: CPU、内存、网络使用率
- **业务指标**: 请求量、响应时间、错误率
- **健康检查**: 服务可用性监控
- **告警机制**: 异常情况自动告警

### 2. 日志系统
- **结构化日志**: JSON格式日志
- **日志级别**: DEBUG、INFO、WARN、ERROR
- **日志轮转**: 自动日志文件轮转
- **集中收集**: 支持日志收集系统

### 3. 审计追踪
- **操作审计**: 所有管理操作记录
- **访问日志**: 代理请求访问记录
- **安全事件**: 认证失败、异常访问
- **性能分析**: 请求性能分析

## 部署架构

### 1. 单机部署
```
┌─────────────────────────────────┐
│         Privacy Gateway         │
│  ┌─────────────┐ ┌─────────────┐│
│  │   Web UI    │ │   API       ││
│  └─────────────┘ └─────────────┘│
│  ┌─────────────┐ ┌─────────────┐│
│  │   Storage   │ │   Logs      ││
│  └─────────────┘ └─────────────┘│
└─────────────────────────────────┘
```

### 2. 高可用部署
```
┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Load Balancer │
└─────────────────┘    └─────────────────┘
         │                       │
    ┌────▼────┐              ┌───▼────┐
    │Gateway 1│              │Gateway 2│
    └─────────┘              └────────┘
         │                       │
    ┌────▼─────────────────────────▼────┐
    │        Shared Storage             │
    └───────────────────────────────────┘
```

### 3. 容器化部署
```yaml
version: '3.8'
services:
  privacy-gateway:
    image: privacy-gateway:latest
    ports:
      - "8080:8080"
    environment:
      - ADMIN_SECRET=${ADMIN_SECRET}
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
```

## 扩展性设计

### 1. 水平扩展
- **无状态设计**: 服务实例无状态
- **负载均衡**: 支持多实例负载均衡
- **会话共享**: 共享存储支持
- **配置同步**: 配置数据同步机制

### 2. 功能扩展
- **插件系统**: 支持功能插件
- **中间件扩展**: 自定义中间件
- **存储扩展**: 多种存储后端
- **认证扩展**: 多种认证方式

### 3. 协议扩展
- **HTTP/2支持**: 支持HTTP/2协议
- **gRPC代理**: 支持gRPC协议代理
- **TCP代理**: 支持TCP层代理
- **自定义协议**: 支持自定义协议扩展

---

**版本**: 2.0  
**最后更新**: 2024年8月29日  
**维护者**: Privacy Gateway开发团队
