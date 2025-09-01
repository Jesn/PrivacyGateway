#!/bin/bash

# 代理请求认证集成测试脚本
# 使用方法: ./scripts/test_proxy_integration.sh [base_url] [admin_secret]

set -e

# 默认参数
BASE_URL=${1:-"http://localhost:10805"}
ADMIN_SECRET=${2:-"12345678"}

echo "🔧 代理请求认证集成测试开始..."
echo "Base URL: $BASE_URL"
echo "Admin Secret: $ADMIN_SECRET"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_request() {
    local description=$1
    local method=$2
    local endpoint=$3
    local headers=$4
    local expected_status=$5
    
    echo -e "${YELLOW}测试: $description${NC}"
    echo "请求: $method $endpoint"
    
    response=$(curl -s -w "\n%{http_code}" -X "$method" $headers "$BASE_URL$endpoint")
    
    # 分离响应体和状态码
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    echo "状态码: $status"
    if [ ${#body} -gt 200 ]; then
        echo "响应: ${body:0:200}..."
    else
        echo "响应: $body"
    fi
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✅ 测试通过${NC}"
        return 0
    else
        echo -e "${RED}❌ 测试失败 (期望状态码: $expected_status)${NC}"
        return 1
    fi
    
    echo ""
}

echo -e "${BLUE}=== 第一步：创建测试配置 ===${NC}"

# 1. 创建代理配置
echo "创建代理配置..."
config_response=$(curl -s -X POST \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "集成测试配置",
        "subdomain": "integration-test",
        "target_url": "https://httpbin.org",
        "protocol": "https",
        "enabled": true
    }' \
    "$BASE_URL/config/proxy")

echo "配置创建响应: $config_response"

# 提取配置ID
CONFIG_ID=$(echo "$config_response" | jq -r '.data.id // empty')

if [ -z "$CONFIG_ID" ] || [ "$CONFIG_ID" = "null" ]; then
    echo -e "${RED}❌ 无法提取配置ID，配置创建可能失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 配置创建成功${NC}"
echo "配置ID: $CONFIG_ID"
echo ""

echo -e "${BLUE}=== 第二步：创建访问令牌 ===${NC}"

# 2. 创建访问令牌
echo "创建访问令牌..."
token_response=$(curl -s -X POST \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "集成测试令牌",
        "description": "用于集成测试的访问令牌"
    }' \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

echo "令牌创建响应: $token_response"

# 提取令牌值
TOKEN_VALUE=$(echo "$token_response" | jq -r '.data.token // empty')

if [ -z "$TOKEN_VALUE" ] || [ "$TOKEN_VALUE" = "null" ]; then
    echo -e "${RED}❌ 无法提取令牌值，令牌创建可能失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 令牌创建成功${NC}"
echo "令牌值: $TOKEN_VALUE"
echo ""

echo -e "${BLUE}=== 第三步：测试HTTP代理认证 ===${NC}"

# 3. 测试HTTP代理 - 管理员密钥认证
test_request "HTTP代理 - 管理员密钥认证" \
    "GET" \
    "/proxy?target=https://httpbin.org/get" \
    "-H 'X-Log-Secret: $ADMIN_SECRET'" \
    "200"

# 4. 测试HTTP代理 - 令牌认证
test_request "HTTP代理 - 令牌认证" \
    "GET" \
    "/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "-H 'X-Proxy-Token: $TOKEN_VALUE'" \
    "200"

# 5. 测试HTTP代理 - Bearer令牌认证
test_request "HTTP代理 - Bearer令牌认证" \
    "GET" \
    "/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "-H 'Authorization: Bearer $TOKEN_VALUE'" \
    "200"

# 6. 测试HTTP代理 - 查询参数令牌认证
test_request "HTTP代理 - 查询参数令牌认证" \
    "GET" \
    "/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID&token=$TOKEN_VALUE" \
    "" \
    "200"

# 7. 测试HTTP代理 - 无效令牌
test_request "HTTP代理 - 无效令牌" \
    "GET" \
    "/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "-H 'X-Proxy-Token: invalid-token'" \
    "401"

# 8. 测试HTTP代理 - 无认证信息
test_request "HTTP代理 - 无认证信息" \
    "GET" \
    "/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" \
    "" \
    "401"

echo -e "${BLUE}=== 第四步：测试子域名代理认证 ===${NC}"

# 9. 测试子域名代理 - 令牌认证
test_request "子域名代理 - 令牌认证" \
    "GET" \
    "/get" \
    "-H 'Host: integration-test.localhost:10805' -H 'X-Proxy-Token: $TOKEN_VALUE'" \
    "200"

# 10. 测试子域名代理 - 无效令牌
test_request "子域名代理 - 无效令牌" \
    "GET" \
    "/get" \
    "-H 'Host: integration-test.localhost:10805' -H 'X-Proxy-Token: invalid-token'" \
    "401"

# 11. 测试子域名代理 - 无认证信息
test_request "子域名代理 - 无认证信息" \
    "GET" \
    "/get" \
    "-H 'Host: integration-test.localhost:10805'" \
    "401"

echo -e "${BLUE}=== 第五步：测试令牌使用统计 ===${NC}"

# 12. 获取初始统计
echo "获取初始令牌统计..."
initial_stats=$(curl -s -H "X-Log-Secret: $ADMIN_SECRET" \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

initial_requests=$(echo "$initial_stats" | jq -r '.data.stats.total_requests // 0')
echo "初始请求数: $initial_requests"

# 13. 执行多个代理请求
echo "执行多个代理请求以更新统计..."
for i in {1..3}; do
    echo "执行第 $i 个请求..."
    curl -s -H "X-Proxy-Token: $TOKEN_VALUE" \
        "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID" > /dev/null
    sleep 1
done

# 14. 获取更新后的统计
echo "获取更新后的令牌统计..."
updated_stats=$(curl -s -H "X-Log-Secret: $ADMIN_SECRET" \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

updated_requests=$(echo "$updated_stats" | jq -r '.data.stats.total_requests // 0')
echo "更新后请求数: $updated_requests"

if [ "$updated_requests" -gt "$initial_requests" ]; then
    echo -e "${GREEN}✅ 令牌使用统计更新正常${NC}"
else
    echo -e "${RED}❌ 令牌使用统计未更新${NC}"
fi

echo ""

echo -e "${BLUE}=== 第六步：测试CORS支持 ===${NC}"

# 15. 测试CORS预检请求
test_request "CORS预检请求" \
    "OPTIONS" \
    "/proxy" \
    "-H 'Origin: https://example.com' -H 'Access-Control-Request-Method: GET' -H 'Access-Control-Request-Headers: X-Proxy-Token'" \
    "200"

echo -e "${BLUE}=== 第七步：清理测试数据 ===${NC}"

# 16. 删除测试令牌
TOKEN_ID=$(echo "$token_response" | jq -r '.data.id // empty')
if [ -n "$TOKEN_ID" ] && [ "$TOKEN_ID" != "null" ]; then
    echo "删除测试令牌..."
    curl -s -X DELETE \
        -H "X-Log-Secret: $ADMIN_SECRET" \
        "$BASE_URL/config/proxy/$CONFIG_ID/tokens/$TOKEN_ID" > /dev/null
    echo -e "${GREEN}✅ 令牌已删除${NC}"
fi

# 17. 删除测试配置
echo "删除测试配置..."
curl -s -X DELETE \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    "$BASE_URL/config/proxy/$CONFIG_ID" > /dev/null
echo -e "${GREEN}✅ 配置已删除${NC}"

echo ""
echo -e "${GREEN}🎉 代理请求认证集成测试完成！${NC}"
echo ""
echo "测试总结:"
echo "- ✅ HTTP代理认证（管理员密钥、令牌、Bearer、查询参数）"
echo "- ✅ 子域名代理认证"
echo "- ✅ 令牌使用统计更新"
echo "- ✅ CORS支持"
echo "- ✅ 错误处理和状态码"
echo ""
echo "支持的认证方式:"
echo "- 管理员密钥: X-Log-Secret 请求头"
echo "- 令牌认证: X-Proxy-Token 请求头"
echo "- Bearer认证: Authorization: Bearer <token> 请求头"
echo "- 查询参数: ?token=<token>"
echo ""
echo "代理端点:"
echo "- HTTP代理: GET /proxy?target=<url>&config_id=<id>"
echo "- 子域名代理: GET http://<subdomain>.domain/<path>"
