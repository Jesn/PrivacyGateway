# 更新日志

本文档记录了Privacy Gateway的所有重要更改。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 新增
- 待发布的新功能

### 修改
- 待发布的功能改进

### 修复
- 待发布的bug修复

## [v1.0.0] - 2024-01-XX

### 新增
- 🔒 **隐私保护代理** - 自动过滤敏感请求头信息
- 🌐 **多协议支持** - HTTP/HTTPS/WebSocket代理功能
- 🚀 **代理支持** - 支持设置上游代理（HTTP/HTTPS/SOCKS5）
- 📊 **访问日志管理** - 实时日志查看和管理界面
- 🔍 **智能搜索** - 多维度日志筛选和搜索功能
- 📋 **curl命令生成** - 自动生成等效curl命令
- 🎯 **请求类型识别** - 自动识别HTTP、HTTPS、WebSocket、SSE
- 🔧 **灵活配置** - 支持环境变量和配置文件
- 🐳 **Docker支持** - 完整的Docker部署方案
- 🔄 **CI/CD流程** - GitHub Actions自动构建和发布
- 📦 **多平台二进制** - 支持6个平台的预编译二进制文件

### 技术特性
- 轻量级设计，Docker镜像仅10MB
- 多平台支持（Linux/macOS/Windows，AMD64/ARM64）
- 内存高效的日志存储系统
- 实时WebSocket代理
- 安全的管理界面认证
- 完整的请求/响应记录
- 智能的日志记录模式（可配置是否记录200状态码详情）

### 配置选项
- `GATEWAY_PORT` - 服务端口配置
- `ADMIN_SECRET` - 管理界面密钥
- `LOG_RECORD_200` - 成功请求详情记录控制
- `SENSITIVE_HEADERS` - 敏感头信息过滤配置
- `DEFAULT_PROXY` - 默认上游代理配置
- `PROXY_WHITELIST` - 代理白名单配置
- 更多配置选项详见 `.env.example`

### 部署方式
- Docker容器部署
- Docker Compose部署
- 预编译二进制文件部署
- 源码编译部署

---

## 版本说明

- **主版本号**：不兼容的API修改
- **次版本号**：向下兼容的功能性新增
- **修订号**：向下兼容的问题修正

## 链接

- [项目主页](https://github.com/Jesn/PrivacyGateway)
- [Docker Hub](https://hub.docker.com/r/richpeople/privacy-gateway)
- [问题反馈](https://github.com/Jesn/PrivacyGateway/issues)
- [发布页面](https://github.com/Jesn/PrivacyGateway/releases)
