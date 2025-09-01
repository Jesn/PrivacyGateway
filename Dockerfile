# ---- Build Stage ----
# 使用官方的 Go 镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件并下载依赖（利用Docker缓存层）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 创建数据目录
RUN mkdir -p data

# 构建优化的二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags "-s -w -extldflags '-static'" \
    -tags netgo \
    -trimpath \
    -o privacy-gateway .

# ---- Final Stage ----
# 使用 Distroless 镜像，它包含了 CA 证书并且为静态 Go 应用提供了更好的兼容性
FROM gcr.io/distroless/static-debian11

# 添加镜像标签
LABEL maintainer="Privacy Gateway Team" \
      description="Lightweight reverse proxy with privacy protection" \
      version="1.0" \
      org.opencontainers.image.source="https://github.com/username/PrivacyGateway"

# 设置工作目录
WORKDIR /app

# 创建非root用户（Distroless镜像中的nobody用户）
USER 65534:65534

# 从构建阶段复制编译好的二进制文件和静态文件
COPY --from=builder --chown=65534:65534 /app/privacy-gateway .
COPY --from=builder --chown=65534:65534 /app/index.html .
COPY --from=builder --chown=65534:65534 /app/data ./data

# 暴露网关将要监听的端口 (默认 10805)
EXPOSE 10805

# 设置默认的环境变量
# 用户可以在 docker run 时覆盖这些变量
ENV GATEWAY_PORT=10805
ENV SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id,x-trace,x-correlation-id,x-country,x-region,x-city"

# 持久化存储配置
ENV PROXY_CONFIG_PERSIST=true
ENV PROXY_CONFIG_FILE=/app/data/proxy-configs.json
ENV PROXY_CONFIG_AUTO_SAVE=true

# 日志相关配置（可选）
# ENV ADMIN_SECRET=""
# ENV LOG_MAX_ENTRIES=1000
# ENV LOG_MAX_BODY_SIZE=1024
# ENV LOG_RETENTION_HOURS=24
# ENV LOG_MAX_MEMORY_MB=50.0
# ENV LOG_RECORD_200=false

# 代理相关配置（可选）
# ENV DEFAULT_PROXY=""
# ENV PROXY_WHITELIST=""
# ENV ALLOW_PRIVATE_PROXY=false

# 创建数据卷
VOLUME ["/app/data"]

# 健康检查 - 由于使用Distroless镜像，健康检查需要在docker-compose中配置

# 容器启动时运行的命令
CMD ["./privacy-gateway"]
