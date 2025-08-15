# 隐私网关 (Privacy Gateway)

这是一个简单的反向代理服务，旨在通过隐藏原始请求中的敏感头信息来增强隐私。

## 构建与运行

本项目支持两种构建和运行方式：本地构建和 Docker 容器化部署。

### 方式一：本地构建和运行

#### 前置要求

- Go 1.19 或更高版本
- Git（用于克隆代码）

#### 1. 克隆项目

```bash
git clone <repository-url>
cd PrivacyGateway
```

#### 2. 本地构建

在项目根目录下运行：

```bash
# 构建可执行文件
go build -o privacy-gateway .

# 或者直接运行（不生成可执行文件）
go run .
```

#### 3. 本地运行

##### 基础运行（使用默认配置）

```bash
# 运行构建好的可执行文件
./privacy-gateway

# 或者直接运行源码
go run .
```

服务将在 `http://localhost:8080` 启动。

##### 自定义配置运行

```bash
# 设置环境变量并运行
export GATEWAY_PORT=9090
export SENSITIVE_HEADERS="cf-,x-forwarded,proxy"
export ADMIN_SECRET="my-admin-secret"
./privacy-gateway

# 或者一行命令设置环境变量
GATEWAY_PORT=8080 ADMIN_SECRET="my-admin-secret" ./privacy-gateway
```



### 方式二：Docker 容器化部署

#### 1. 构建 Docker 镜像

在项目根目录下，运行以下命令来构建镜像：

```bash
docker build -t privacygateway .
```

#### 2. 运行 Docker 容器

使用以下命令来启动容器。这会将容器的 `8080` 端口映射到你本机的 `8080` 端口。

```bash
docker run -d -p 8080:8080 --name privacy-gateway-container privacygateway
```

## 如何使用

网关通过 `/proxy` 端点提供服务。你需要通过 `target` 查询参数来指定你想要访问的目标 URL。

### 基本格式

```
http://localhost:8080/proxy?target=<你的目标URL>
```

### 使用示例

网关会自动转发原始请求的 HTTP 方法（GET, POST, PUT, DELETE 等）以及请求体（Body）。

#### 1. GET 请求 (获取数据)

这是最常见的用法，用于从服务器获取数据。

```bash
# 示例: 获取你的 IP 地址信息
curl "http://localhost:8080/proxy?target=http://iprust.io/ip.json"

# 示例: 获取一个网页
curl "http://localhost:8080/proxy?target=http://example.com"
```

#### 2. POST 请求 (提交数据)

用于向服务器提交数据，例如发送一个 JSON 对象。

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"test", "value":123}' \
  "http://localhost:8080/proxy?target=https://httpbin.org/post"
```
*httpbin.org 会将你发送的数据原样返回在响应的 `json` 字段中。*

#### 3. PUT 请求 (更新数据)

用于更新或替换服务器上的资源。

```bash
curl -X PUT \
  -H "Content-Type: application/json" \
  -d '{"id": 1, "status":"updated"}' \
  "http://localhost:8080/proxy?target=https://httpbin.org/put"
```

#### 4. DELETE 请求 (删除数据)

用于请求服务器删除指定的资源。

```bash
curl -X DELETE "http://localhost:8080/proxy?target=https://httpbin.org/delete"
```

#### 5. WebSocket 代理

网关还支持代理 WebSocket 连接，使用 `/ws` 端点。

**连接格式:**
```
ws://localhost:8080/ws?target=<你的目标ws或wss地址>
```

**示例:**
连接到一个公共的回显服务器。
```bash
# 你可以使用任何支持 WebSocket 的客户端工具, 例如 wscat
# npm install -g wscat
wscat -c "ws://localhost:8080/ws?target=wss://echo.websocket.events"
```
连接成功后，你发送的任何消息都会被服务器原样返回。

## 配置参数

隐私网关支持通过环境变量进行配置，以下是所有可用的配置选项：

### 基础配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `GATEWAY_PORT` | `8080` | 网关监听的端口号 |
| `SENSITIVE_HEADERS` | `cf-,x-forwarded,proxy,via,x-request-id,x-trace,x-correlation-id,x-country,x-region,x-city` | 要过滤的敏感头信息关键字，用逗号分隔 |

### 代理配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `DEFAULT_PROXY` | 无 | 默认上游代理服务器URL，支持 HTTP/HTTPS/SOCKS5 |
| `PROXY_WHITELIST` | 无 | 允许使用的代理服务器白名单，用逗号分隔 |
| `ALLOW_PRIVATE_PROXY` | `false` | 是否允许使用私有IP地址的代理（用于开发测试） |

### 管理功能配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `ADMIN_SECRET` | 无 | 管理功能访问密钥，设置后可通过 `/logs` 端点查看访问日志等管理功能 |
| `LOG_MAX_ENTRIES` | `1000` | 内存中保存的最大日志条数 |
| `LOG_MAX_BODY_SIZE` | `1024` | 记录的响应体最大大小（字节） |
| `LOG_RETENTION_HOURS` | `24` | 日志保留时间（小时） |
| `LOG_MAX_MEMORY_MB` | `50.0` | 日志功能最大内存使用量（MB） |

### 配置示例

#### 本地运行配置

```bash
# 基础配置 - Linux/macOS
export GATEWAY_PORT=8080
export SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via"
./privacy-gateway

# 完整配置 - Linux/macOS
export GATEWAY_PORT=8080
export SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id"
export DEFAULT_PROXY="http://proxy.example.com:8080"
export PROXY_WHITELIST="proxy1.example.com,proxy2.example.com"
export ADMIN_SECRET="your-admin-secret"
export LOG_MAX_ENTRIES=2000
./privacy-gateway

# 一行命令运行 - Linux/macOS
GATEWAY_PORT=8080 ADMIN_SECRET="your-admin-secret" ./privacy-gateway

# Windows 配置
set GATEWAY_PORT=8080
set SENSITIVE_HEADERS=cf-,x-forwarded,proxy,via
set ADMIN_SECRET=your-admin-secret
privacy-gateway.exe
```

#### Docker 运行配置

```bash
# 基础配置
docker run -d \
  -p 8080:8080 \
  -e GATEWAY_PORT=8080 \
  -e SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via" \
  --name privacy-gateway \
  privacygateway

# 带代理和日志功能的完整配置
docker run -d \
  -p 8080:8080 \
  -e GATEWAY_PORT=8080 \
  -e SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id" \
  -e DEFAULT_PROXY="http://proxy.example.com:8080" \
  -e PROXY_WHITELIST="proxy1.example.com,proxy2.example.com" \
  -e ADMIN_SECRET="your-admin-secret" \
  -e LOG_MAX_ENTRIES=2000 \
  --name privacy-gateway \
  privacygateway
```

## 高级用法

### 代理配置

除了通过环境变量设置默认代理外，还可以通过以下方式为单个请求指定代理：

#### 1. 查询参数方式

```bash
# 使用 HTTP 代理
curl "http://localhost:8080/proxy?target=http://example.com&proxy=http://proxy.example.com:8080"

# 使用 SOCKS5 代理（带认证）
curl "http://localhost:8080/proxy?target=http://example.com&proxy=socks5://user:pass@proxy.example.com:1080"
```

#### 2. 请求头方式

通过 `X-Proxy-Config` 请求头传递 Base64 编码的 JSON 配置：

```bash
# 创建代理配置 JSON
proxy_config='{"url":"http://proxy.example.com:8080","timeout":30}'

# Base64 编码
proxy_config_b64=$(echo -n "$proxy_config" | base64)

# 发送请求
curl -H "X-Proxy-Config: $proxy_config_b64" \
     "http://localhost:8080/proxy?target=http://example.com"
```

代理配置 JSON 格式：
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

### 支持的代理协议

- **HTTP/HTTPS**: `http://proxy.example.com:8080`
- **SOCKS5**: `socks5://proxy.example.com:1080`
- **带认证的代理**: `http://user:pass@proxy.example.com:8080`

### 日志查看

如果设置了 `ADMIN_SECRET` 环境变量，可以通过以下方式查看访问日志：

```bash
# 访问日志查看页面
http://localhost:8080/logs?secret=your-admin-secret
```

日志功能提供：
- 实时访问日志查看
- 按域名、状态码、时间范围筛选
- 请求和响应详情查看
- 自动刷新和分页功能

### 注意事项

- **URL 编码**: 如果你的目标 URL 包含特殊字符（如 `&`），请确保对其进行正确的 URL 编码。
- **HTTPS**: 网关会直接请求你在 `target` 参数中提供的 URL，无论是 `http` 还是 `https`。
- **代理优先级**: 请求头配置 > 查询参数配置 > 默认环境变量配置
- **安全性**: 生产环境中建议设置代理白名单，避免被滥用
