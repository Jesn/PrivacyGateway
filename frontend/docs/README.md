# Privacy Gateway Frontend

一个现代化的前端应用，基于模块化架构和组件系统构建，为Privacy Gateway提供用户界面。

## 🚀 项目概览

Privacy Gateway Frontend是一个使用现代Web技术栈构建的单页应用程序，提供了直观的用户界面来管理代理配置、查看访问日志和监控系统状态。

### ✨ 主要特性

- 🎨 **现代化UI设计** - 使用Tailwind CSS构建的响应式界面
- 🧩 **模块化架构** - 基于ES6模块的清晰代码组织
- ⚡ **高性能** - 代码分割、懒加载和性能优化
- 🔧 **组件系统** - 可复用的UI组件库
- 📊 **状态管理** - 集中式状态管理和数据流
- 🧪 **完整测试** - 单元测试和端到端测试覆盖
- 🛠️ **开发工具** - 现代化的开发和构建工具链

## 📁 项目结构

```
frontend/
├── assets/                 # 静态资源
│   ├── css/                # 样式文件
│   │   ├── main.css        # 主样式文件
│   │   └── components.css  # 组件样式
│   └── js/                 # JavaScript模块
│       ├── api/            # API接口层
│       ├── components/     # UI组件
│       ├── modules/        # 功能模块
│       └── utils/          # 工具函数
├── docs/                   # 项目文档
├── tests/                  # 测试文件
│   ├── e2e/               # 端到端测试
│   ├── integration/       # 集成测试
│   └── utils/             # 单元测试
├── index.html             # 主页面
├── index-optimized.html   # 优化版页面
├── package.json           # 项目配置
├── vite.config.js         # Vite构建配置
├── vitest.config.js       # 测试配置
└── playwright.config.js   # E2E测试配置
```

## 🛠️ 技术栈

### 核心技术
- **JavaScript ES6+** - 现代JavaScript语法和特性
- **HTML5** - 语义化标记和现代Web标准
- **CSS3** - 现代样式和动画

### 样式框架
- **Tailwind CSS** - 实用优先的CSS框架
- **自定义组件样式** - 针对特定组件的样式增强

### 构建工具
- **Vite** - 快速的构建工具和开发服务器
- **ESLint** - 代码质量检查
- **Prettier** - 代码格式化

### 测试框架
- **Vitest** - 单元测试和集成测试
- **Playwright** - 端到端测试
- **Happy DOM** - 轻量级DOM环境

## 🚀 快速开始

### 环境要求

- Node.js 16+ 
- npm 7+ 或 yarn 1.22+

### 安装依赖

```bash
cd frontend
npm install
```

### 开发模式

```bash
# 启动开发服务器
npm run dev

# 启动后端服务器（在另一个终端）
cd ..
ADMIN_SECRET=12345678 go run main.go
```

访问 http://localhost:3000 查看应用。

### 构建生产版本

```bash
# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

## 🧪 测试

### 运行所有测试

```bash
# 单元测试
npm test

# 单元测试（监听模式）
npm run test:watch

# 测试覆盖率
npm run test:coverage

# 端到端测试
npm run test:e2e
```

### 测试结构

- **单元测试** - 测试独立的函数和类
- **集成测试** - 测试模块间的交互
- **端到端测试** - 测试完整的用户流程

## 📚 架构文档

### 模块化架构

项目采用ES6模块系统，按功能划分为不同的模块：

- **API层** (`/api`) - 处理与后端的通信
- **组件层** (`/components`) - 可复用的UI组件
- **工具层** (`/utils`) - 通用工具函数
- **模块层** (`/modules`) - 业务逻辑模块

### 组件系统

基于类的组件系统，提供：
- 生命周期管理
- 状态管理
- 事件处理
- DOM操作封装

### 状态管理

集中式状态管理系统，支持：
- 响应式状态更新
- 状态持久化
- 中间件支持
- 计算属性

## 🎨 UI组件

### 基础组件

- **Modal** - 模态对话框
- **DataTable** - 数据表格
- **Form** - 表单组件
- **Button** - 按钮组件

### 使用示例

```javascript
import { Modal } from '@components/Modal.js';

// 显示确认对话框
Modal.confirm({
  title: '确认删除',
  content: '确定要删除这个配置吗？',
  onConfirm: () => {
    // 删除逻辑
  }
});
```

## 🔧 开发指南

### 代码规范

项目使用ESLint和Prettier确保代码质量：

```bash
# 检查代码规范
npm run lint

# 自动修复代码格式
npm run format
```

### 添加新组件

1. 在 `assets/js/components/` 创建组件文件
2. 继承基础 `Component` 类
3. 实现必要的生命周期方法
4. 添加对应的测试文件

### 添加新工具函数

1. 在 `assets/js/utils/` 创建工具文件
2. 导出纯函数
3. 添加JSDoc注释
4. 编写单元测试

## 📈 性能优化

### 已实现的优化

- **代码分割** - 按需加载模块
- **懒加载** - 图片和组件懒加载
- **缓存策略** - 浏览器缓存优化
- **压缩优化** - 代码和资源压缩

### 性能监控

使用内置的性能监控工具：

```javascript
import { performanceMonitor } from '@utils/Performance.js';

// 获取性能报告
const report = performanceMonitor.generateReport();
console.log(report);
```

## 🐛 调试

### 开发工具

- **浏览器开发者工具** - 调试JavaScript和CSS
- **Vue DevTools** - 组件状态调试（如果使用）
- **网络面板** - API请求调试

### 调试模式

在URL中添加 `?debug=true` 启用调试模式：

```
http://localhost:3000?debug=true
```

## 📝 更新日志

### v2.0.0 (当前版本)

- ✨ 全新的模块化架构
- 🎨 Tailwind CSS集成
- ⚡ Vite构建工具
- 🧪 完整的测试覆盖
- 📊 性能监控系统
- 🔧 现代化开发工具链

### v1.0.0

- 🎯 基础功能实现
- 📱 响应式设计
- 🔐 认证系统
- 📊 数据可视化

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

如果您遇到问题或有疑问：

1. 查看 [文档](docs/)
2. 搜索 [Issues](../../issues)
3. 创建新的 [Issue](../../issues/new)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和用户！
