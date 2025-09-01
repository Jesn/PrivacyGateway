# Privacy Gateway API 路由文档

## 概述

Privacy Gateway 提供了完整的代理服务和管理API，支持多种认证方式和CORS跨域访问。

## 主要服务端点

### 静态文件和子域名代理
- **路径**: `/`
- **方法**: `GET, POST, PUT, DELETE, OPTIONS`
- **功能**: 
  - 静态文件服务（当不是子域名请求时）
  - 子域名代理服务（当检测到子域名时）
- **认证**: 子域名代理需要管理员密钥或访问令牌

### HTTP代理服务
- **路径**: `/proxy`
- **方法**: `GET, POST, PUT, DELETE, OPTIONS`
- **参数**: 
  - `target` (必需): 目标URL
  - `config_id` (可选): 配置ID，用于令牌认证
- **认证**: 管理员密钥或访问令牌
- **示例**: 
  ```bash
  # 使用管理员密钥
  curl -H "X-Log-Secret: your-secret" \
    "http://localhost:10805/proxy?target=https://httpbin.org/get"
  
  # 使用访问令牌
  curl -H "X-Proxy-Token: your-token" \
    "http://localhost:10805/proxy?target=https://httpbin.org/get&config_id=config-123"
  ```

### WebSocket代理服务
- **路径**: `/ws`
- **方法**: `GET` (WebSocket升级)
- **参数**: 
  - `target` (必需): 目标WebSocket URL
- **认证**: 管理员密钥
- **功能**: WebSocket协议代理

## API管理端点

### 代理配置管理
- **路径**: `/config/proxy`
- **方法**: `GET, POST, PUT, DELETE, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**: 
  - `GET`: 获取配置列表
  - `POST`: 创建新配置
  - `PUT`: 更新配置（需要配置ID）
  - `DELETE`: 删除配置（需要配置ID）

### 配置导出
- **路径**: `/config/proxy/export`
- **方法**: `GET, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**: 导出所有配置为JSON格式

### 配置导入
- **路径**: `/config/proxy/import`
- **方法**: `POST, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**: 从JSON文件导入配置

### 批量操作
- **路径**: `/config/proxy/batch`
- **方法**: `POST, DELETE, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**: 批量创建、更新或删除配置

## 令牌管理API

### 令牌列表和创建
- **路径**: `/config/proxy/{configID}/tokens`
- **方法**: `GET, POST, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**:
  - `GET`: 获取指定配置的令牌列表和统计信息
  - `POST`: 为指定配置创建新的访问令牌

#### 创建令牌示例
```bash
curl -X POST \
  -H "X-Log-Secret: your-admin-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API访问令牌",
    "description": "用于API访问的令牌",
    "expires_at": "2024-12-31T23:59:59Z"
  }' \
  "http://localhost:10805/config/proxy/config-123/tokens"
```

### 令牌操作
- **路径**: `/config/proxy/{configID}/tokens/{tokenID}`
- **方法**: `GET, PUT, DELETE, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**:
  - `GET`: 获取指定令牌详情
  - `PUT`: 更新令牌信息
  - `DELETE`: 删除令牌

#### 更新令牌示例
```bash
curl -X PUT \
  -H "X-Log-Secret: your-admin-secret" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新后的令牌名称",
    "enabled": false
  }' \
  "http://localhost:10805/config/proxy/config-123/tokens/token-456"
```

## 日志查看

### 访问日志
- **路径**: `/logs`, `/logs/`
- **方法**: `GET, OPTIONS`
- **认证**: 仅管理员密钥
- **功能**: 查看系统访问日志和统计信息

## 认证方式

### 管理员密钥认证
用于配置管理和令牌管理，具有最高权限。

**支持方式**:
- 请求头: `X-Log-Secret: your-admin-secret`
- 查询参数: `?secret=your-admin-secret`

### 访问令牌认证
用于代理请求，权限限制在特定配置范围内。

**支持方式**:
- 专用请求头: `X-Proxy-Token: your-access-token`
- Bearer认证: `Authorization: Bearer your-access-token`
- 查询参数: `?token=your-access-token`

**注意**: 令牌认证需要提供 `config_id` 参数来指定配置范围。

## CORS支持

所有端点都支持CORS跨域访问：

- **允许来源**: `*` (所有来源)
- **允许方法**: `GET, POST, PUT, DELETE, OPTIONS`
- **允许头部**: 
  - `Content-Type`
  - `Authorization`
  - `X-Requested-With`
  - `X-Log-Secret`
  - `X-Proxy-Token`
  - `X-Config-ID`
- **缓存时间**: 24小时

## 响应格式

### 成功响应
```json
{
  "success": true,
  "data": {
    // 响应数据
  },
  "status": 200
}
```

### 错误响应
```json
{
  "success": false,
  "error": "错误描述",
  "error_code": "ERROR_CODE",
  "status": 400
}
```

## 状态码

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如重复名称）
- `500 Internal Server Error`: 服务器内部错误

## 安全注意事项

1. **管理员密钥**: 具有最高权限，请妥善保管
2. **访问令牌**: 仅在指定配置范围内有效，支持过期时间设置
3. **HTTPS**: 生产环境建议使用HTTPS协议
4. **日志记录**: 所有认证尝试都会被记录，便于安全审计
5. **令牌轮换**: 定期更新访问令牌以提高安全性

## 使用示例

### 完整的令牌管理流程

```bash
# 1. 创建代理配置
CONFIG_RESPONSE=$(curl -s -X POST \
  -H "X-Log-Secret: admin-secret" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试配置","subdomain":"test","target_url":"https://httpbin.org"}' \
  "http://localhost:10805/config/proxy")

CONFIG_ID=$(echo $CONFIG_RESPONSE | jq -r '.data.id')

# 2. 创建访问令牌
TOKEN_RESPONSE=$(curl -s -X POST \
  -H "X-Log-Secret: admin-secret" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试令牌","description":"用于测试"}' \
  "http://localhost:10805/config/proxy/$CONFIG_ID/tokens")

TOKEN_VALUE=$(echo $TOKEN_RESPONSE | jq -r '.data.token')

# 3. 使用令牌进行代理请求
curl -H "X-Proxy-Token: $TOKEN_VALUE" \
  "http://localhost:10805/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID"

# 4. 查看令牌使用统计
curl -H "X-Log-Secret: admin-secret" \
  "http://localhost:10805/config/proxy/$CONFIG_ID/tokens"
```

## 更新日志

- **v1.0.0**: 初始版本，支持基本代理功能
- **v2.0.0**: 添加令牌管理功能，支持细粒度访问控制
- **v2.1.0**: 完善路由管理，添加CORS支持和中间件架构
