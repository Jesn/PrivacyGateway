#!/bin/bash

# Privacy Gateway 版本发布脚本
# 用法: ./scripts/release.sh <version>
# 示例: ./scripts/release.sh 1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ $# -eq 0 ]; then
    print_error "请提供版本号"
    echo "用法: $0 <version>"
    echo "示例: $0 1.0.0"
    exit 1
fi

VERSION=$1

# 验证版本号格式 (语义化版本)
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "版本号格式无效。请使用语义化版本格式 (例如: 1.0.0)"
    exit 1
fi

TAG="v$VERSION"

print_info "准备发布版本: $TAG"

# 检查是否在git仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "当前目录不是git仓库"
    exit 1
fi

# 检查工作目录是否干净
if ! git diff-index --quiet HEAD --; then
    print_error "工作目录有未提交的更改，请先提交或暂存"
    git status --short
    exit 1
fi

# 检查是否在main分支
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    print_warning "当前不在main分支 (当前: $CURRENT_BRANCH)"
    read -p "是否继续? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "发布已取消"
        exit 0
    fi
fi

# 检查标签是否已存在
if git tag -l | grep -q "^$TAG$"; then
    print_error "标签 $TAG 已存在"
    exit 1
fi

# 拉取最新代码
print_info "拉取最新代码..."
git pull origin $CURRENT_BRANCH

# 运行测试（如果存在）
if [ -f "go.mod" ]; then
    print_info "运行测试..."
    go test ./... || {
        print_error "测试失败，发布已取消"
        exit 1
    }
fi

# 构建本地测试
print_info "构建本地测试..."
go build -o privacy-gateway . || {
    print_error "构建失败，发布已取消"
    exit 1
}
rm -f privacy-gateway

# 创建标签
print_info "创建标签 $TAG..."
git tag -a "$TAG" -m "Release $TAG"

# 推送标签
print_info "推送标签到远程仓库..."
git push origin "$TAG"

print_success "版本 $TAG 发布成功!"
print_info "GitHub Actions将自动构建并推送Docker镜像"
print_info "查看构建状态: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
print_info "Docker镜像将推送到: richpeople/privacy-gateway:$VERSION"

echo
print_info "发布完成后，你可以使用以下命令拉取镜像:"
echo "  docker pull richpeople/privacy-gateway:$VERSION"
echo "  docker pull richpeople/privacy-gateway:latest"
