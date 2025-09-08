# 部署指南

Privacy Gateway Frontend 的部署和生产环境配置指南。

## 📋 目录

- [环境要求](#环境要求)
- [构建配置](#构建配置)
- [部署方式](#部署方式)
- [性能优化](#性能优化)
- [监控和维护](#监控和维护)

## 🔧 环境要求

### 开发环境

- **Node.js**: 16.0+ (推荐 18.0+)
- **npm**: 7.0+ 或 yarn 1.22+
- **Go**: 1.19+ (后端服务)
- **Git**: 2.0+

### 生产环境

- **服务器**: Linux/macOS/Windows
- **内存**: 最少 512MB，推荐 1GB+
- **存储**: 最少 100MB，推荐 500MB+
- **网络**: 支持HTTP/HTTPS

## 🏗️ 构建配置

### 开发构建

```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 代码检查
npm run lint

# 运行测试
npm test
```

### 生产构建

```bash
# 构建生产版本
npm run build

# 预览构建结果
npm run preview

# 生成测试覆盖率报告
npm run test:coverage
```

### 构建优化

在 `vite.config.js` 中配置生产优化：

```javascript
export default defineConfig({
  build: {
    // 输出目录
    outDir: 'dist',
    
    // 启用源码映射（可选）
    sourcemap: false,
    
    // 压缩配置
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    
    // 代码分割
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['axios', 'dayjs'],
          utils: ['./assets/js/utils/'],
          components: ['./assets/js/components/']
        }
      }
    }
  }
});
```

## 🚀 部署方式

### 1. 静态文件部署

最简单的部署方式，适用于大多数场景。

```bash
# 构建项目
npm run build

# 部署到Web服务器
cp -r dist/* /var/www/html/
```

**Nginx 配置示例：**

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /var/www/html;
    index index.html;
    
    # 启用gzip压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
    
    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
    # SPA路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # API代理
    location /api/ {
        proxy_pass http://localhost:10805;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. Docker 部署

使用Docker容器化部署。

**Dockerfile:**

```dockerfile
# 构建阶段
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# 生产阶段
FROM nginx:alpine

# 复制构建结果
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制Nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

**docker-compose.yml:**

```yaml
version: '3.8'

services:
  frontend:
    build: .
    ports:
      - "80:80"
    environment:
      - NODE_ENV=production
    restart: unless-stopped
    
  backend:
    build: ../
    ports:
      - "10805:10805"
    environment:
      - ADMIN_SECRET=${ADMIN_SECRET}
    restart: unless-stopped
    
networks:
  default:
    name: privacy-gateway
```

### 3. CDN 部署

使用CDN加速静态资源。

```bash
# 构建项目
npm run build

# 上传到CDN
aws s3 sync dist/ s3://your-bucket-name --delete

# 配置CloudFront分发
```

### 4. 集成部署

与Go后端一起部署。

```bash
# 构建前端
cd frontend
npm run build

# 将构建结果复制到Go项目
cp -r dist/* ../static/

# 构建Go应用
cd ..
go build -o privacy-gateway main.go

# 运行
./privacy-gateway
```

## ⚡ 性能优化

### 1. 构建优化

```javascript
// vite.config.js
export default defineConfig({
  build: {
    // 启用代码分割
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            return 'vendor';
          }
          if (id.includes('utils')) {
            return 'utils';
          }
          if (id.includes('components')) {
            return 'components';
          }
        }
      }
    },
    
    // 压缩配置
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        pure_funcs: ['console.log']
      }
    }
  }
});
```

### 2. 资源优化

```bash
# 图片压缩
npm install --save-dev imagemin imagemin-webp

# CSS优化
npm install --save-dev cssnano

# JavaScript优化
npm install --save-dev terser
```

### 3. 缓存策略

```nginx
# 静态资源长期缓存
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
    add_header Vary "Accept-Encoding";
}

# HTML文件不缓存
location ~* \.html$ {
    expires -1;
    add_header Cache-Control "no-cache, no-store, must-revalidate";
}
```

### 4. 预加载优化

```html
<!-- 关键资源预加载 -->
<link rel="preload" href="/assets/css/main.css" as="style">
<link rel="preload" href="/assets/js/main.js" as="script">

<!-- DNS预解析 -->
<link rel="dns-prefetch" href="//cdn.example.com">

<!-- 预连接 -->
<link rel="preconnect" href="https://api.example.com">
```

## 📊 监控和维护

### 1. 性能监控

```javascript
// 性能监控代码
import { performanceMonitor } from '@utils/Performance.js';

// 监控页面加载性能
const metrics = performanceMonitor.getPageMetrics();
console.log('页面性能指标:', metrics);

// 监控资源加载
const resourceStats = performanceMonitor.getResourceStats();
console.log('资源统计:', resourceStats);

// 发送到监控服务
fetch('/api/metrics', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ metrics, resourceStats })
});
```

### 2. 错误监控

```javascript
// 全局错误处理
window.addEventListener('error', (event) => {
  console.error('JavaScript错误:', event.error);
  
  // 发送错误报告
  fetch('/api/errors', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message: event.error.message,
      stack: event.error.stack,
      url: window.location.href,
      timestamp: new Date().toISOString()
    })
  });
});

// Promise错误处理
window.addEventListener('unhandledrejection', (event) => {
  console.error('未处理的Promise拒绝:', event.reason);
});
```

### 3. 日志收集

```javascript
// 日志收集器
class Logger {
  static log(level, message, data = {}) {
    const logEntry = {
      level,
      message,
      data,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent
    };
    
    // 本地存储
    console[level](message, data);
    
    // 发送到服务器
    if (level === 'error' || level === 'warn') {
      this.sendToServer(logEntry);
    }
  }
  
  static sendToServer(logEntry) {
    fetch('/api/logs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(logEntry)
    }).catch(err => {
      console.error('日志发送失败:', err);
    });
  }
}

// 使用示例
Logger.log('info', '用户登录成功', { userId: 123 });
Logger.log('error', 'API请求失败', { url: '/api/data', status: 500 });
```

### 4. 健康检查

```javascript
// 健康检查端点
app.get('/health', (req, res) => {
  const health = {
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    memory: process.memoryUsage(),
    version: process.env.npm_package_version
  };
  
  res.json(health);
});
```

## 🔒 安全配置

### 1. HTTPS配置

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # 安全头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
}
```

### 2. 环境变量

```bash
# .env.production
NODE_ENV=production
VITE_API_BASE_URL=https://api.your-domain.com
VITE_APP_VERSION=2.0.0
VITE_ENABLE_ANALYTICS=true
```

### 3. 访问控制

```nginx
# IP白名单（可选）
location /admin {
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    try_files $uri $uri/ /index.html;
}

# 速率限制
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://backend;
}
```

## 🔄 CI/CD 配置

### GitHub Actions 示例

```yaml
name: Deploy Frontend

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    
    - name: Install dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Run tests
      run: |
        cd frontend
        npm test
    
    - name: Build
      run: |
        cd frontend
        npm run build
    
    - name: Deploy to server
      run: |
        rsync -avz --delete frontend/dist/ user@server:/var/www/html/
```

## 📝 维护清单

### 日常维护

- [ ] 检查应用性能指标
- [ ] 查看错误日志
- [ ] 监控资源使用情况
- [ ] 验证备份完整性

### 定期维护

- [ ] 更新依赖包
- [ ] 安全漏洞扫描
- [ ] 性能优化评估
- [ ] 文档更新

### 应急响应

- [ ] 错误报警处理流程
- [ ] 服务降级方案
- [ ] 数据恢复计划
- [ ] 联系人信息更新

## 🆘 故障排除

### 常见问题

1. **构建失败**
   - 检查Node.js版本
   - 清理node_modules重新安装
   - 检查依赖版本冲突

2. **部署后白屏**
   - 检查控制台错误
   - 验证资源路径
   - 检查服务器配置

3. **API请求失败**
   - 检查代理配置
   - 验证CORS设置
   - 检查网络连接

4. **性能问题**
   - 分析bundle大小
   - 检查资源加载
   - 优化代码分割

### 调试工具

```bash
# 分析bundle大小
npm run build -- --analyze

# 检查依赖
npm audit

# 性能分析
npm run build -- --profile
```
