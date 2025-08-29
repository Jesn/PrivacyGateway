#!/bin/bash

# =============================================================================
# Privacy Gateway 配置示例脚本
# =============================================================================
# 
# 这个文件包含了不同使用场景的配置示例
# 你可以复制相应的配置到 start.sh 中使用
# =============================================================================

echo "🔧 Privacy Gateway 配置示例"
echo "=========================================="
echo ""

# =============================================================================
# 示例1: 开发环境配置
# =============================================================================
echo "📝 示例1: 开发环境配置"
echo "适用于: 本地开发、测试"
echo ""
cat << 'EOF'
# 开发环境配置
export ADMIN_SECRET=dev123456
export LOG_RECORD_200=true
export ALLOW_PRIVATE_PROXY=true
export DEBUG=true
export LOG_LEVEL=debug
export DEV_MODE=true
export VERBOSE_HTTP_LOG=true
EOF
echo ""

# =============================================================================
# 示例2: 生产环境配置
# =============================================================================
echo "📝 示例2: 生产环境配置"
echo "适用于: 生产部署、高性能"
echo ""
cat << 'EOF'
# 生产环境配置
export ADMIN_SECRET=your-super-secure-secret-key-here
export LOG_RECORD_200=false
export ALLOW_PRIVATE_PROXY=false
export LOG_LEVEL=warn
export GLOBAL_RATE_LIMIT=10000
export IP_RATE_LIMIT=1000
export HTTP_CLIENT_MAX_IDLE_CONNS=200
export GOMAXPROCS=8
export GOMEMLIMIT=1024
export ENABLE_METRICS=true
EOF
echo ""

# =============================================================================
# 示例3: 高安全性配置
# =============================================================================
echo "📝 示例3: 高安全性配置"
echo "适用于: 企业环境、严格安全要求"
echo ""
cat << 'EOF'
# 高安全性配置
export ADMIN_SECRET=ultra-secure-admin-key-2024
export ALLOW_PRIVATE_PROXY=false
export ALLOW_LOOPBACK_PROXY=false
export HTTP_CLIENT_VERIFY_SSL=true
export GLOBAL_RATE_LIMIT=5000
export IP_RATE_LIMIT=500
export MAX_REQUEST_BODY_SIZE=5
export LOG_RECORD_HEADERS=true
export LOG_RECORD_RESPONSE_HEADERS=true
export CORS_ALLOWED_ORIGINS=https://yourdomain.com
export CORS_ALLOW_CREDENTIALS=false
EOF
echo ""

# =============================================================================
# 示例4: 高性能配置
# =============================================================================
echo "📝 示例4: 高性能配置"
echo "适用于: 高并发、大流量场景"
echo ""
cat << 'EOF'
# 高性能配置
export ADMIN_SECRET=performance-optimized-key
export LOG_RECORD_200=false
export ACCESS_LOG_MAX_ENTRIES=50000
export ACCESS_LOG_MAX_MEMORY=500
export HTTP_CLIENT_MAX_IDLE_CONNS=500
export HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST=50
export HTTP_CLIENT_IDLE_CONN_TIMEOUT=300
export SERVER_READ_TIMEOUT=60
export SERVER_WRITE_TIMEOUT=60
export SERVER_IDLE_TIMEOUT=300
export GOMAXPROCS=16
export GOGC=50
export GOMEMLIMIT=2048
EOF
echo ""

# =============================================================================
# 示例5: 调试和监控配置
# =============================================================================
echo "📝 示例5: 调试和监控配置"
echo "适用于: 问题排查、性能分析"
echo ""
cat << 'EOF'
# 调试和监控配置
export ADMIN_SECRET=debug-monitoring-key
export DEBUG=true
export LOG_LEVEL=debug
export LOG_RECORD_200=true
export VERBOSE_HTTP_LOG=true
export ENABLE_PPROF=true
export PPROF_PORT=6060
export ENABLE_METRICS=true
export METRICS_PATH=/metrics
export LOG_FILE=logs/privacy-gateway.log
export LOG_MAX_SIZE=50
export LOG_MAX_BACKUPS=10
export LOG_MAX_AGE=7
EOF
echo ""

# =============================================================================
# 示例6: 内存存储配置
# =============================================================================
echo "📝 示例6: 内存存储配置"
echo "适用于: 临时使用、无持久化需求"
echo ""
cat << 'EOF'
# 内存存储配置
export ADMIN_SECRET=memory-only-key
export PROXY_CONFIG_PERSIST=false
export LOG_RECORD_200=true
export ALLOW_PRIVATE_PROXY=true
export ACCESS_LOG_MAX_ENTRIES=1000
export ACCESS_LOG_MAX_MEMORY=50
EOF
echo ""

# =============================================================================
# 示例7: 自定义端口和路径配置
# =============================================================================
echo "📝 示例7: 自定义端口和路径配置"
echo "适用于: 多实例部署、端口冲突"
echo ""
cat << 'EOF'
# 自定义端口和路径配置
export ADMIN_SECRET=custom-port-key
export PORT=8080
export HOST=127.0.0.1
export PROXY_CONFIG_FILE=/etc/privacy-gateway/configs.json
export LOG_FILE=/var/log/privacy-gateway/app.log
export HEALTH_CHECK_PATH=/status
export METRICS_PATH=/stats
EOF
echo ""

# =============================================================================
# 示例8: Docker容器配置
# =============================================================================
echo "📝 示例8: Docker容器配置"
echo "适用于: Docker部署、容器化环境"
echo ""
cat << 'EOF'
# Docker容器配置
export ADMIN_SECRET=${ADMIN_SECRET:-docker-default-key}
export PORT=${PORT:-10805}
export HOST=0.0.0.0
export LOG_FORMAT=json
export LOG_LEVEL=${LOG_LEVEL:-info}
export PROXY_CONFIG_FILE=/app/data/configs.json
export ACCESS_LOG_MAX_ENTRIES=20000
export GOMAXPROCS=${GOMAXPROCS:-4}
export GOMEMLIMIT=${GOMEMLIMIT:-512}
EOF
echo ""

# =============================================================================
# 使用说明
# =============================================================================
echo "📖 使用说明:"
echo "1. 选择适合你场景的配置示例"
echo "2. 复制相应的环境变量到 start.sh 中"
echo "3. 根据实际需求调整参数值"
echo "4. 运行 ./start.sh 启动服务"
echo ""
echo "💡 提示:"
echo "• 生产环境请务必修改默认的 ADMIN_SECRET"
echo "• 根据服务器性能调整 GOMAXPROCS 和 GOMEMLIMIT"
echo "• 启用监控和日志有助于问题排查"
echo "• 高并发场景建议调整连接池和超时参数"
echo ""
echo "🔗 更多信息:"
echo "• 配置文档: https://docs.privacy-gateway.com/config"
echo "• 性能调优: https://docs.privacy-gateway.com/performance"
echo "• 安全指南: https://docs.privacy-gateway.com/security"
