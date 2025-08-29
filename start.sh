#!/bin/bash

# =============================================================================
# Privacy Gateway 启动脚本
# =============================================================================
#
# 这个脚本包含了所有可用的环境变量配置选项
# 你可以根据需要取消注释并修改相应的配置
#
# 项目地址: https://github.com/your-org/privacy-gateway
# 文档地址: https://docs.privacy-gateway.com
# =============================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 显示启动信息
echo -e "${BLUE}🚀 Privacy Gateway 启动脚本${NC}"
echo -e "${CYAN}===============================================${NC}"
echo ""

# =============================================================================
# 🔐 核心安全配置 (必需)
# =============================================================================
echo -e "${YELLOW}📋 配置说明:${NC}"
echo -e "${GREEN}✅ 持久化存储: 默认启用${NC}"
echo -e "${GREEN}✅ 自动保存: 默认启用 (30秒间隔)${NC}"
echo -e "${GREEN}✅ 配置文件: data/proxy-configs.json${NC}"
echo ""

# 管理员密钥 (必需) - 用于访问管理界面和API
export ADMIN_SECRET=12345678
echo -e "${RED}🔑 管理员密钥: ${ADMIN_SECRET}${NC}"

# =============================================================================
# 🌐 服务器配置
# =============================================================================

# 服务器端口 (默认: 10805)
# export PORT=10805

# 服务器绑定地址 (默认: 0.0.0.0)
# export HOST=0.0.0.0

# 服务器读取超时 (默认: 30秒)
# export SERVER_READ_TIMEOUT=30

# 服务器写入超时 (默认: 30秒)
# export SERVER_WRITE_TIMEOUT=30

# 服务器空闲超时 (默认: 120秒)
# export SERVER_IDLE_TIMEOUT=120

# =============================================================================
# 📊 日志配置
# =============================================================================

# 是否记录200状态码的请求 (默认: false)
export LOG_RECORD_200=true

# 日志级别 (debug, info, warn, error) (默认: info)
# export LOG_LEVEL=info

# 日志格式 (json, text) (默认: json)
# export LOG_FORMAT=json

# 日志文件路径 (默认: 不写入文件，仅控制台输出)
# export LOG_FILE=logs/privacy-gateway.log

# 日志文件最大大小 (MB) (默认: 100)
# export LOG_MAX_SIZE=100

# 日志文件最大备份数 (默认: 3)
# export LOG_MAX_BACKUPS=3

# 日志文件最大保留天数 (默认: 28)
# export LOG_MAX_AGE=28

# =============================================================================
# 🔒 代理安全配置
# =============================================================================

# 是否允许代理私有IP地址 (默认: false)
export ALLOW_PRIVATE_PROXY=true

# 是否允许代理本地回环地址 (默认: false)
# export ALLOW_LOOPBACK_PROXY=false

# 全局速率限制 (请求/分钟) (默认: 0 = 无限制)
# export GLOBAL_RATE_LIMIT=1000

# 单IP速率限制 (请求/分钟) (默认: 0 = 无限制)
# export IP_RATE_LIMIT=100

# 请求超时时间 (秒) (默认: 30)
# export REQUEST_TIMEOUT=30

# 最大请求体大小 (MB) (默认: 10)
# export MAX_REQUEST_BODY_SIZE=10

# =============================================================================
# 💾 数据存储配置
# =============================================================================

# 是否启用持久化存储 (默认: true)
# 设置为 false 将使用内存存储，重启后配置丢失
# export PROXY_CONFIG_PERSIST=false

# 配置文件路径 (默认: data/proxy-configs.json)
# export PROXY_CONFIG_FILE=data/proxy-configs.json

# 是否启用自动保存 (默认: true)
# export PROXY_CONFIG_AUTO_SAVE=true

# 自动保存间隔 (秒) (默认: 30)
# export PROXY_CONFIG_SAVE_INTERVAL=30

# 最大配置数量 (默认: 1000)
# export PROXY_CONFIG_MAX_ENTRIES=1000

# =============================================================================
# 📈 访问日志配置
# =============================================================================

# 访问日志最大条数 (默认: 10000)
# export ACCESS_LOG_MAX_ENTRIES=10000

# 访问日志最大内存使用 (MB) (默认: 100)
# export ACCESS_LOG_MAX_MEMORY=100

# 访问日志记录的最大请求体大小 (KB) (默认: 1)
# export LOG_MAX_BODY_SIZE=1

# 是否记录请求头 (默认: true)
# export LOG_RECORD_HEADERS=true

# 是否记录响应头 (默认: false)
# export LOG_RECORD_RESPONSE_HEADERS=false

# =============================================================================
# 🔧 高级配置
# =============================================================================

# HTTP客户端超时 (秒) (默认: 30)
# export HTTP_CLIENT_TIMEOUT=30

# HTTP客户端最大空闲连接数 (默认: 100)
# export HTTP_CLIENT_MAX_IDLE_CONNS=100

# HTTP客户端每个主机最大空闲连接数 (默认: 10)
# export HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST=10

# HTTP客户端空闲连接超时 (秒) (默认: 90)
# export HTTP_CLIENT_IDLE_CONN_TIMEOUT=90

# 是否跟随重定向 (默认: true)
# export HTTP_CLIENT_FOLLOW_REDIRECTS=true

# 最大重定向次数 (默认: 10)
# export HTTP_CLIENT_MAX_REDIRECTS=10

# 是否验证SSL证书 (默认: true)
# export HTTP_CLIENT_VERIFY_SSL=true

# =============================================================================
# 🌍 CORS配置
# =============================================================================

# 允许的源 (默认: *)
# export CORS_ALLOWED_ORIGINS=*

# 允许的方法 (默认: GET,POST,PUT,DELETE,OPTIONS)
# export CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS

# 允许的头部 (默认: Content-Type,Authorization,X-Requested-With,X-Log-Secret)
# export CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With,X-Log-Secret

# 是否允许凭证 (默认: false)
# export CORS_ALLOW_CREDENTIALS=false

# 预检请求缓存时间 (秒) (默认: 86400)
# export CORS_MAX_AGE=86400

# =============================================================================
# 🔍 监控和健康检查
# =============================================================================

# 是否启用健康检查端点 (默认: true)
# export ENABLE_HEALTH_CHECK=true

# 健康检查路径 (默认: /health)
# export HEALTH_CHECK_PATH=/health

# 是否启用指标收集 (默认: false)
# export ENABLE_METRICS=false

# 指标收集路径 (默认: /metrics)
# export METRICS_PATH=/metrics

# =============================================================================
# 🚀 性能优化
# =============================================================================

# Go运行时最大进程数 (默认: 自动检测CPU核心数)
# export GOMAXPROCS=4

# 垃圾回收目标百分比 (默认: 100)
# export GOGC=100

# 内存限制 (MB) (默认: 无限制)
# export GOMEMLIMIT=512

# =============================================================================
# 🐛 调试配置
# =============================================================================

# 是否启用调试模式 (默认: false)
# export DEBUG=false

# 是否启用性能分析 (默认: false)
# export ENABLE_PPROF=false

# 性能分析端口 (默认: 6060)
# export PPROF_PORT=6060

# 是否打印详细的HTTP请求日志 (默认: false)
# export VERBOSE_HTTP_LOG=false

# =============================================================================
# 🔧 开发环境配置
# =============================================================================

# 开发模式 (启用热重载等) (默认: false)
# export DEV_MODE=false

# 静态文件目录 (默认: 内嵌文件)
# export STATIC_DIR=./static

# 模板文件目录 (默认: 内嵌文件)
# export TEMPLATE_DIR=./templates

# =============================================================================
# 启动服务
# =============================================================================

echo ""
echo -e "${CYAN}===============================================${NC}"
echo -e "${GREEN}🚀 启动 Privacy Gateway 服务...${NC}"
echo ""
echo -e "${YELLOW}📊 当前配置:${NC}"
echo -e "   🔑 管理员密钥: ${ADMIN_SECRET}"
echo -e "   📁 配置文件: ${PROXY_CONFIG_FILE:-data/proxy-configs.json}"
echo -e "   💾 持久化存储: ${PROXY_CONFIG_PERSIST:-true}"
echo -e "   🔄 自动保存: ${PROXY_CONFIG_AUTO_SAVE:-true}"
echo -e "   📊 记录200状态: ${LOG_RECORD_200}"
echo -e "   🔒 允许私有代理: ${ALLOW_PRIVATE_PROXY}"
echo ""
echo -e "${BLUE}🌐 访问地址:${NC}"
echo -e "   📱 管理界面: http://localhost:${PORT:-10805}"
echo -e "   📊 健康检查: http://localhost:${PORT:-10805}/health"
echo -e "   📋 API文档: http://localhost:${PORT:-10805}/docs"
echo ""
echo -e "${PURPLE}💡 提示:${NC}"
echo -e "   • 按 Ctrl+C 停止服务"
echo -e "   • 查看日志了解运行状态"
echo -e "   • 修改配置后重启服务生效"
echo ""

# 检查可执行文件是否存在
if [ ! -f "./privacy-gateway" ]; then
    echo -e "${RED}❌ 错误: 找不到 privacy-gateway 可执行文件${NC}"
    echo -e "${YELLOW}💡 请先运行: go build -o privacy-gateway .${NC}"
    exit 1
fi

# 创建必要的目录
mkdir -p data logs

# 启动服务
echo -e "${GREEN}🎯 正在启动服务...${NC}"
./privacy-gateway
