# API 文档

Privacy Gateway Frontend 的 API 接口文档。

## 📋 目录

- [认证 API](#认证-api)
- [代理配置 API](#代理配置-api)
- [访问日志 API](#访问日志-api)
- [系统状态 API](#系统状态-api)
- [错误处理](#错误处理)
- [使用示例](#使用示例)

## 🔐 认证 API

### 登录验证

验证管理员密钥并获取访问权限。

**请求**
```http
POST /config/proxy
Content-Type: application/json

{
  "admin_secret": "your-admin-secret"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "configs": [],
    "session_valid": true
  }
}
```

**错误响应**
```json
{
  "success": false,
  "error": "Invalid admin secret"
}
```

### 会话验证

检查当前会话是否有效。

**请求**
```http
GET /config/proxy
```

**响应**
```json
{
  "success": true,
  "data": {
    "configs": [...],
    "session_valid": true
  }
}
```

## ⚙️ 代理配置 API

### 获取所有配置

获取所有代理配置列表。

**请求**
```http
GET /config/proxy
```

**响应**
```json
{
  "success": true,
  "data": {
    "configs": [
      {
        "id": "config-1",
        "name": "示例配置",
        "target_url": "https://api.example.com",
        "listen_path": "/api/v1",
        "enabled": true,
        "created_at": "2023-10-15T08:30:00Z",
        "updated_at": "2023-10-15T08:30:00Z"
      }
    ],
    "total": 1
  }
}
```

### 创建配置

创建新的代理配置。

**请求**
```http
POST /config/proxy
Content-Type: application/json

{
  "name": "新配置",
  "target_url": "https://api.example.com",
  "listen_path": "/api/v1",
  "enabled": true,
  "headers": {
    "Authorization": "Bearer token"
  },
  "rate_limit": {
    "requests_per_minute": 100
  }
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "config-2",
    "name": "新配置",
    "target_url": "https://api.example.com",
    "listen_path": "/api/v1",
    "enabled": true,
    "created_at": "2023-10-15T09:00:00Z"
  }
}
```

### 更新配置

更新现有的代理配置。

**请求**
```http
PUT /config/proxy/{id}
Content-Type: application/json

{
  "name": "更新的配置",
  "enabled": false
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": "config-1",
    "name": "更新的配置",
    "enabled": false,
    "updated_at": "2023-10-15T09:30:00Z"
  }
}
```

### 删除配置

删除指定的代理配置。

**请求**
```http
DELETE /config/proxy/{id}
```

**响应**
```json
{
  "success": true,
  "message": "配置已删除"
}
```

## 📋 访问日志 API

### 获取访问日志

获取代理访问日志列表。

**请求**
```http
GET /logs?page=1&limit=20&start_time=2023-10-15T00:00:00Z&end_time=2023-10-15T23:59:59Z
```

**查询参数**
- `page` (可选): 页码，默认为 1
- `limit` (可选): 每页数量，默认为 20
- `start_time` (可选): 开始时间 (ISO 8601)
- `end_time` (可选): 结束时间 (ISO 8601)
- `config_id` (可选): 配置ID过滤
- `status` (可选): HTTP状态码过滤

**响应**
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log-1",
        "config_id": "config-1",
        "config_name": "示例配置",
        "method": "GET",
        "path": "/api/v1/users",
        "target_url": "https://api.example.com/users",
        "status_code": 200,
        "response_time": 150,
        "request_size": 0,
        "response_size": 1024,
        "client_ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "timestamp": "2023-10-15T08:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### 获取日志统计

获取访问日志的统计信息。

**请求**
```http
GET /logs/stats?period=24h
```

**查询参数**
- `period`: 统计周期 (1h, 24h, 7d, 30d)

**响应**
```json
{
  "success": true,
  "data": {
    "total_requests": 1000,
    "success_rate": 95.5,
    "avg_response_time": 200,
    "error_count": 45,
    "top_endpoints": [
      {
        "path": "/api/v1/users",
        "count": 300
      }
    ],
    "status_distribution": {
      "200": 850,
      "404": 30,
      "500": 15
    }
  }
}
```

## 📊 系统状态 API

### 获取系统状态

获取系统运行状态和健康信息。

**请求**
```http
GET /status
```

**响应**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "uptime": 86400,
    "version": "2.0.0",
    "memory_usage": {
      "used": 50,
      "total": 100,
      "percentage": 50
    },
    "active_connections": 25,
    "total_configs": 5,
    "active_configs": 4
  }
}
```

### 获取性能指标

获取系统性能指标。

**请求**
```http
GET /metrics
```

**响应**
```json
{
  "success": true,
  "data": {
    "requests_per_second": 10.5,
    "avg_response_time": 150,
    "error_rate": 0.05,
    "cpu_usage": 25.5,
    "memory_usage": 60.2,
    "disk_usage": 45.8
  }
}
```

## ❌ 错误处理

### 错误响应格式

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

### 常见错误码

| 错误码 | HTTP状态码 | 描述 |
|--------|------------|------|
| `INVALID_ADMIN_SECRET` | 401 | 管理员密钥无效 |
| `SESSION_EXPIRED` | 401 | 会话已过期 |
| `CONFIG_NOT_FOUND` | 404 | 配置不存在 |
| `INVALID_REQUEST` | 400 | 请求参数无效 |
| `DUPLICATE_CONFIG` | 409 | 配置已存在 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 错误处理示例

```javascript
try {
  const response = await fetch('/config/proxy', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(configData)
  });
  
  const result = await response.json();
  
  if (!result.success) {
    throw new Error(result.error);
  }
  
  return result.data;
} catch (error) {
  console.error('API请求失败:', error.message);
  throw error;
}
```

## 💡 使用示例

### 完整的认证流程

```javascript
import { authModule } from '@modules/auth.js';

// 登录
try {
  await authModule.login('your-admin-secret');
  console.log('登录成功');
} catch (error) {
  console.error('登录失败:', error.message);
}

// 检查会话状态
if (authModule.isLoggedIn()) {
  console.log('用户已登录');
} else {
  console.log('用户未登录');
}

// 退出登录
authModule.logout();
```

### 配置管理示例

```javascript
import { proxyAPI } from '@api/proxy.js';

// 获取所有配置
const configs = await proxyAPI.getConfigs();

// 创建新配置
const newConfig = await proxyAPI.createConfig({
  name: '新API配置',
  target_url: 'https://api.example.com',
  listen_path: '/api/v1',
  enabled: true
});

// 更新配置
const updatedConfig = await proxyAPI.updateConfig(newConfig.id, {
  enabled: false
});

// 删除配置
await proxyAPI.deleteConfig(newConfig.id);
```

### 日志查询示例

```javascript
import { logsAPI } from '@api/logs.js';

// 获取最近的日志
const logs = await logsAPI.getLogs({
  page: 1,
  limit: 50,
  start_time: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
});

// 获取统计信息
const stats = await logsAPI.getStats('24h');
console.log(`成功率: ${stats.success_rate}%`);
```

### WebSocket 连接示例

```javascript
import { wsManager } from '@utils/WebSocket.js';

// 连接WebSocket
wsManager.connect();

// 监听实时日志
wsManager.on('log', (logEntry) => {
  console.log('新日志:', logEntry);
});

// 监听系统状态
wsManager.on('status', (status) => {
  console.log('系统状态:', status);
});
```

## 🔧 API客户端配置

### 基础配置

```javascript
// api/base.js
export const API_BASE_URL = 'http://localhost:10805';
export const API_TIMEOUT = 10000;

export const defaultHeaders = {
  'Content-Type': 'application/json',
  'Accept': 'application/json'
};
```

### 请求拦截器

```javascript
// 添加请求拦截器
httpClient.interceptors.request.use((config) => {
  // 添加认证头
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  return config;
});

// 添加响应拦截器
httpClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 处理认证失败
      authModule.logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

## 📝 注意事项

1. **认证**: 所有API请求都需要有效的会话或管理员密钥
2. **限流**: API可能有速率限制，请注意请求频率
3. **缓存**: 某些数据可能被缓存，使用适当的缓存策略
4. **错误处理**: 始终处理API错误和网络异常
5. **版本兼容**: API版本可能会更新，注意兼容性

## 🔄 API版本

当前API版本: **v2.0**

版本更新时会保持向后兼容，重大变更会提前通知。
