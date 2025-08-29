# Privacy Gateway 启动指南

## 🚀 快速开始

### 1. 编译项目
```bash
go build -o privacy-gateway .
```

### 2. 启动服务
```bash
./start.sh
```

### 3. 访问管理界面
打开浏览器访问：http://localhost:10805

## 📁 文件说明

### `start.sh` - 主启动脚本
- 包含所有可用的环境变量配置选项
- 提供详细的配置说明和默认值
- 彩色输出，显示当前配置状态
- 自动检查可执行文件和创建必要目录

### `start-examples.sh` - 配置示例脚本
- 提供8种不同场景的配置示例
- 包含开发、生产、高安全性、高性能等配置
- 运行查看所有配置示例：`./start-examples.sh`

## ⚙️ 配置分类

### 🔐 核心安全配置
- `ADMIN_SECRET` - 管理员密钥（必需）
- `ALLOW_PRIVATE_PROXY` - 是否允许代理私有IP
- `GLOBAL_RATE_LIMIT` - 全局速率限制

### 🌐 服务器配置
- `PORT` - 服务器端口（默认：10805）
- `HOST` - 绑定地址（默认：0.0.0.0）
- `SERVER_*_TIMEOUT` - 各种超时设置

### 📊 日志配置
- `LOG_RECORD_200` - 是否记录200状态码
- `LOG_LEVEL` - 日志级别
- `LOG_FILE` - 日志文件路径

### 💾 数据存储配置
- `PROXY_CONFIG_PERSIST` - 持久化存储（默认：true）
- `PROXY_CONFIG_FILE` - 配置文件路径
- `PROXY_CONFIG_AUTO_SAVE` - 自动保存（默认：true）

### 🔧 高级配置
- `HTTP_CLIENT_*` - HTTP客户端设置
- `CORS_*` - 跨域配置
- `GOMAXPROCS` - Go运行时配置

## 🎯 使用场景

### 开发环境
```bash
# 复制到 start.sh 中
export ADMIN_SECRET=dev123456
export DEBUG=true
export LOG_LEVEL=debug
export VERBOSE_HTTP_LOG=true
```

### 生产环境
```bash
# 复制到 start.sh 中
export ADMIN_SECRET=your-super-secure-secret-key
export LOG_LEVEL=warn
export GLOBAL_RATE_LIMIT=10000
export ENABLE_METRICS=true
```

### 高性能场景
```bash
# 复制到 start.sh 中
export HTTP_CLIENT_MAX_IDLE_CONNS=500
export GOMAXPROCS=16
export GOMEMLIMIT=2048
```

## 📋 配置检查清单

### 部署前检查
- [ ] 修改默认的 `ADMIN_SECRET`
- [ ] 设置合适的 `LOG_LEVEL`
- [ ] 配置速率限制 `GLOBAL_RATE_LIMIT`
- [ ] 检查端口 `PORT` 是否可用
- [ ] 确认存储配置 `PROXY_CONFIG_FILE`

### 性能优化
- [ ] 根据CPU核心数设置 `GOMAXPROCS`
- [ ] 设置内存限制 `GOMEMLIMIT`
- [ ] 调整连接池大小 `HTTP_CLIENT_MAX_IDLE_CONNS`
- [ ] 配置超时参数 `REQUEST_TIMEOUT`

### 安全加固
- [ ] 禁用私有IP代理 `ALLOW_PRIVATE_PROXY=false`
- [ ] 启用SSL验证 `HTTP_CLIENT_VERIFY_SSL=true`
- [ ] 设置CORS限制 `CORS_ALLOWED_ORIGINS`
- [ ] 配置IP白名单/黑名单

## 🔍 监控和调试

### 启用监控
```bash
export ENABLE_METRICS=true
export METRICS_PATH=/metrics
```
访问：http://localhost:10805/metrics

### 启用调试
```bash
export DEBUG=true
export ENABLE_PPROF=true
export PPROF_PORT=6060
```
访问：http://localhost:6060/debug/pprof/

### 日志文件
```bash
export LOG_FILE=logs/privacy-gateway.log
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=3
```

## 🐳 Docker 部署

### Dockerfile 示例
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o privacy-gateway .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/privacy-gateway .
COPY --from=builder /app/start.sh .
RUN chmod +x start.sh
EXPOSE 10805
CMD ["./start.sh"]
```

### Docker Compose 示例
```yaml
version: '3.8'
services:
  privacy-gateway:
    build: .
    ports:
      - "10805:10805"
    environment:
      - ADMIN_SECRET=your-secret-key
      - LOG_LEVEL=info
      - PROXY_CONFIG_FILE=/app/data/configs.json
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
```

## 🆘 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   export PORT=8080  # 使用其他端口
   ```

2. **配置文件权限问题**
   ```bash
   mkdir -p data logs
   chmod 755 data logs
   ```

3. **内存不足**
   ```bash
   export GOMEMLIMIT=512  # 限制内存使用
   export GOGC=50         # 更频繁的垃圾回收
   ```

4. **连接超时**
   ```bash
   export REQUEST_TIMEOUT=60
   export HTTP_CLIENT_TIMEOUT=60
   ```

### 日志分析
```bash
# 查看错误日志
grep "error" logs/privacy-gateway.log

# 查看访问统计
grep "forwarding request" logs/privacy-gateway.log | wc -l

# 监控内存使用
grep "memory" logs/privacy-gateway.log
```

## 📞 支持

- 🐛 问题报告：[GitHub Issues](https://github.com/your-org/privacy-gateway/issues)
- 📖 文档：[在线文档](https://docs.privacy-gateway.com)
- 💬 讨论：[GitHub Discussions](https://github.com/your-org/privacy-gateway/discussions)

---

**注意**：生产环境部署前请仔细阅读安全指南，确保正确配置所有安全参数。
