# Privacy Gateway 部署和迁移指南

## 系统要求

### 最低要求
- **操作系统**: Linux (Ubuntu 18.04+, CentOS 7+), macOS 10.14+, Windows 10
- **CPU**: 1核心 (推荐2核心以上)
- **内存**: 512MB (推荐1GB以上)
- **存储**: 100MB (推荐1GB以上用于日志和数据)
- **网络**: 稳定的网络连接

### 推荐配置
- **操作系统**: Linux (Ubuntu 20.04 LTS)
- **CPU**: 4核心
- **内存**: 4GB
- **存储**: 10GB SSD
- **网络**: 千兆网络

### 软件依赖
- **Go**: 1.19+ (仅开发环境)
- **Git**: 2.0+ (仅开发环境)
- **curl**: 用于健康检查
- **systemd**: 用于服务管理 (Linux)

## 快速部署

### 1. 下载和安装

#### 方式一：预编译二进制文件
```bash
# 下载最新版本
wget https://github.com/your-org/privacy-gateway/releases/latest/download/privacy-gateway-linux-amd64.tar.gz

# 解压
tar -xzf privacy-gateway-linux-amd64.tar.gz

# 移动到系统目录
sudo mv privacy-gateway /usr/local/bin/
sudo chmod +x /usr/local/bin/privacy-gateway
```

#### 方式二：从源码编译
```bash
# 克隆仓库
git clone https://github.com/your-org/privacy-gateway.git
cd privacy-gateway

# 编译
go build -o privacy-gateway main.go

# 安装
sudo mv privacy-gateway /usr/local/bin/
```

### 2. 创建配置文件

```bash
# 创建配置目录
sudo mkdir -p /etc/privacy-gateway
sudo mkdir -p /var/lib/privacy-gateway
sudo mkdir -p /var/log/privacy-gateway

# 创建基本配置文件
sudo tee /etc/privacy-gateway/config.yaml << EOF
# Privacy Gateway 配置文件
server:
  port: "8080"
  host: "0.0.0.0"
  
security:
  admin_secret: "$(openssl rand -hex 32)"
  
storage:
  type: "file"
  path: "/var/lib/privacy-gateway/data"
  
logging:
  level: "info"
  path: "/var/log/privacy-gateway"
  
cors:
  enabled: true
  origins: ["*"]
  methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  headers: ["*"]
EOF
```

### 3. 创建系统服务

```bash
# 创建systemd服务文件
sudo tee /etc/systemd/system/privacy-gateway.service << EOF
[Unit]
Description=Privacy Gateway
After=network.target

[Service]
Type=simple
User=privacy-gateway
Group=privacy-gateway
WorkingDirectory=/var/lib/privacy-gateway
ExecStart=/usr/local/bin/privacy-gateway -config /etc/privacy-gateway/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/privacy-gateway /var/log/privacy-gateway

[Install]
WantedBy=multi-user.target
EOF
```

### 4. 创建用户和设置权限

```bash
# 创建专用用户
sudo useradd -r -s /bin/false privacy-gateway

# 设置目录权限
sudo chown -R privacy-gateway:privacy-gateway /var/lib/privacy-gateway
sudo chown -R privacy-gateway:privacy-gateway /var/log/privacy-gateway
sudo chmod 750 /var/lib/privacy-gateway
sudo chmod 750 /var/log/privacy-gateway
```

### 5. 启动服务

```bash
# 重新加载systemd配置
sudo systemctl daemon-reload

# 启用并启动服务
sudo systemctl enable privacy-gateway
sudo systemctl start privacy-gateway

# 检查服务状态
sudo systemctl status privacy-gateway
```

### 6. 验证部署

```bash
# 检查服务是否运行
curl -f http://localhost:8080/ || echo "Service not responding"

# 检查日志
sudo journalctl -u privacy-gateway -f
```

## 高级部署

### 1. 反向代理配置 (Nginx)

```nginx
# /etc/nginx/sites-available/privacy-gateway
server {
    listen 80;
    server_name your-domain.com *.your-domain.com;
    
    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com *.your-domain.com;
    
    # SSL配置
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;
    
    # 安全头部
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    
    # 代理配置
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # 静态文件缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        proxy_pass http://127.0.0.1:8080;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 2. Docker部署

#### Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o privacy-gateway main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/privacy-gateway .
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./privacy-gateway"]
```

#### docker-compose.yml
```yaml
version: '3.8'

services:
  privacy-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ADMIN_SECRET=${ADMIN_SECRET:-default-secret-change-me}
      - PORT=8080
      - LOG_LEVEL=info
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - privacy-gateway
    restart: unless-stopped
```

### 3. Kubernetes部署

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: privacy-gateway
  labels:
    app: privacy-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: privacy-gateway
  template:
    metadata:
      labels:
        app: privacy-gateway
    spec:
      containers:
      - name: privacy-gateway
        image: privacy-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: ADMIN_SECRET
          valueFrom:
            secretKeyRef:
              name: privacy-gateway-secret
              key: admin-secret
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: data-volume
          mountPath: /app/data
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: privacy-gateway-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: privacy-gateway-service
spec:
  selector:
    app: privacy-gateway
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
---
apiVersion: v1
kind: Secret
metadata:
  name: privacy-gateway-secret
type: Opaque
data:
  admin-secret: <base64-encoded-secret>
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: privacy-gateway-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

## 数据迁移

### 1. 从旧版本升级

#### 备份现有数据
```bash
# 停止服务
sudo systemctl stop privacy-gateway

# 备份数据
sudo cp -r /var/lib/privacy-gateway /var/lib/privacy-gateway.backup.$(date +%Y%m%d_%H%M%S)

# 备份配置
sudo cp -r /etc/privacy-gateway /etc/privacy-gateway.backup.$(date +%Y%m%d_%H%M%S)
```

#### 执行迁移
```bash
# 下载新版本
wget https://github.com/your-org/privacy-gateway/releases/latest/download/privacy-gateway-linux-amd64.tar.gz

# 替换二进制文件
sudo systemctl stop privacy-gateway
sudo cp /usr/local/bin/privacy-gateway /usr/local/bin/privacy-gateway.old
sudo tar -xzf privacy-gateway-linux-amd64.tar.gz
sudo mv privacy-gateway /usr/local/bin/
sudo chmod +x /usr/local/bin/privacy-gateway

# 运行数据迁移工具
sudo -u privacy-gateway /usr/local/bin/privacy-gateway -migrate -config /etc/privacy-gateway/config.yaml

# 启动服务
sudo systemctl start privacy-gateway
```

### 2. 配置迁移

#### 从环境变量迁移到配置文件
```bash
# 创建迁移脚本
cat > migrate_config.sh << 'EOF'
#!/bin/bash

# 读取环境变量并生成配置文件
cat > /etc/privacy-gateway/config.yaml << EOL
server:
  port: "${PORT:-8080}"
  host: "${HOST:-0.0.0.0}"
  
security:
  admin_secret: "${ADMIN_SECRET}"
  
storage:
  type: "file"
  path: "${DATA_PATH:-/var/lib/privacy-gateway/data}"
  
logging:
  level: "${LOG_LEVEL:-info}"
  path: "${LOG_PATH:-/var/log/privacy-gateway}"
EOL
EOF

chmod +x migrate_config.sh
sudo ./migrate_config.sh
```

### 3. 数据导入导出

#### 导出配置和令牌
```bash
# 导出所有配置
curl -H "X-Log-Secret: your-admin-secret" \
     "http://localhost:8080/config/proxy/export" > backup.json

# 验证导出文件
jq . backup.json > /dev/null && echo "Export file is valid JSON"
```

#### 导入配置和令牌
```bash
# 导入配置
curl -X POST \
     -H "X-Log-Secret: your-admin-secret" \
     -H "Content-Type: application/json" \
     -d @backup.json \
     "http://localhost:8080/config/proxy/import"
```

## 监控和维护

### 1. 健康检查

```bash
# 创建健康检查脚本
cat > /usr/local/bin/privacy-gateway-health.sh << 'EOF'
#!/bin/bash

# 检查服务状态
if ! systemctl is-active --quiet privacy-gateway; then
    echo "ERROR: Privacy Gateway service is not running"
    exit 1
fi

# 检查HTTP响应
if ! curl -f -s http://localhost:8080/ > /dev/null; then
    echo "ERROR: Privacy Gateway is not responding"
    exit 1
fi

# 检查磁盘空间
DISK_USAGE=$(df /var/lib/privacy-gateway | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    echo "WARNING: Disk usage is ${DISK_USAGE}%"
fi

echo "OK: Privacy Gateway is healthy"
EOF

chmod +x /usr/local/bin/privacy-gateway-health.sh
```

### 2. 日志轮转

```bash
# 创建logrotate配置
sudo tee /etc/logrotate.d/privacy-gateway << EOF
/var/log/privacy-gateway/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 privacy-gateway privacy-gateway
    postrotate
        systemctl reload privacy-gateway
    endscript
}
EOF
```

### 3. 自动备份

```bash
# 创建备份脚本
cat > /usr/local/bin/privacy-gateway-backup.sh << 'EOF'
#!/bin/bash

BACKUP_DIR="/var/backups/privacy-gateway"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据
tar -czf "$BACKUP_DIR/data_$DATE.tar.gz" -C /var/lib/privacy-gateway .

# 备份配置
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" -C /etc/privacy-gateway .

# 清理旧备份（保留30天）
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
EOF

chmod +x /usr/local/bin/privacy-gateway-backup.sh

# 添加到crontab
echo "0 2 * * * /usr/local/bin/privacy-gateway-backup.sh" | sudo crontab -
```

## 故障排除

### 1. 常见问题

#### 服务无法启动
```bash
# 检查服务状态
sudo systemctl status privacy-gateway

# 查看详细日志
sudo journalctl -u privacy-gateway -f

# 检查配置文件
sudo /usr/local/bin/privacy-gateway -config /etc/privacy-gateway/config.yaml -check
```

#### 端口冲突
```bash
# 检查端口占用
sudo netstat -tlnp | grep :8080

# 修改配置文件中的端口
sudo nano /etc/privacy-gateway/config.yaml
```

#### 权限问题
```bash
# 检查文件权限
ls -la /var/lib/privacy-gateway
ls -la /var/log/privacy-gateway

# 修复权限
sudo chown -R privacy-gateway:privacy-gateway /var/lib/privacy-gateway
sudo chown -R privacy-gateway:privacy-gateway /var/log/privacy-gateway
```

### 2. 性能调优

#### 系统级优化
```bash
# 增加文件描述符限制
echo "privacy-gateway soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "privacy-gateway hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# 优化网络参数
echo "net.core.somaxconn = 65536" | sudo tee -a /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

#### 应用级优化
```yaml
# 在config.yaml中添加性能配置
performance:
  max_connections: 1000
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  max_header_bytes: 1048576
```

## 安全加固

### 1. 系统安全
```bash
# 禁用不必要的服务
sudo systemctl disable apache2 nginx mysql

# 配置防火墙
sudo ufw enable
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 8080/tcp  # 只允许内部访问
```

### 2. 应用安全
```yaml
# 在config.yaml中添加安全配置
security:
  admin_secret: "strong-random-secret"
  rate_limit:
    enabled: true
    requests_per_minute: 100
  cors:
    origins: ["https://your-domain.com"]
    credentials: false
```

---

**版本**: 1.0  
**最后更新**: 2024年8月29日  
**维护者**: Privacy Gateway开发团队
