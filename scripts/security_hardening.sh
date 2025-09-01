#!/bin/bash

# Privacy Gateway 安全加固脚本
# 使用方法: ./scripts/security_hardening.sh [environment]

set -e

ENVIRONMENT=${1:-"development"}

echo "🔒 Privacy Gateway 安全加固开始..."
echo "环境: $ENVIRONMENT"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 安全检查函数
security_check() {
    local check_name=$1
    local check_command=$2
    local expected_result=$3
    
    echo -e "${YELLOW}检查: $check_name${NC}"
    
    if eval "$check_command"; then
        if [ "$expected_result" = "pass" ]; then
            echo -e "${GREEN}✅ 通过${NC}"
            return 0
        else
            echo -e "${RED}❌ 失败 (预期失败但通过了)${NC}"
            return 1
        fi
    else
        if [ "$expected_result" = "fail" ]; then
            echo -e "${GREEN}✅ 通过 (正确失败)${NC}"
            return 0
        else
            echo -e "${RED}❌ 失败${NC}"
            return 1
        fi
    fi
}

echo -e "${BLUE}=== 第一步：环境安全检查 ===${NC}"

# 1. 检查Go版本
echo "检查Go版本..."
GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+\.[0-9]\+')
echo "当前Go版本: $GO_VERSION"

# 检查是否为安全版本
if [[ "$GO_VERSION" < "go1.19.0" ]]; then
    echo -e "${RED}⚠️  警告: Go版本过低，建议升级到1.19+${NC}"
else
    echo -e "${GREEN}✅ Go版本安全${NC}"
fi

# 2. 检查依赖安全性
echo ""
echo "检查依赖安全性..."
if command -v govulncheck >/dev/null 2>&1; then
    echo "运行漏洞扫描..."
    govulncheck ./... || echo -e "${YELLOW}⚠️  发现潜在漏洞，请检查输出${NC}"
else
    echo -e "${YELLOW}⚠️  govulncheck未安装，跳过漏洞扫描${NC}"
    echo "安装命令: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

echo ""
echo -e "${BLUE}=== 第二步：代码安全检查 ===${NC}"

# 3. 检查敏感信息泄露
echo "检查敏感信息泄露..."
SENSITIVE_PATTERNS=(
    "password.*=.*['\"][^'\"]*['\"]"
    "secret.*=.*['\"][^'\"]*['\"]"
    "token.*=.*['\"][^'\"]*['\"]"
    "api_key.*=.*['\"][^'\"]*['\"]"
    "private_key"
)

for pattern in "${SENSITIVE_PATTERNS[@]}"; do
    if grep -r -i "$pattern" --include="*.go" --include="*.json" --include="*.yaml" --include="*.yml" . 2>/dev/null; then
        echo -e "${RED}⚠️  发现可能的敏感信息泄露: $pattern${NC}"
    fi
done

# 4. 检查不安全的函数使用
echo ""
echo "检查不安全的函数使用..."
UNSAFE_FUNCTIONS=(
    "fmt\.Sprintf.*%s.*%s"
    "exec\.Command"
    "os\.Exec"
    "unsafe\."
    "reflect\.UnsafeAddr"
)

for func in "${UNSAFE_FUNCTIONS[@]}"; do
    if grep -r "$func" --include="*.go" . 2>/dev/null; then
        echo -e "${YELLOW}⚠️  发现可能不安全的函数使用: $func${NC}"
    fi
done

echo ""
echo -e "${BLUE}=== 第三步：配置安全检查 ===${NC}"

# 5. 检查默认密钥
echo "检查默认密钥配置..."
DEFAULT_SECRETS=(
    "admin"
    "password"
    "123456"
    "secret"
    "12345678"
    "test"
)

for secret in "${DEFAULT_SECRETS[@]}"; do
    if grep -r -i "$secret" --include="*.go" --include="*.json" --include="*.env" . 2>/dev/null | grep -v test | grep -v example; then
        echo -e "${RED}⚠️  发现可能的默认密钥: $secret${NC}"
    fi
done

# 6. 检查文件权限
echo ""
echo "检查关键文件权限..."
CRITICAL_FILES=(
    "main.go"
    "internal/config/config.go"
    "internal/handler/token_api.go"
)

for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$file" ]; then
        perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [ "$perms" -gt 644 ]; then
            echo -e "${YELLOW}⚠️  文件权限过宽: $file ($perms)${NC}"
        else
            echo -e "${GREEN}✅ 文件权限安全: $file ($perms)${NC}"
        fi
    fi
done

echo ""
echo -e "${BLUE}=== 第四步：运行时安全检查 ===${NC}"

# 7. 检查端口配置
echo "检查端口配置..."
if grep -r "0\.0\.0\.0" --include="*.go" . 2>/dev/null; then
    echo -e "${YELLOW}⚠️  发现绑定到所有接口的配置，生产环境建议限制${NC}"
fi

# 8. 检查日志配置
echo ""
echo "检查日志配置..."
if ! grep -r "log\." --include="*.go" . 2>/dev/null | grep -i "error\|warn" >/dev/null; then
    echo -e "${YELLOW}⚠️  可能缺少错误日志记录${NC}"
else
    echo -e "${GREEN}✅ 发现错误日志记录${NC}"
fi

echo ""
echo -e "${BLUE}=== 第五步：安全测试 ===${NC}"

# 9. 运行安全测试
echo "运行安全测试..."
if [ -f "test/security/security_test.go" ]; then
    echo "执行安全测试套件..."
    go test ./test/security/ -v -timeout=60s || echo -e "${RED}⚠️  安全测试发现问题${NC}"
else
    echo -e "${YELLOW}⚠️  未找到安全测试文件${NC}"
fi

# 10. 运行基准测试
echo ""
echo "运行性能基准测试..."
if go test -bench=. -benchtime=1s ./... 2>/dev/null; then
    echo -e "${GREEN}✅ 基准测试完成${NC}"
else
    echo -e "${YELLOW}⚠️  基准测试未找到或失败${NC}"
fi

echo ""
echo -e "${BLUE}=== 第六步：生成安全报告 ===${NC}"

# 11. 生成安全报告
REPORT_FILE="security_report_$(date +%Y%m%d_%H%M%S).txt"
echo "生成安全报告: $REPORT_FILE"

cat > "$REPORT_FILE" << EOF
Privacy Gateway 安全检查报告
生成时间: $(date)
环境: $ENVIRONMENT
Go版本: $GO_VERSION

=== 检查项目 ===
1. Go版本检查: $([ "$GO_VERSION" \> "go1.19.0" ] && echo "通过" || echo "需要升级")
2. 依赖安全检查: $(command -v govulncheck >/dev/null && echo "已执行" || echo "未安装工具")
3. 敏感信息检查: 已执行
4. 不安全函数检查: 已执行
5. 默认密钥检查: 已执行
6. 文件权限检查: 已执行
7. 端口配置检查: 已执行
8. 日志配置检查: 已执行
9. 安全测试: $([ -f "test/security/security_test.go" ] && echo "已执行" || echo "未找到")
10. 基准测试: 已执行

=== 建议措施 ===
EOF

# 根据环境添加特定建议
if [ "$ENVIRONMENT" = "production" ]; then
    cat >> "$REPORT_FILE" << EOF

生产环境安全建议:
1. 使用强密钥和证书
2. 启用HTTPS和安全头部
3. 配置防火墙和访问控制
4. 启用日志监控和告警
5. 定期进行安全扫描
6. 实施备份和恢复策略
EOF
else
    cat >> "$REPORT_FILE" << EOF

开发环境安全建议:
1. 不要使用生产数据
2. 定期更新依赖
3. 使用代码扫描工具
4. 进行安全测试
5. 遵循安全编码规范
EOF
fi

echo -e "${GREEN}✅ 安全报告已生成: $REPORT_FILE${NC}"

echo ""
echo -e "${BLUE}=== 第七步：安全加固建议 ===${NC}"

echo "根据检查结果，建议采取以下安全加固措施:"
echo ""
echo "🔒 立即措施:"
echo "  1. 修复发现的高风险漏洞"
echo "  2. 更新不安全的依赖"
echo "  3. 修改默认密钥和配置"
echo ""
echo "🛡️  短期措施:"
echo "  1. 实施速率限制"
echo "  2. 添加安全监控"
echo "  3. 启用安全头部"
echo ""
echo "🔐 长期措施:"
echo "  1. 定期安全审计"
echo "  2. 实施安全培训"
echo "  3. 建立安全响应流程"

echo ""
echo -e "${GREEN}🎉 安全加固检查完成！${NC}"
echo ""
echo "详细报告: $REPORT_FILE"
echo "下一步: 根据报告中的建议进行安全加固"
