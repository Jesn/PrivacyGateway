# 隐私网关 (Privacy Gateway)

这是一个简单的反向代理服务，旨在通过隐藏原始请求中的敏感头信息来增强隐私。

## 构建与运行

本项目使用 Docker 进行构建和部署。

### 1. 构建 Docker 镜像

在项目根目录下，运行以下命令来构建镜像：

```bash
docker build -t privacygateway .
```

### 2. 运行 Docker 容器

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

### 注意事项

- **URL 编码**: 如果你的目标 URL 包含特殊字符（如 `&`），请确保对其进行正确的 URL 编码。
- **HTTPS**: 网关会直接请求你在 `target` 参数中提供的 URL，无论是 `http` 还是 `https`。
