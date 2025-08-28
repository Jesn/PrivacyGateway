# 代理配置管理 API 接口规范

## 📋 概述

本文档定义了Privacy Gateway代理配置管理功能的API接口规范，包括请求格式、响应格式、错误处理等。

## 🔐 认证方式

所有配置管理API都需要管理员权限，支持以下认证方式：

### 请求头认证（推荐）
```http
X-Log-Secret: your-admin-secret
```

### 查询参数认证
```http
GET /config/proxy?secret=your-admin-secret
```

## 📊 数据结构

### ProxyConfig 对象
```json
{
  "id": "uuid-string",
  "name": "配置名称",
  "subdomain": "子域名",
  "target_url": "目标URL",
  "protocol": "http|https",
  "enabled": true,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### ConfigResponse 对象
```json
{
  "configs": [ProxyConfig],
  "total": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5
}
```

## 🔗 API 端点

### 1. 获取配置列表

**请求**
```http
GET /config/proxy
Headers: X-Log-Secret: {admin_secret}
```

**查询参数**
- `search` (string, 可选): 搜索关键词，匹配名称、子域名、目标URL
- `enabled` (boolean, 可选): 筛选启用状态 (true/false)
- `page` (int, 可选): 页码，默认1
- `limit` (int, 可选): 每页数量，默认20，最大100

**响应示例**
```json
{
  "configs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "GitHub API",
      "subdomain": "api",
      "target_url": "https://api.github.com",
      "protocol": "https",
      "enabled": true,
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "total_pages": 1
}
```

**状态码**
- `200 OK`: 成功
- `401 Unauthorized`: 认证失败
- `500 Internal Server Error`: 服务器错误

### 2. 创建配置

**请求**
```http
POST /config/proxy
Headers: 
  X-Log-Secret: {admin_secret}
  Content-Type: application/json
```

**请求体**
```json
{
  "name": "GitHub API",
  "subdomain": "api",
  "target_url": "https://api.github.com",
  "protocol": "https",
  "enabled": true
}
```

**字段说明**
- `name` (string, 必需): 配置名称，1-100字符
- `subdomain` (string, 必需): 子域名，1-63字符，只能包含字母、数字、连字符
- `target_url` (string, 必需): 目标URL，必须是有效的HTTP/HTTPS URL
- `protocol` (string, 必需): 协议类型，"http" 或 "https"
- `enabled` (boolean, 可选): 是否启用，默认true

**响应示例**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "GitHub API",
  "subdomain": "api",
  "target_url": "https://api.github.com",
  "protocol": "https",
  "enabled": true,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**状态码**
- `201 Created`: 创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `409 Conflict`: 子域名已存在
- `500 Internal Server Error`: 服务器错误

### 3. 更新配置

**请求**
```http
PUT /config/proxy?id={config_id}
Headers: 
  X-Log-Secret: {admin_secret}
  Content-Type: application/json
```

**请求体**
```json
{
  "name": "GitHub API Updated",
  "subdomain": "github-api",
  "target_url": "https://api.github.com",
  "protocol": "https",
  "enabled": false
}
```

**响应示例**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "GitHub API Updated",
  "subdomain": "github-api",
  "target_url": "https://api.github.com",
  "protocol": "https",
  "enabled": false,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T11:00:00Z"
}
```

**状态码**
- `200 OK`: 更新成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `404 Not Found`: 配置不存在
- `409 Conflict`: 子域名已存在
- `500 Internal Server Error`: 服务器错误

### 4. 删除配置

**请求**
```http
DELETE /config/proxy?id={config_id}
Headers: X-Log-Secret: {admin_secret}
```

**响应**
- 无响应体

**状态码**
- `204 No Content`: 删除成功
- `401 Unauthorized`: 认证失败
- `404 Not Found`: 配置不存在
- `500 Internal Server Error`: 服务器错误

## ❌ 错误处理

### 错误响应格式
```json
{
  "error": "错误描述",
  "code": "ERROR_CODE",
  "details": "详细错误信息"
}
```

### 常见错误码

| HTTP状态码 | 错误描述 | 解决方案 |
|-----------|----------|----------|
| 400 | 请求参数错误 | 检查请求格式和参数 |
| 401 | 认证失败 | 检查管理员密钥 |
| 404 | 配置不存在 | 检查配置ID是否正确 |
| 409 | 子域名冲突 | 使用不同的子域名 |
| 500 | 服务器内部错误 | 检查服务器日志 |

### 字段验证错误

**名称验证**
- 不能为空
- 长度不能超过100字符

**子域名验证**
- 不能为空
- 长度1-63字符
- 只能包含字母、数字、连字符
- 不能以连字符开头或结尾
- 不能使用保留子域名 (www, api, admin, mail, ftp, localhost)

**目标URL验证**
- 不能为空
- 必须是有效的URL格式
- 必须使用http或https协议
- 必须包含主机名

**协议验证**
- 必须是 "http" 或 "https"

## 🔄 使用示例

### JavaScript 示例
```javascript
// 获取配置列表
async function getConfigs() {
    const response = await fetch('/config/proxy', {
        headers: {
            'X-Log-Secret': 'your-admin-secret'
        }
    });
    return await response.json();
}

// 创建配置
async function createConfig(config) {
    const response = await fetch('/config/proxy', {
        method: 'POST',
        headers: {
            'X-Log-Secret': 'your-admin-secret',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(config)
    });
    return await response.json();
}

// 更新配置
async function updateConfig(id, config) {
    const response = await fetch(`/config/proxy?id=${id}`, {
        method: 'PUT',
        headers: {
            'X-Log-Secret': 'your-admin-secret',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(config)
    });
    return await response.json();
}

// 删除配置
async function deleteConfig(id) {
    const response = await fetch(`/config/proxy?id=${id}`, {
        method: 'DELETE',
        headers: {
            'X-Log-Secret': 'your-admin-secret'
        }
    });
    return response.ok;
}
```

### curl 示例
```bash
# 获取配置列表
curl -H "X-Log-Secret: your-admin-secret" \
     "http://localhost:10805/config/proxy"

# 创建配置
curl -X POST \
     -H "X-Log-Secret: your-admin-secret" \
     -H "Content-Type: application/json" \
     -d '{"name":"GitHub API","subdomain":"api","target_url":"https://api.github.com","protocol":"https","enabled":true}' \
     "http://localhost:10805/config/proxy"

# 更新配置
curl -X PUT \
     -H "X-Log-Secret: your-admin-secret" \
     -H "Content-Type: application/json" \
     -d '{"name":"GitHub API Updated","subdomain":"github-api","target_url":"https://api.github.com","protocol":"https","enabled":false}' \
     "http://localhost:10805/config/proxy?id=550e8400-e29b-41d4-a716-446655440000"

# 删除配置
curl -X DELETE \
     -H "X-Log-Secret: your-admin-secret" \
     "http://localhost:10805/config/proxy?id=550e8400-e29b-41d4-a716-446655440000"
```

## 📈 性能考虑

- **分页**: 默认每页20条记录，最大100条
- **搜索**: 支持名称、子域名、目标URL的模糊搜索
- **缓存**: 配置数据存储在内存中，响应速度快
- **并发**: 支持并发读写操作，使用读写锁保证数据一致性

## 🔒 安全考虑

- **认证**: 所有操作都需要管理员密钥
- **验证**: 严格的输入验证，防止恶意数据
- **CORS**: 支持跨域请求，便于前端集成
- **日志**: 记录所有配置变更操作

---

**文档版本**: v1.0  
**最后更新**: 2024-01-XX  
**维护者**: Privacy Gateway Team
