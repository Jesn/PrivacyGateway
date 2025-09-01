# Privacy Gateway API 文档

## 概述

Privacy Gateway 提供了完整的代理配置和令牌管理API，支持细粒度的访问控制和安全认证机制。

## 认证方式

### 管理员认证
用于配置管理和令牌管理的管理员认证：

```bash
# 通过HTTP头部
curl -H "X-Log-Secret: your-admin-secret" \
     https://your-domain.com/config/proxy

# 通过查询参数
curl "https://your-domain.com/config/proxy?secret=your-admin-secret"
```

### 令牌认证
用于代理请求的令牌认证：

```bash
# 通过HTTP头部 (推荐)
curl -H "X-Proxy-Token: your-access-token" \
     "https://your-domain.com/proxy?target=https://api.example.com&config_id=config-uuid"

# 通过Authorization头部
curl -H "Authorization: Bearer your-access-token" \
     "https://your-domain.com/proxy?target=https://api.example.com&config_id=config-uuid"

# 通过查询参数
curl "https://your-domain.com/proxy?target=https://api.example.com&config_id=config-uuid&token=your-access-token"
```

## 代理配置管理API

### 获取所有配置

```http
GET /config/proxy
```

**认证**: 管理员密钥

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "API代理配置",
      "subdomain": "api",
      "target_url": "https://api.example.com",
      "protocol": "https",
      "enabled": true,
      "created_at": "2024-08-29T10:00:00Z",
      "updated_at": "2024-08-29T10:00:00Z",
      "token_stats": {
        "total_tokens": 5,
        "active_tokens": 3,
        "total_requests": 1250,
        "last_token_used": "2024-08-29T09:30:00Z"
      }
    }
  ]
}
```

### 创建配置

```http
POST /config/proxy
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "新的API代理",
  "subdomain": "newapi",
  "target_url": "https://newapi.example.com",
  "protocol": "https",
  "enabled": true
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "新的API代理",
    "subdomain": "newapi",
    "target_url": "https://newapi.example.com",
    "protocol": "https",
    "enabled": true,
    "created_at": "2024-08-29T10:30:00Z",
    "updated_at": "2024-08-29T10:30:00Z",
    "token_stats": {
      "total_tokens": 0,
      "active_tokens": 0,
      "total_requests": 0,
      "last_token_used": "0001-01-01T00:00:00Z"
    }
  }
}
```

### 更新配置

```http
PUT /config/proxy/{config_id}
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "更新的API代理",
  "target_url": "https://updated-api.example.com",
  "enabled": false
}
```

### 删除配置

```http
DELETE /config/proxy/{config_id}
```

## 令牌管理API

### 获取配置的所有令牌

```http
GET /config/proxy/{config_id}/tokens
```

**认证**: 管理员密钥

**响应示例**:
```json
{
  "success": true,
  "data": {
    "config_id": "550e8400-e29b-41d4-a716-446655440000",
    "tokens": [
      {
        "id": "token-uuid-1",
        "name": "生产环境令牌",
        "description": "用于生产环境的API访问",
        "enabled": true,
        "permissions": ["read", "write"],
        "expires_at": "2025-08-29T10:00:00Z",
        "created_at": "2024-08-29T10:00:00Z",
        "updated_at": "2024-08-29T10:00:00Z",
        "created_by": "admin",
        "last_used_at": "2024-08-29T09:30:00Z",
        "usage_count": 125,
        "status": "active"
      }
    ]
  }
}
```

### 创建令牌

```http
POST /config/proxy/{config_id}/tokens
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "新的访问令牌",
  "description": "用于测试环境的API访问",
  "permissions": ["read"],
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": "token-uuid-2",
    "name": "新的访问令牌",
    "description": "用于测试环境的API访问",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "enabled": true,
    "permissions": ["read"],
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2024-08-29T10:30:00Z",
    "updated_at": "2024-08-29T10:30:00Z",
    "created_by": "admin",
    "last_used_at": "0001-01-01T00:00:00Z",
    "usage_count": 0,
    "status": "active"
  }
}
```

**重要**: 令牌值只在创建时返回一次，请妥善保存。

### 更新令牌

```http
PUT /config/proxy/{config_id}/tokens/{token_id}
Content-Type: application/json
```

**请求体**:
```json
{
  "name": "更新的令牌名称",
  "description": "更新的描述",
  "enabled": false,
  "permissions": ["read", "write"]
}
```

### 删除令牌

```http
DELETE /config/proxy/{config_id}/tokens/{token_id}
```

### 获取单个令牌信息

```http
GET /config/proxy/{config_id}/tokens/{token_id}
```

## 代理请求API

### HTTP代理

```http
GET /proxy?target={target_url}&config_id={config_id}
```

**认证**: 访问令牌

**参数**:
- `target`: 目标URL (必需)
- `config_id`: 配置ID (必需)
- `token`: 访问令牌 (如果不通过头部传递)

**示例**:
```bash
curl -H "X-Proxy-Token: your-token" \
     "https://your-domain.com/proxy?target=https://api.example.com/users&config_id=config-uuid"
```

### 子域名代理

```http
GET https://{subdomain}.your-domain.com/path
```

**认证**: 访问令牌

**示例**:
```bash
curl -H "X-Proxy-Token: your-token" \
     "https://api.your-domain.com/users"
```

## WebSocket代理

```javascript
const ws = new WebSocket('wss://your-domain.com/ws?target=wss://api.example.com/ws&config_id=config-uuid&token=your-token');
```

## 错误响应格式

所有API错误都遵循统一的响应格式：

```json
{
  "success": false,
  "error": "错误描述",
  "error_code": "ERROR_CODE",
  "details": {
    "field": "具体错误信息"
  }
}
```

### 常见错误代码

- `UNAUTHORIZED`: 认证失败
- `FORBIDDEN`: 权限不足
- `NOT_FOUND`: 资源不存在
- `VALIDATION_ERROR`: 请求参数验证失败
- `TOKEN_EXPIRED`: 令牌已过期
- `TOKEN_DISABLED`: 令牌已禁用
- `CONFIG_NOT_FOUND`: 配置不存在
- `DUPLICATE_SUBDOMAIN`: 子域名已存在
- `MAX_TOKENS_EXCEEDED`: 超过最大令牌数量限制

## 状态码

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突
- `429 Too Many Requests`: 请求频率过高
- `500 Internal Server Error`: 服务器内部错误

## 速率限制

API实施了速率限制以防止滥用：

- **管理API**: 每分钟100次请求
- **代理请求**: 根据令牌配置的限制
- **令牌创建**: 每小时10个新令牌

超过限制时返回`429 Too Many Requests`状态码。

## 分页

对于返回列表的API，支持分页参数：

```http
GET /config/proxy?page=1&limit=20&sort=created_at&order=desc
```

**参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 20, 最大: 100)
- `sort`: 排序字段 (默认: created_at)
- `order`: 排序方向 (asc/desc, 默认: desc)

**响应包含分页信息**:
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```
