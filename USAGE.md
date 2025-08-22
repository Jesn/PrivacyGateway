# 使用指南

本文档提供了Privacy Gateway的详细使用示例和高级功能说明。

## 📖 目录

- [基本代理功能](#基本代理功能)
- [WebSocket代理](#websocket代理)
- [Server-Sent Events (SSE)](#server-sent-events-sse)
- [代理配置](#代理配置)
- [敏感头信息过滤](#敏感头信息过滤)
- [管理界面](#管理界面)
- [日志记录模式](#日志记录模式)
- [高级配置](#高级配置)
- [性能优化](#性能优化)
- [集成示例](#集成示例)
- [故障排除](#故障排除)

## 基本代理功能

### HTTP/HTTPS 代理

Privacy Gateway通过 `/proxy` 端点提供HTTP/HTTPS代理服务。

#### 基本格式

```
http://localhost:10805/proxy?target=<目标URL>
```

#### GET 请求示例

```bash
# 获取IP信息
curl "http://localhost:10805/proxy?target=https://httpbin.org/ip"

# 获取网页内容
curl "http://localhost:10805/proxy?target=https://example.com"

# 带查询参数的请求
curl "http://localhost:10805/proxy?target=https://httpbin.org/get?param1=value1&param2=value2"
```

#### POST 请求示例

```bash
# 发送JSON数据
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"test", "value":123}' \
  "http://localhost:10805/proxy?target=https://httpbin.org/post"

# 发送表单数据
curl -X POST \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=secret" \
  "http://localhost:10805/proxy?target=https://httpbin.org/post"

# 上传文件
curl -X POST \
  -F "file=@example.txt" \
  "http://localhost:10805/proxy?target=https://httpbin.org/post"
```

#### 其他HTTP方法

```bash
# PUT 请求
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"id": 1, "status":"updated"}' \
  "http://localhost:10805/proxy?target=https://httpbin.org/put"

# DELETE 请求
curl -X DELETE \
  "http://localhost:10805/proxy?target=https://httpbin.org/delete"

# PATCH 请求
curl -X PATCH \
  -H "Content-Type: application/json" \
  -d '{"status":"patched"}' \
  "http://localhost:10805/proxy?target=https://httpbin.org/patch"
```

#### 自定义请求头

```bash
# 添加自定义请求头
curl -H "Authorization: Bearer token123" \
     -H "X-Custom-Header: custom-value" \
     "http://localhost:10805/proxy?target=https://httpbin.org/headers"

# User-Agent 设置
curl -H "User-Agent: MyApp/1.0" \
     "http://localhost:10805/proxy?target=https://httpbin.org/user-agent"
```

## WebSocket代理

### 基本WebSocket连接

```bash
# WebSocket连接格式
ws://localhost:10805/ws?target=<目标WebSocket地址>

# 使用wscat测试（需要先安装：npm install -g wscat）
wscat -c "ws://localhost:10805/ws?target=wss://echo.websocket.events"

# 连接到WebSocket API
wscat -c "ws://localhost:10805/ws?target=wss://api.example.com/websocket"
```

### WebSocket with 认证

```bash
# 带认证头的WebSocket连接
wscat -c "ws://localhost:10805/ws?target=wss://api.example.com/ws" \
  -H "Authorization: Bearer your-token"
```

## Server-Sent Events (SSE)

Privacy Gateway支持代理Server-Sent Events流。

### 基本SSE代理

```bash
# 代理SSE流
curl -N "http://localhost:10805/proxy?target=https://api.example.com/events"

# 带认证的SSE流
curl -N -H "Authorization: Bearer token" \
     "http://localhost:10805/proxy?target=https://api.example.com/events"
```

### 实时事件监听

```bash
# 监听实时事件并处理
curl -N "http://localhost:10805/proxy?target=https://api.example.com/events" | \
while IFS= read -r line; do
    echo "Received: $line"
    # 处理事件数据
done
```

## 代理配置

### 环境变量配置

```bash
# 设置默认上游代理
export DEFAULT_PROXY="http://proxy.company.com:8080"

# 设置代理白名单
export PROXY_WHITELIST="proxy1.company.com,proxy2.company.com"

# 启动服务
./privacy-gateway
```

### 查询参数方式指定代理

```bash
# 使用HTTP代理
curl "http://localhost:10805/proxy?target=https://httpbin.org/ip&proxy=http://proxy.example.com:8080"

# 使用SOCKS5代理
curl "http://localhost:10805/proxy?target=https://httpbin.org/ip&proxy=socks5://proxy.example.com:1080"

# 使用带认证的代理
curl "http://localhost:10805/proxy?target=https://httpbin.org/ip&proxy=http://user:pass@proxy.example.com:8080"
```

### 请求头方式配置代理

```bash
# 创建代理配置JSON
proxy_config='{"url":"http://proxy.example.com:8080","timeout":30}'

# Base64编码
proxy_config_b64=$(echo -n "$proxy_config" | base64)

# 发送请求
curl -H "X-Proxy-Config: $proxy_config_b64" \
     "http://localhost:10805/proxy?target=https://httpbin.org/ip"
```

#### 代理配置JSON格式

```json
{
  "url": "http://proxy.example.com:8080",
  "type": "http",
  "timeout": 30,
  "auth": {
    "username": "user",
    "password": "pass"
  }
}
```

## 敏感头信息过滤

Privacy Gateway会自动过滤敏感的请求头信息，保护用户隐私。

### 默认过滤的头信息

```bash
# 默认过滤的头信息前缀
SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id,x-trace,x-correlation-id,x-country,x-region,x-city"
```

### 自定义过滤规则

```bash
# 自定义要过滤的头信息
export SENSITIVE_HEADERS="cf-,x-forwarded,x-real-ip,x-custom-"

# 启动服务
./privacy-gateway
```

### 测试头信息过滤

```bash
# 发送包含敏感头信息的请求
curl -H "CF-Ray: 123456789" \
     -H "X-Forwarded-For: 192.168.1.1" \
     -H "X-Real-IP: 10.0.0.1" \
     -H "Custom-Header: safe-value" \
     "http://localhost:10805/proxy?target=https://httpbin.org/headers"

# 查看结果，敏感头信息会被过滤掉
```

## 管理界面

### 访问管理界面

```bash
# 设置管理密钥
export ADMIN_SECRET="your-secure-secret"

# 启动服务
./privacy-gateway

# 访问管理界面
open http://localhost:10805/logs?secret=your-secure-secret
```

### API接口

```bash
# 获取日志列表
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api"

# 按状态码筛选
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api?status=404"

# 按域名筛选
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api?domain=httpbin.org"

# 搜索功能
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api?search=json"

# 分页查询
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api?page=2&limit=50"
```

### 获取特定日志详情

```bash
# 根据ID获取日志详情
curl -H "X-Log-Secret: your-secure-secret" \
     "http://localhost:10805/logs/api?id=log-id-here"
```

## 日志记录模式

Privacy Gateway支持两种日志记录模式，可以根据需要选择。

### 默认模式 (LOG_RECORD_200=false)

适合生产环境，节省存储空间和内存。

```bash
# 设置默认模式
export LOG_RECORD_200=false
export ADMIN_SECRET="your-secret"
./privacy-gateway
```

**记录内容**:
- ✅ 所有请求的基本信息（时间、方法、目标、状态码、耗时）
- ✅ 非200状态码请求的详细信息（请求体、响应体、请求头）
- ❌ 200状态码请求的详细信息（仅记录基本信息）

### 完整记录模式 (LOG_RECORD_200=true)

适合开发环境和调试，记录所有详细信息。

```bash
# 设置完整记录模式
export LOG_RECORD_200=true
export ADMIN_SECRET="your-secret"
./privacy-gateway
```

**记录内容**:
- ✅ 所有请求的基本信息
- ✅ 所有请求的详细信息（包括200状态码）
- ✅ 完整的请求体、响应体、请求头

### 模式对比

| 功能 | 默认模式 | 完整记录模式 |
|------|---------|-------------|
| 基本信息记录 | ✅ | ✅ |
| 200状态码详情 | ❌ | ✅ |
| 错误状态码详情 | ✅ | ✅ |
| 内存使用 | 低 | 高 |
| 存储空间 | 小 | 大 |
| 调试能力 | 基础 | 完整 |
| 生产环境适用 | ✅ | ❌ |

### 动态切换模式

```bash
# 运行时无法动态切换，需要重启服务
# 停止服务
docker stop privacy-gateway

# 更改配置并重启
docker run -d -p 10805:10805 \
  -e LOG_RECORD_200=true \
  -e ADMIN_SECRET="your-secret" \
  --name privacy-gateway \
  richpeople/privacy-gateway:latest
```

## 高级配置

### 完整环境变量配置

```bash
# 基础配置
export GATEWAY_PORT=10805
export SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id"

# 代理配置
export DEFAULT_PROXY="http://proxy.company.com:8080"
export PROXY_WHITELIST="proxy1.company.com,proxy2.company.com"
export ALLOW_PRIVATE_PROXY=false

# 日志配置
export ADMIN_SECRET="your-secure-admin-secret"
export LOG_MAX_ENTRIES=2000
export LOG_MAX_BODY_SIZE=2048
export LOG_RETENTION_HOURS=48
export LOG_MAX_MEMORY_MB=100.0
export LOG_RECORD_200=true

# 启动服务
./privacy-gateway
```

### Docker配置示例

```bash
# 生产环境配置
docker run -d \
  -p 10805:10805 \
  -e GATEWAY_PORT=10805 \
  -e SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via" \
  -e ADMIN_SECRET="prod-secret-2024" \
  -e LOG_RECORD_200=false \
  -e LOG_MAX_ENTRIES=1000 \
  --name privacy-gateway \
  richpeople/privacy-gateway:latest

# 开发环境配置
docker run -d \
  -p 10805:10805 \
  -e ADMIN_SECRET="dev-secret-123" \
  -e LOG_RECORD_200=true \
  -e LOG_MAX_ENTRIES=2000 \
  -e ALLOW_PRIVATE_PROXY=true \
  --name privacy-gateway-dev \
  richpeople/privacy-gateway:latest
```

## 故障排除

### 常见问题

#### 1. 连接被拒绝

```bash
# 检查服务是否启动
curl http://localhost:10805/

# 检查端口是否被占用
lsof -i :10805

# 检查防火墙设置
sudo ufw status
```

#### 2. 代理认证失败

```bash
# 检查代理配置
curl -v "http://localhost:10805/proxy?target=https://httpbin.org/ip&proxy=http://user:pass@proxy.example.com:8080"

# 测试代理连通性
curl --proxy http://proxy.example.com:8080 https://httpbin.org/ip
```

#### 3. WebSocket连接失败

```bash
# 检查WebSocket升级
curl -i \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Sec-WebSocket-Version: 13" \
  "http://localhost:10805/ws?target=wss://echo.websocket.events"
```

#### 4. 管理界面无法访问

```bash
# 检查管理密钥设置
echo $ADMIN_SECRET

# 测试API接口
curl -H "X-Log-Secret: $ADMIN_SECRET" \
     "http://localhost:10805/logs/api"
```

### 调试模式

```bash
# 启用详细日志
export LOG_LEVEL=debug
./privacy-gateway

# 查看Docker容器日志
docker logs privacy-gateway

# 实时查看日志
docker logs -f privacy-gateway
```

### 性能调优

```bash
# 调整日志配置以优化性能
export LOG_MAX_ENTRIES=500        # 减少内存使用
export LOG_MAX_BODY_SIZE=512      # 减少存储大小
export LOG_RECORD_200=false       # 只记录错误请求
export LOG_RETENTION_HOURS=12     # 减少保留时间

./privacy-gateway
```

## 性能优化

### 内存优化

```bash
# 减少内存使用的配置
export LOG_MAX_ENTRIES=500        # 减少日志条数
export LOG_MAX_BODY_SIZE=512      # 减少响应体大小
export LOG_MAX_MEMORY_MB=25.0     # 限制最大内存使用
export LOG_RECORD_200=false       # 不记录成功请求详情
export LOG_RETENTION_HOURS=6      # 减少保留时间

./privacy-gateway
```

### 高性能配置

```bash
# 高性能生产环境配置
export LOG_MAX_ENTRIES=2000       # 增加缓存容量
export LOG_MAX_BODY_SIZE=2048     # 增加响应体大小
export LOG_MAX_MEMORY_MB=100.0    # 增加内存限制
export LOG_RETENTION_HOURS=24     # 标准保留时间

./privacy-gateway
```

### 监控性能

```bash
# 查看内存使用情况
docker stats privacy-gateway

# 查看日志统计
curl -H "X-Log-Secret: your-secret" \
     "http://localhost:10805/logs/stats"

# 监控响应时间
curl -w "@curl-format.txt" \
     "http://localhost:10805/proxy?target=https://httpbin.org/get"
```

### curl格式文件 (curl-format.txt)

```
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
```

## 集成示例

### 与curl配合使用

```bash
# 创建别名简化使用
alias pgateway='curl -s "http://localhost:10805/proxy?target="'

# 使用别名
pgateway "https://httpbin.org/ip"
```

### 与脚本集成

```bash
#!/bin/bash
# proxy-request.sh

GATEWAY="http://localhost:10805/proxy"
TARGET="$1"

if [ -z "$TARGET" ]; then
    echo "Usage: $0 <target-url>"
    exit 1
fi

curl -s "${GATEWAY}?target=${TARGET}"
```

### 与其他工具集成

```bash
# 与jq配合处理JSON
curl -s "http://localhost:10805/proxy?target=https://httpbin.org/json" | jq .

# 与grep配合过滤内容
curl -s "http://localhost:10805/proxy?target=https://example.com" | grep -i "title"
```
