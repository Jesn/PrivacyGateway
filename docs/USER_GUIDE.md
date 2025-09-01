# Privacy Gateway 用户使用指南

## 简介

Privacy Gateway 是一个强大的代理服务，提供安全的API访问控制和令牌管理功能。本指南将帮助您快速上手并充分利用所有功能。

## 快速开始

### 1. 访问管理界面

打开浏览器，访问您的Privacy Gateway管理界面：
```
https://your-domain.com
```

### 2. 管理员认证

首次访问需要输入管理员密钥。请联系系统管理员获取密钥。

### 3. 创建第一个代理配置

1. 点击"添加配置"按钮
2. 填写配置信息：
   - **名称**: 给配置起一个有意义的名称
   - **子域名**: 用于访问的子域名（如：api）
   - **目标URL**: 要代理的目标服务地址
   - **协议**: 选择HTTP或HTTPS
3. 点击"保存"完成创建

### 4. 创建访问令牌

1. 在配置列表中找到刚创建的配置
2. 点击"管理令牌"按钮
3. 点击"创建令牌"
4. 填写令牌信息：
   - **名称**: 令牌的名称
   - **描述**: 令牌的用途说明
   - **权限**: 选择读取或读写权限
   - **过期时间**: 设置令牌的有效期
5. 点击"创建"并复制生成的令牌

⚠️ **重要**: 令牌只显示一次，请立即复制并妥善保存！

## 功能详解

### 代理配置管理

#### 创建配置
- **名称**: 配置的显示名称，便于识别
- **子域名**: 用于子域名代理的前缀，如设置为"api"，则可通过 `api.your-domain.com` 访问
- **目标URL**: 实际要代理的服务地址
- **协议**: 目标服务使用的协议
- **启用状态**: 控制配置是否生效

#### 编辑配置
1. 点击配置行的"编辑"按钮
2. 修改需要更改的字段
3. 点击"保存"应用更改

#### 删除配置
1. 点击配置行的"删除"按钮
2. 确认删除操作

⚠️ **注意**: 删除配置会同时删除所有相关的访问令牌！

### 令牌管理

#### 令牌类型
- **读取令牌**: 只能进行GET请求
- **读写令牌**: 可以进行所有HTTP方法的请求

#### 创建令牌
1. 在配置管理页面点击"管理令牌"
2. 点击"创建令牌"按钮
3. 填写令牌信息
4. 设置权限和过期时间
5. 保存并复制令牌

#### 管理令牌
- **查看令牌**: 查看令牌的详细信息和使用统计
- **编辑令牌**: 修改令牌的名称、描述、权限等
- **禁用令牌**: 临时禁用令牌而不删除
- **删除令牌**: 永久删除令牌

#### 令牌状态
- 🟢 **活跃**: 令牌正常可用
- 🟡 **即将过期**: 令牌将在7天内过期
- 🔴 **已过期**: 令牌已过期，需要更新
- ⚫ **已禁用**: 令牌被手动禁用

### 使用代理服务

#### 方式一：直接代理
使用 `/proxy` 端点进行代理请求：

```bash
# 使用HTTP头部传递令牌（推荐）
curl -H "X-Proxy-Token: your-token" \
     "https://your-domain.com/proxy?target=https://api.example.com/users&config_id=config-uuid"

# 使用查询参数传递令牌
curl "https://your-domain.com/proxy?target=https://api.example.com/users&config_id=config-uuid&token=your-token"
```

#### 方式二：子域名代理
通过配置的子域名直接访问：

```bash
# 如果配置的子域名是"api"
curl -H "X-Proxy-Token: your-token" \
     "https://api.your-domain.com/users"
```

#### 方式三：WebSocket代理
支持WebSocket连接的代理：

```javascript
const ws = new WebSocket('wss://your-domain.com/ws?target=wss://api.example.com/ws&config_id=config-uuid&token=your-token');
```

## 最佳实践

### 安全建议

1. **令牌管理**
   - 为不同的应用或环境创建独立的令牌
   - 定期轮换令牌，特别是长期使用的令牌
   - 设置合理的过期时间，避免永不过期的令牌
   - 及时删除不再使用的令牌

2. **权限控制**
   - 遵循最小权限原则，只授予必要的权限
   - 读取操作使用只读令牌
   - 写入操作使用读写令牌

3. **令牌存储**
   - 不要在代码中硬编码令牌
   - 使用环境变量或安全的配置管理系统
   - 不要在日志中记录令牌信息

### 监控和维护

1. **使用统计**
   - 定期查看令牌使用统计
   - 识别异常的使用模式
   - 监控令牌的活跃度

2. **日志监控**
   - 关注认证失败的日志
   - 监控异常的访问模式
   - 设置告警机制

3. **定期审查**
   - 定期审查所有令牌的必要性
   - 清理过期和未使用的令牌
   - 更新令牌的权限设置

## 故障排除

### 常见问题

#### 1. 认证失败 (401错误)
**可能原因**:
- 令牌无效或已过期
- 令牌被禁用
- 配置ID不正确

**解决方法**:
- 检查令牌是否正确复制
- 确认令牌未过期
- 验证配置ID是否正确

#### 2. 权限不足 (403错误)
**可能原因**:
- 令牌权限不足
- 尝试访问未授权的资源

**解决方法**:
- 检查令牌权限设置
- 确认操作类型是否被允许

#### 3. 配置不存在 (404错误)
**可能原因**:
- 配置ID错误
- 配置已被删除

**解决方法**:
- 验证配置ID
- 检查配置是否存在

#### 4. 代理请求失败
**可能原因**:
- 目标服务不可达
- 网络连接问题
- 目标服务返回错误

**解决方法**:
- 检查目标服务状态
- 验证网络连接
- 查看详细错误信息

### 调试技巧

1. **查看响应头**
   ```bash
   curl -I -H "X-Proxy-Token: your-token" \
        "https://your-domain.com/proxy?target=https://api.example.com&config_id=config-uuid"
   ```

2. **启用详细输出**
   ```bash
   curl -v -H "X-Proxy-Token: your-token" \
        "https://your-domain.com/proxy?target=https://api.example.com&config_id=config-uuid"
   ```

3. **检查令牌状态**
   - 在管理界面查看令牌的详细信息
   - 确认令牌的状态和权限

## 高级功能

### 批量操作

#### 批量创建令牌
可以通过API批量创建多个令牌：

```bash
curl -X POST \
     -H "X-Log-Secret: admin-secret" \
     -H "Content-Type: application/json" \
     -d '{
       "tokens": [
         {"name": "Token 1", "permissions": ["read"]},
         {"name": "Token 2", "permissions": ["read", "write"]}
       ]
     }' \
     "https://your-domain.com/config/proxy/config-uuid/tokens/batch"
```

#### 批量更新令牌
```bash
curl -X PUT \
     -H "X-Log-Secret: admin-secret" \
     -H "Content-Type: application/json" \
     -d '{
       "updates": [
         {"id": "token-1", "enabled": false},
         {"id": "token-2", "permissions": ["read"]}
       ]
     }' \
     "https://your-domain.com/config/proxy/config-uuid/tokens/batch"
```

### 配置导入导出

#### 导出配置
```bash
curl -H "X-Log-Secret: admin-secret" \
     "https://your-domain.com/config/proxy/export" > backup.json
```

#### 导入配置
```bash
curl -X POST \
     -H "X-Log-Secret: admin-secret" \
     -H "Content-Type: application/json" \
     -d @backup.json \
     "https://your-domain.com/config/proxy/import"
```

## 支持和帮助

如果您遇到问题或需要帮助：

1. 查看本用户指南的故障排除部分
2. 检查系统日志获取详细错误信息
3. 联系系统管理员获取技术支持
4. 查看API文档了解更多技术细节

---

**版本**: 1.0  
**最后更新**: 2024年8月29日
