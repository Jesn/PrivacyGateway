#!/bin/bash

# 性能测试脚本
# 使用方法: ./scripts/test_performance.sh [base_url] [admin_secret] [config_id]

set -e

# 默认参数
BASE_URL=${1:-"http://localhost:10805"}
ADMIN_SECRET=${2:-"12345678"}
CONFIG_ID=${3:-""}

echo "🚀 Privacy Gateway 性能测试开始..."
echo "Base URL: $BASE_URL"
echo "Admin Secret: $ADMIN_SECRET"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 创建临时文件
TEMP_DIR=$(mktemp -d)
RESULTS_FILE="$TEMP_DIR/results.txt"
TOKENS_FILE="$TEMP_DIR/tokens.txt"

cleanup() {
    echo "清理临时文件..."
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# 性能测试函数
performance_test() {
    local description=$1
    local url=$2
    local method=${3:-"GET"}
    local headers=$4
    local data=$5
    local concurrent=${6:-10}
    local requests=${7:-100}
    
    echo -e "${YELLOW}性能测试: $description${NC}"
    echo "并发数: $concurrent, 总请求数: $requests"
    
    # 使用ab进行性能测试
    if command -v ab >/dev/null 2>&1; then
        if [ -n "$data" ]; then
            # POST请求
            echo "$data" > "$TEMP_DIR/post_data"
            ab_result=$(ab -n $requests -c $concurrent -T "application/json" -H "$headers" -p "$TEMP_DIR/post_data" "$url" 2>/dev/null | grep -E "(Requests per second|Time per request|Transfer rate)")
        else
            # GET请求
            ab_result=$(ab -n $requests -c $concurrent -H "$headers" "$url" 2>/dev/null | grep -E "(Requests per second|Time per request|Transfer rate)")
        fi
        echo "$ab_result"
    else
        echo "Apache Bench (ab) 未安装，使用curl进行简单测试..."
        
        start_time=$(date +%s.%N)
        for i in $(seq 1 $requests); do
            if [ -n "$data" ]; then
                curl -s -X "$method" -H "$headers" -H "Content-Type: application/json" -d "$data" "$url" >/dev/null
            else
                curl -s -H "$headers" "$url" >/dev/null
            fi
        done
        end_time=$(date +%s.%N)
        
        duration=$(echo "$end_time - $start_time" | bc)
        rps=$(echo "scale=2; $requests / $duration" | bc)
        avg_time=$(echo "scale=2; $duration * 1000 / $requests" | bc)
        
        echo "总耗时: ${duration}s"
        echo "平均响应时间: ${avg_time}ms"
        echo "每秒请求数: ${rps} req/s"
    fi
    
    echo ""
}

# 1. 创建测试配置（如果未提供）
if [ -z "$CONFIG_ID" ]; then
    echo -e "${BLUE}=== 创建测试配置 ===${NC}"
    config_response=$(curl -s -X POST \
        -H "X-Log-Secret: $ADMIN_SECRET" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "性能测试配置",
            "subdomain": "perf-test",
            "target_url": "https://httpbin.org",
            "protocol": "https",
            "enabled": true
        }' \
        "$BASE_URL/config/proxy")
    
    CONFIG_ID=$(echo "$config_response" | jq -r '.id // empty')
    
    if [ -z "$CONFIG_ID" ] || [ "$CONFIG_ID" = "null" ]; then
        echo -e "${RED}❌ 无法创建测试配置${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 测试配置创建成功: $CONFIG_ID${NC}"
    echo ""
fi

# 2. 创建测试令牌
echo -e "${BLUE}=== 创建测试令牌 ===${NC}"
token_response=$(curl -s -X POST \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "性能测试令牌",
        "description": "用于性能测试的令牌"
    }' \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

TOKEN_VALUE=$(echo "$token_response" | jq -r '.data.token // .token // empty')

if [ -z "$TOKEN_VALUE" ] || [ "$TOKEN_VALUE" = "null" ]; then
    echo -e "${RED}❌ 无法创建测试令牌${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 测试令牌创建成功${NC}"
echo "令牌值: $TOKEN_VALUE"
echo ""

# 3. API性能测试
echo -e "${BLUE}=== API性能测试 ===${NC}"

# 3.1 令牌列表API性能
performance_test "令牌列表API" \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens" \
    "GET" \
    "X-Log-Secret: $ADMIN_SECRET" \
    "" \
    5 \
    50

# 3.2 令牌创建API性能
performance_test "令牌创建API" \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens" \
    "POST" \
    "X-Log-Secret: $ADMIN_SECRET" \
    '{"name":"性能测试令牌","description":"性能测试"}' \
    3 \
    20

# 3.3 配置列表API性能
performance_test "配置列表API" \
    "$BASE_URL/config/proxy" \
    "GET" \
    "X-Log-Secret: $ADMIN_SECRET" \
    "" \
    10 \
    100

# 4. 代理性能测试
echo -e "${BLUE}=== 代理性能测试 ===${NC}"

# 4.1 管理员密钥认证代理
performance_test "管理员密钥代理" \
    "$BASE_URL/proxy?target=https://httpbin.org/get" \
    "GET" \
    "X-Log-Secret: $ADMIN_SECRET" \
    "" \
    5 \
    30

# 4.2 令牌认证代理
performance_test "令牌认证代理" \
    "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "GET" \
    "X-Proxy-Token: $TOKEN_VALUE" \
    "" \
    5 \
    30

# 4.3 Bearer令牌认证代理
performance_test "Bearer令牌代理" \
    "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "GET" \
    "Authorization: Bearer $TOKEN_VALUE" \
    "" \
    5 \
    30

# 5. 认证性能测试
echo -e "${BLUE}=== 认证性能测试 ===${NC}"

# 5.1 测试认证失败的性能
performance_test "无效令牌认证" \
    "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "GET" \
    "X-Proxy-Token: invalid-token" \
    "" \
    10 \
    100

# 5.2 测试无认证信息的性能
performance_test "无认证信息" \
    "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "GET" \
    "" \
    "" \
    10 \
    100

# 6. 并发令牌创建测试
echo -e "${BLUE}=== 并发令牌创建测试 ===${NC}"

concurrent_token_test() {
    local concurrent=$1
    local total_tokens=$2
    
    echo -e "${YELLOW}并发令牌创建测试: $concurrent 并发, $total_tokens 个令牌${NC}"
    
    start_time=$(date +%s.%N)
    
    # 并发创建令牌
    for i in $(seq 1 $concurrent); do
        {
            for j in $(seq 1 $((total_tokens / concurrent))); do
                curl -s -X POST \
                    -H "X-Log-Secret: $ADMIN_SECRET" \
                    -H "Content-Type: application/json" \
                    -d "{\"name\":\"并发令牌-$i-$j\",\"description\":\"并发测试令牌\"}" \
                    "$BASE_URL/config/proxy/$CONFIG_ID/tokens" >/dev/null
            done
        } &
    done
    
    wait
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    tokens_per_second=$(echo "scale=2; $total_tokens / $duration" | bc)
    
    echo "总耗时: ${duration}s"
    echo "令牌创建速率: ${tokens_per_second} tokens/s"
    echo ""
}

concurrent_token_test 5 25
concurrent_token_test 10 50

# 7. 内存和资源使用测试
echo -e "${BLUE}=== 资源使用测试 ===${NC}"

if command -v ps >/dev/null 2>&1; then
    echo "当前系统资源使用情况:"
    echo "内存使用:"
    free -h 2>/dev/null || vm_stat | head -5
    echo ""
    echo "CPU使用:"
    top -l 1 | head -10 2>/dev/null || ps aux | head -10
    echo ""
fi

# 8. 生成性能报告
echo -e "${BLUE}=== 性能测试总结 ===${NC}"

echo "测试完成时间: $(date)"
echo "测试配置ID: $CONFIG_ID"
echo "测试令牌: $TOKEN_VALUE"
echo ""

echo -e "${GREEN}🎉 性能测试完成！${NC}"
echo ""
echo "性能测试要点:"
echo "- ✅ API响应时间测试"
echo "- ✅ 代理请求性能测试"
echo "- ✅ 认证性能测试"
echo "- ✅ 并发操作测试"
echo "- ✅ 资源使用监控"
echo ""

echo "性能优化建议:"
echo "1. 如果API响应时间 > 100ms，考虑优化数据库查询"
echo "2. 如果代理请求时间 > 500ms，检查网络连接和目标服务器"
echo "3. 如果认证时间 > 10ms，考虑添加令牌缓存"
echo "4. 如果并发性能不佳，考虑增加连接池大小"
echo "5. 监控内存使用，避免内存泄漏"
echo ""

# 清理测试数据
echo -e "${YELLOW}清理测试数据...${NC}"
curl -s -X DELETE \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    "$BASE_URL/config/proxy/$CONFIG_ID" >/dev/null

echo -e "${GREEN}✅ 测试数据清理完成${NC}"
