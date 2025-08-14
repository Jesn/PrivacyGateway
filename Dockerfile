# ---- Build Stage ----
# 使用官方的 Go 镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum (如果存在) 以便缓存依赖
# 由于我们没有外部依赖，这一步可以简化，但保留是个好习惯
# COPY go.mod go.sum ./
# RUN go mod download

# 复制所有源代码
COPY . .

# 构建一个静态链接的、移除调试信息的二进制文件
# CGO_ENABLED=0 确保静态链接
# -ldflags "-s -w" 移除调试符号和 DWARF 信息，减小体积
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -o privacy-gateway .

# ---- Final Stage ----
# 使用 Distroless 镜像，它包含了 CA 证书并且为静态 Go 应用提供了更好的兼容性
FROM gcr.io/distroless/static-debian11

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件和静态文件
COPY --from=builder /app/privacy-gateway .
COPY --from=builder /app/index.html .

# 暴露网关将要监听的端口 (默认 8080)
# 这只是元数据，实际端口由 GATEWAY_PORT 环境变量决定
EXPOSE 8080

# 设置默认的环境变量
# 用户可以在 docker run 时覆盖这些变量
ENV GATEWAY_PORT=8080
ENV SENSITIVE_HEADERS="cf-,x-forwarded,proxy,via,x-request-id,x-trace,x-correlation-id,x-country,x-region,x-city"

# 容器启动时运行的命令
CMD ["./privacy-gateway"]
