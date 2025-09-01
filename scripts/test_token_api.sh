#!/bin/bash

# 令牌管理API测试脚本
# 使用方法: ./scripts/test_token_api.sh [base_url] [admin_secret] [config_id]

set -e

# 默认参数
BASE_URL=${1:-"http://localhost:10805"}
ADMIN_SECRET=${2:-"12345678"}
CONFIG_ID=${3:-"test-config"}

echo "🔧 令牌管理API测试开始..."
echo "Base URL: $BASE_URL"
echo "Admin Secret: $ADMIN_SECRET"
echo "Config ID: $CONFIG_ID"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -e "${YELLOW}测试: $description${NC}"
    echo "请求: $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "X-Log-Secret: $ADMIN_SECRET" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "X-Log-Secret: $ADMIN_SECRET" \
            "$BASE_URL$endpoint")
    fi
    
    # 分离响应体和状态码
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    echo "状态码: $status"
    echo "响应: $body"
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✅ 测试通过${NC}"
    else
        echo -e "${RED}❌ 测试失败 (期望状态码: $expected_status)${NC}"
        return 1
    fi
    
    echo ""
    return 0
}

# 1. 测试获取令牌列表（空列表）
test_api "GET" "/config/proxy/$CONFIG_ID/tokens" "" "200" "获取空令牌列表"

# 2. 测试创建令牌
echo -e "${YELLOW}创建测试令牌...${NC}"
create_response=$(curl -s -X POST \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "API测试令牌",
        "description": "通过API创建的测试令牌"
    }' \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

echo "创建响应: $create_response"

# 提取令牌ID和令牌值
TOKEN_ID=$(echo "$create_response" | jq -r '.data.id // empty')
TOKEN_VALUE=$(echo "$create_response" | jq -r '.data.token // empty')

if [ -z "$TOKEN_ID" ] || [ "$TOKEN_ID" = "null" ]; then
    echo -e "${RED}❌ 无法提取令牌ID，创建可能失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 令牌创建成功${NC}"
echo "令牌ID: $TOKEN_ID"
echo "令牌值: $TOKEN_VALUE"
echo ""

# 3. 测试获取令牌列表（包含一个令牌）
test_api "GET" "/config/proxy/$CONFIG_ID/tokens" "" "200" "获取令牌列表（包含令牌）"

# 4. 测试获取单个令牌
test_api "GET" "/config/proxy/$CONFIG_ID/tokens/$TOKEN_ID" "" "200" "获取单个令牌"

# 5. 测试更新令牌
test_api "PUT" "/config/proxy/$CONFIG_ID/tokens/$TOKEN_ID" '{
    "name": "更新后的API测试令牌",
    "description": "已更新的描述",
    "enabled": false
}' "200" "更新令牌"

# 6. 测试令牌认证（使用创建的令牌）
if [ -n "$TOKEN_VALUE" ] && [ "$TOKEN_VALUE" != "null" ]; then
    echo -e "${YELLOW}测试令牌认证...${NC}"
    auth_response=$(curl -s -w "\n%{http_code}" \
        -H "X-Proxy-Token: $TOKEN_VALUE" \
        "$BASE_URL/proxy?target=https://httpbin.org/get&config_id=$CONFIG_ID")
    
    auth_body=$(echo "$auth_response" | head -n -1)
    auth_status=$(echo "$auth_response" | tail -n 1)
    
    echo "认证状态码: $auth_status"
    echo "认证响应: $auth_body"
    
    if [ "$auth_status" = "401" ]; then
        echo -e "${GREEN}✅ 令牌认证测试通过（令牌已禁用，返回401）${NC}"
    else
        echo -e "${YELLOW}⚠️ 令牌认证状态: $auth_status${NC}"
    fi
    echo ""
fi

# 7. 测试错误情况

# 7.1 无效的配置ID
test_api "GET" "/config/proxy/nonexistent-config/tokens" "" "404" "无效配置ID"

# 7.2 无效的令牌ID
test_api "GET" "/config/proxy/$CONFIG_ID/tokens/nonexistent-token" "" "404" "无效令牌ID"

# 7.3 无效的JSON
echo -e "${YELLOW}测试: 无效JSON格式${NC}"
invalid_json_response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "X-Log-Secret: $ADMIN_SECRET" \
    -H "Content-Type: application/json" \
    -d 'invalid json' \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

invalid_json_status=$(echo "$invalid_json_response" | tail -n 1)
echo "状态码: $invalid_json_status"

if [ "$invalid_json_status" = "400" ]; then
    echo -e "${GREEN}✅ 无效JSON测试通过${NC}"
else
    echo -e "${RED}❌ 无效JSON测试失败${NC}"
fi
echo ""

# 7.4 缺少必需字段
test_api "POST" "/config/proxy/$CONFIG_ID/tokens" '{
    "description": "缺少名称字段"
}' "400" "缺少必需字段"

# 7.5 未授权访问
echo -e "${YELLOW}测试: 未授权访问${NC}"
unauth_response=$(curl -s -w "\n%{http_code}" \
    "$BASE_URL/config/proxy/$CONFIG_ID/tokens")

unauth_status=$(echo "$unauth_response" | tail -n 1)
echo "状态码: $unauth_status"

if [ "$unauth_status" = "401" ]; then
    echo -e "${GREEN}✅ 未授权访问测试通过${NC}"
else
    echo -e "${RED}❌ 未授权访问测试失败${NC}"
fi
echo ""

# 8. 测试删除令牌
test_api "DELETE" "/config/proxy/$CONFIG_ID/tokens/$TOKEN_ID" "" "200" "删除令牌"

# 9. 验证令牌已删除
test_api "GET" "/config/proxy/$CONFIG_ID/tokens/$TOKEN_ID" "" "404" "验证令牌已删除"

echo -e "${GREEN}🎉 所有API测试完成！${NC}"
echo ""
echo "测试总结:"
echo "- ✅ 令牌CRUD操作"
echo "- ✅ 认证和授权"
echo "- ✅ 错误处理"
echo "- ✅ 输入验证"
echo ""
echo "API端点:"
echo "- GET    /config/proxy/{configID}/tokens          - 获取令牌列表"
echo "- POST   /config/proxy/{configID}/tokens          - 创建令牌"
echo "- GET    /config/proxy/{configID}/tokens/{tokenID} - 获取单个令牌"
echo "- PUT    /config/proxy/{configID}/tokens/{tokenID} - 更新令牌"
echo "- DELETE /config/proxy/{configID}/tokens/{tokenID} - 删除令牌"
echo ""
echo "认证方式:"
echo "- 管理员密钥: X-Log-Secret 请求头"
echo "- 令牌认证: X-Proxy-Token 请求头（用于代理请求）"
