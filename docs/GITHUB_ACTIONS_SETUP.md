# GitHub Actions Docker 构建设置指南

本文档说明如何设置GitHub Actions来自动构建和推送Docker镜像到Docker Hub。

## 🔧 前置要求

1. **GitHub仓库** - 代码已推送到GitHub
2. **Docker Hub账户** - 已创建仓库 `richpeople/privacy-gateway`
3. **Docker Hub访问令牌** - 用于GitHub Actions认证

## 📋 设置步骤

### 1. 创建Docker Hub访问令牌

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 点击右上角头像 → **Account Settings**
3. 选择 **Security** 标签
4. 点击 **New Access Token**
5. 填写令牌名称（如：`github-actions`）
6. 选择权限：**Read, Write, Delete**
7. 点击 **Generate** 并**复制令牌**（只显示一次）

### 2. 在GitHub仓库中设置Secrets

1. 进入GitHub仓库页面
2. 点击 **Settings** 标签
3. 左侧菜单选择 **Secrets and variables** → **Actions**
4. 点击 **New repository secret** 添加以下secrets：

| Secret名称 | 值 | 说明 |
|-----------|---|------|
| `DOCKER_USERNAME` | 你的Docker Hub用户名 | 用于登录Docker Hub |
| `DOCKER_PASSWORD` | 刚创建的访问令牌 | Docker Hub访问令牌（不是密码） |

### 3. 工作流触发条件

GitHub Actions将在以下情况自动触发：

- ✅ **推送到main分支** - 构建并推送 `latest` 标签
- ✅ **创建版本标签** - 构建并推送版本标签（如 `v1.0.0`）
- ✅ **Pull Request** - 仅构建，不推送（用于测试）
- ✅ **手动触发** - 在Actions页面手动运行

### 4. 标签策略

| 触发条件 | 生成的标签 | 示例 |
|---------|-----------|------|
| 推送到main分支 | `latest`, `main` | `richpeople/privacy-gateway:latest` |
| 创建标签 `v1.2.3` | `1.2.3`, `1.2`, `1`, `latest` | `richpeople/privacy-gateway:1.2.3` |
| Pull Request | `pr-123` | 仅构建，不推送 |

### 5. 多平台支持

构建的镜像支持以下平台：
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM64/Apple Silicon)

## 🚀 使用方法

### 发布新版本

1. **创建版本标签**：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **推送到main分支**：
   ```bash
   git push origin main
   ```

### 手动触发构建

1. 进入GitHub仓库的 **Actions** 页面
2. 选择 **Build and Push Docker Image** 工作流
3. 点击 **Run workflow**
4. 选择分支并点击 **Run workflow**

## 📊 构建状态

可以在以下位置查看构建状态：

- **GitHub Actions页面**: `https://github.com/你的用户名/PrivacyGateway/actions`
- **Docker Hub仓库**: `https://hub.docker.com/r/richpeople/privacy-gateway`

## 🔍 故障排除

### 常见问题

1. **认证失败**
   - 检查 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` secrets是否正确设置
   - 确认使用的是访问令牌，不是密码

2. **构建失败**
   - 检查Dockerfile语法
   - 查看Actions日志中的详细错误信息

3. **推送失败**
   - 确认Docker Hub仓库 `richpeople/privacy-gateway` 存在
   - 检查访问令牌权限

### 查看构建日志

1. 进入GitHub仓库的 **Actions** 页面
2. 点击失败的工作流运行
3. 展开相应的步骤查看详细日志

## 🎯 最佳实践

1. **版本管理**
   - 使用语义化版本号（如 `v1.2.3`）
   - 为重要版本创建GitHub Release

2. **安全性**
   - 定期轮换Docker Hub访问令牌
   - 不要在代码中硬编码敏感信息

3. **监控**
   - 关注构建失败通知
   - 定期检查Docker Hub镜像大小和安全扫描结果

## 📝 工作流文件说明

工作流文件位置：`.github/workflows/docker-build.yml`

主要功能：
- 多平台构建（AMD64 + ARM64）
- 智能标签管理
- GitHub Actions缓存优化
- 安全的Docker Hub推送
- 完整的镜像元数据
