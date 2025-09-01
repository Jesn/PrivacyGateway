package logviewer

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"privacygateway/internal/accesslog"
)

// TemplateData 模板数据
type TemplateData struct {
	Title        string                  `json:"title"`
	Logs         []accesslog.AccessLog   `json:"logs"`
	Filter       *FilterParams           `json:"filter"`
	Pagination   *PaginationData         `json:"pagination"`
	Stats        *accesslog.StorageStats `json:"stats"`
	StatusGroups map[string][]int        `json:"status_groups"`
	Error        string                  `json:"error,omitempty"`
	LogRecord200 bool                    `json:"log_record_200"` // 是否记录200状态码详情
}

// PaginationData 分页数据
type PaginationData struct {
	CurrentPage  int    `json:"current_page"`
	TotalPages   int    `json:"total_pages"`
	TotalItems   int    `json:"total_items"`
	ItemsPerPage int    `json:"items_per_page"`
	HasPrev      bool   `json:"has_prev"`
	HasNext      bool   `json:"has_next"`
	PrevPage     int    `json:"prev_page"`
	NextPage     int    `json:"next_page"`
	QueryString  string `json:"query_string"`
}

// LogViewTemplate 日志查看模板
const LogViewTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Privacy Gateway Logs</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; }
        .container { max-width: 1280px; margin: 0 auto; padding: 20px; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header h1 { color: #333; margin-bottom: 10px; }
        .stats { display: flex; gap: 20px; flex-wrap: wrap; }
        .stat-item { background: #f8f9fa; padding: 10px 15px; border-radius: 6px; }
        .stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
        .stat-value { font-size: 18px; font-weight: bold; color: #333; }
        
        .filters { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .filter-row { display: flex; gap: 15px; margin-bottom: 15px; flex-wrap: wrap; align-items: end; }
        .filter-group { flex: 1; min-width: 200px; }
        .filter-group label { display: block; margin-bottom: 5px; font-weight: 500; color: #333; }
        .filter-group input, .filter-group select { width: 100%; padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; }
        .filter-actions { display: flex; gap: 10px; }
        .btn { padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; text-decoration: none; display: inline-block; }
        .btn-primary { background: #007bff; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .btn:hover { opacity: 0.9; }
        
        .logs-container { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden; }
        .logs-table { width: 100%; border-collapse: collapse; }
        .logs-table th, .logs-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        .logs-table th { background: #f8f9fa; font-weight: 600; color: #333; }
        .logs-table tr:hover { background: #f8f9fa; }
        .status-badge { padding: 2px 6px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .status-2xx { background: #d4edda; color: #155724; }
        .status-3xx { background: #d1ecf1; color: #0c5460; }
        .status-4xx { background: #f8d7da; color: #721c24; }
        .status-5xx { background: #f5c6cb; color: #721c24; }
        .method-badge { padding: 2px 6px; border-radius: 3px; font-size: 11px; font-weight: bold; background: #e9ecef; color: #495057; }
        .type-badge { padding: 2px 6px; border-radius: 3px; font-size: 11px; font-weight: bold; }
        .type-HTTP { background: #d1ecf1; color: #0c5460; }
        .type-HTTPS { background: #d4edda; color: #155724; }
        .type-WebSocket { background: #fff3cd; color: #856404; }
        .type-SSE { background: #e2e3ff; color: #383d41; }
        .clickable-target { cursor: pointer; color: #007bff; text-decoration: none; }
        .clickable-target:hover { text-decoration: underline; }
        .non-clickable { color: #333; }
        
        .pagination { display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 20px; }
        .pagination a, .pagination span { padding: 8px 12px; border: 1px solid #ddd; border-radius: 4px; text-decoration: none; color: #333; }
        .pagination a:hover { background: #f8f9fa; }
        .pagination .current { background: #007bff; color: white; border-color: #007bff; }
        .pagination .disabled { color: #ccc; cursor: not-allowed; }
        
        .error { background: #f8d7da; color: #721c24; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
        .empty { text-align: center; padding: 40px; color: #666; }

        /* 弹窗样式 */
        .modal { display: none; position: fixed; z-index: 1000; left: 0; top: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.5); }
        .modal-content { background-color: white; margin: 5% auto; padding: 0; border-radius: 8px; width: 90%; max-width: 800px; max-height: 80vh; overflow: hidden; }
        .modal-header { background: #f8f9fa; padding: 20px; border-bottom: 1px solid #dee2e6; display: flex; justify-content: space-between; align-items: center; }
        .modal-header h2 { margin: 0; color: #333; font-size: 18px; }
        .modal-body { padding: 20px; max-height: 60vh; overflow-y: auto; }
        .close { color: #aaa; font-size: 28px; font-weight: bold; cursor: pointer; }
        .close:hover { color: #000; }
        .detail-row { margin-bottom: 15px; }
        .detail-label { font-weight: bold; color: #555; margin-bottom: 5px; }
        .detail-value { background: #f8f9fa; padding: 10px; border-radius: 4px; font-family: monospace; word-break: break-all; }
        .status-detail { display: inline-block; padding: 4px 8px; border-radius: 4px; font-weight: bold; }
        .copy-btn { background: #007bff; color: white; border: none; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 12px; transition: background-color 0.2s; }
        .copy-btn:hover { background: #0056b3; }
        .copy-btn:active { background: #004085; }

        @media (max-width: 768px) {
            .filter-row { flex-direction: column; }
            .filter-group { min-width: auto; }
            .logs-table { font-size: 14px; }
            .logs-table th, .logs-table td { padding: 8px; }
            .modal-content { width: 95%; margin: 10% auto; }
            .modal-body { max-height: 50vh; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                <h1>{{.Title}}</h1>
                <button onclick="logout()" class="btn btn-secondary" style="margin: 0;">退出登录</button>
            </div>
            {{if .Stats}}
            <div class="stats">
                <div class="stat-item">
                    <div class="stat-label">总日志数</div>
                    <div class="stat-value">{{.Stats.CurrentEntries}}</div>
                </div>
                <div class="stat-item">
                    <div class="stat-label">内存使用</div>
                    <div class="stat-value">{{printf "%.2f MB" .Stats.MemoryUsageMB}}</div>
                </div>
                <div class="stat-item">
                    <div class="stat-label">清理次数</div>
                    <div class="stat-value">{{.Stats.CleanupCount}}</div>
                </div>
                {{if .Stats.NewestEntry}}
                <div class="stat-item">
                    <div class="stat-label">最新日志</div>
                    <div class="stat-value">{{formatTime .Stats.NewestEntry}}</div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        {{if .Error}}
        <div class="error">{{.Error}}</div>
        {{end}}

        <div class="filters">
            <form id="filterForm" method="GET">
                <div class="filter-row">
                    <div class="filter-group">
                        <label for="domain">域名</label>
                        <input type="text" id="domain" name="domain" value="{{.Filter.Domain}}" placeholder="例如: httpbin.org">
                    </div>
                    <div class="filter-group">
                        <label for="status">状态码</label>
                        <select id="status" name="status">
                            <option value="">全部状态码</option>
                            <option value="2xx">2xx (成功)</option>
                            <option value="3xx">3xx (重定向)</option>
                            <option value="4xx">4xx (客户端错误)</option>
                            <option value="5xx">5xx (服务器错误)</option>
                            <option value="200">200</option>
                            <option value="404">404</option>
                            <option value="500">500</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label for="from">开始时间</label>
                        <input type="datetime-local" id="from" name="from" value="{{formatDateTime .Filter.FromTime}}">
                    </div>
                    <div class="filter-group">
                        <label for="to">结束时间</label>
                        <input type="datetime-local" id="to" name="to" value="{{formatDateTime .Filter.ToTime}}">
                    </div>
                </div>
                <div class="filter-row">
                    <div class="filter-group">
                        <label for="search">搜索</label>
                        <input type="text" id="search" name="search" value="{{.Filter.Search}}" placeholder="搜索路径、IP等">
                    </div>
                    <div class="filter-group">
                        <label for="limit">每页条数</label>
                        <select id="limit" name="limit">
                            <option value="25">25</option>
                            <option value="50" selected>50</option>
                            <option value="100">100</option>
                        </select>
                    </div>
                    <div class="filter-actions">
                        <button type="submit" class="btn btn-primary">筛选</button>
                        <a href="#" onclick="resetFilters()" class="btn btn-secondary">重置</a>
                    </div>
                </div>
            </form>
        </div>

        <div class="logs-container">
            {{if .Logs}}
            <table class="logs-table">
                <thead>
                    <tr>
                        <th>方法</th>
                        <th>类型</th>
                        <th>目标</th>
                        <th>状态码</th>
                        <th>耗时</th>
                        <th>客户端IP</th>
                        <th>时间</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Logs}}
                    <tr>
                        <td><span class="method-badge">{{.Method}}</span></td>
                        <td><span class="type-badge type-{{getTypeClass .RequestType}}">{{.RequestType}}</span></td>
                        <td>
                            {{if or (ne .StatusCode 200) $.LogRecord200}}
                            <a href="#" class="clickable-target" onclick="showLogDetailById('{{.ID}}'); return false;">
                                {{.TargetHost}}{{.TargetPath}}
                            </a>
                            {{else}}
                            <span class="non-clickable">{{.TargetHost}}{{.TargetPath}}</span>
                            {{end}}
                        </td>
                        <td><span class="status-badge status-{{getStatusClass .StatusCode}}">{{.StatusCode}}</span></td>
                        <td>{{.Duration}}ms</td>
                        <td>{{.ClientIP}}</td>
                        <td>{{formatLogTime .Timestamp}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
            {{else}}
            <div class="empty">
                <p>没有找到匹配的日志记录</p>
                <p style="margin-top: 10px; color: #999;">尝试调整筛选条件或检查时间范围</p>
            </div>
            {{end}}
        </div>

        {{if .Pagination}}
        <div class="pagination">
            {{if .Pagination.HasPrev}}
            <a href="?page={{.Pagination.PrevPage}}&{{.Pagination.QueryString}}">上一页</a>
            {{else}}
            <span class="disabled">上一页</span>
            {{end}}
            
            <span class="current">第 {{.Pagination.CurrentPage}} 页 / 共 {{.Pagination.TotalPages}} 页</span>
            
            {{if .Pagination.HasNext}}
            <a href="?page={{.Pagination.NextPage}}&{{.Pagination.QueryString}}">下一页</a>
            {{else}}
            <span class="disabled">下一页</span>
            {{end}}
            
            <span style="margin-left: 20px; color: #666;">
                共 {{.Pagination.TotalItems}} 条记录
            </span>
        </div>
        {{end}}
    </div>

    <!-- 详情弹窗 -->
    <div id="logDetailModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>请求详情</h2>
                <span class="close" onclick="closeModal()">&times;</span>
            </div>
            <div class="modal-body">
                <div class="detail-row">
                    <div class="detail-label">请求ID</div>
                    <div class="detail-value" id="detail-id"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">请求方法</div>
                    <div class="detail-value" id="detail-method"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">目标地址</div>
                    <div class="detail-value" id="detail-target"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">状态码</div>
                    <div class="detail-value" id="detail-status"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">请求时间</div>
                    <div class="detail-value" id="detail-time"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">处理耗时</div>
                    <div class="detail-value" id="detail-duration"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">客户端IP</div>
                    <div class="detail-value" id="detail-ip"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">User-Agent</div>
                    <div class="detail-value" id="detail-useragent"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">代理服务器</div>
                    <div class="detail-value" id="detail-proxy"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label">响应内容</div>
                    <div class="detail-value" id="detail-response"></div>
                </div>
                <div class="detail-row">
                    <div class="detail-label" style="display: flex; justify-content: space-between; align-items: center;">
                        <span>等效curl命令</span>
                        <button onclick="copyCurlCommand()" class="copy-btn" title="复制命令">📋</button>
                    </div>
                    <div class="detail-value" id="detail-curl" style="background: #f8f9fa; font-family: monospace; font-size: 12px; white-space: pre-wrap; word-break: break-all;"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // 密钥管理
        const LogAuth = {
            // 从localStorage获取加密的密钥
            getSecret: function() {
                return localStorage.getItem('log_viewer_secret');
            },

            // 保存加密的密钥到localStorage
            setSecret: function(secret) {
                // 简单的Base64编码（在生产环境中应使用更强的加密）
                const encoded = btoa(secret + ':' + Date.now());
                localStorage.setItem('log_viewer_secret', encoded);
            },

            // 解码密钥
            decodeSecret: function(encoded) {
                try {
                    const decoded = atob(encoded);
                    return decoded.split(':')[0];
                } catch (e) {
                    return null;
                }
            },

            // 清除密钥
            clearSecret: function() {
                localStorage.removeItem('log_viewer_secret');
            },

            // 检查是否已认证
            isAuthenticated: function() {
                return this.getSecret() !== null;
            }
        };

        // 页面加载时的初始化
        document.addEventListener('DOMContentLoaded', function() {
            // 如果URL中有secret参数，保存到localStorage并清除URL参数
            const urlParams = new URLSearchParams(window.location.search);
            const secretFromUrl = urlParams.get('secret');

            if (secretFromUrl) {
                LogAuth.setSecret(secretFromUrl);
                // 移除URL中的secret参数，避免在URL中暴露密钥
                urlParams.delete('secret');
                const newUrl = window.location.pathname + (urlParams.toString() ? '?' + urlParams.toString() : '');
                window.history.replaceState({}, '', newUrl);
            }

            // 为所有请求添加认证头
            setupAuthHeaders();
        });

        // 设置认证头
        function setupAuthHeaders() {
            const secret = LogAuth.getSecret();
            if (secret) {
                const decodedSecret = LogAuth.decodeSecret(secret);
                if (decodedSecret) {
                    // 拦截所有fetch请求
                    const originalFetch = window.fetch;
                    window.fetch = function(url, options = {}) {
                        options.headers = options.headers || {};
                        options.headers['X-Log-Secret'] = decodedSecret;
                        return originalFetch(url, options);
                    };

                    // 拦截所有XMLHttpRequest
                    const originalXHR = XMLHttpRequest.prototype.open;
                    XMLHttpRequest.prototype.open = function(method, url, async, user, password) {
                        this.addEventListener('readystatechange', function() {
                            if (this.readyState === 1) { // OPENED
                                this.setRequestHeader('X-Log-Secret', decodedSecret);
                            }
                        });
                        return originalXHR.apply(this, arguments);
                    };
                }
            }
        }

        // 表单提交处理
        document.getElementById('filterForm').addEventListener('submit', function(e) {
            const secret = LogAuth.getSecret();
            if (secret) {
                const decodedSecret = LogAuth.decodeSecret(secret);
                if (decodedSecret) {
                    // 创建一个隐藏的input来传递密钥
                    const secretInput = document.createElement('input');
                    secretInput.type = 'hidden';
                    secretInput.name = 'secret';
                    secretInput.value = decodedSecret;
                    this.appendChild(secretInput);
                }
            }
        });

        // 重置筛选
        function resetFilters() {
            const secret = LogAuth.getSecret();
            if (secret) {
                const decodedSecret = LogAuth.decodeSecret(secret);
                window.location.href = '/logs?secret=' + encodeURIComponent(decodedSecret);
            } else {
                window.location.href = '/logs';
            }
        }

        // 退出登录
        function logout() {
            if (confirm('确定要退出登录吗？')) {
                LogAuth.clearSecret();
                // 调用服务器端退出接口
                fetch('/logs/logout', {
                    method: 'POST',
                    headers: {
                        'X-Log-Secret': LogAuth.decodeSecret(LogAuth.getSecret()) || ''
                    }
                }).then(() => {
                    window.location.href = '/logs';
                }).catch(() => {
                    // 即使服务器端失败，也清除本地存储
                    window.location.href = '/logs';
                });
            }
        }

        // 通过ID显示日志详情弹窗
        function showLogDetailById(logId) {
            // 通过API获取详细信息
            fetch('/logs/api?id=' + encodeURIComponent(logId), {
                headers: {
                    'X-Log-Secret': LogAuth.decodeSecret(LogAuth.getSecret()) || ''
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.logs && data.logs.length > 0) {
                    const log = data.logs[0];
                    showLogDetail(log);
                } else {
                    alert('无法获取日志详情');
                }
            })
            .catch(error => {
                console.error('获取日志详情失败:', error);
                alert('获取日志详情失败');
            });
        }

        // 显示日志详情弹窗
        function showLogDetail(log) {
            document.getElementById('detail-id').textContent = log.id;
            document.getElementById('detail-method').textContent = log.method;
            document.getElementById('detail-target').textContent = log.target_host + log.target_path;

            // 设置状态码样式
            const statusElement = document.getElementById('detail-status');
            statusElement.innerHTML = '<span class="status-badge status-' + getStatusClass(log.status_code) + '">' + log.status_code + '</span>';

            document.getElementById('detail-time').textContent = formatLogTime(log.timestamp);
            document.getElementById('detail-duration').textContent = log.duration_ms + 'ms';
            document.getElementById('detail-ip').textContent = log.client_ip || '未知';
            document.getElementById('detail-useragent').textContent = log.user_agent || '未设置';
            document.getElementById('detail-proxy').textContent = log.proxy_info || 'Privacy Gateway';
            document.getElementById('detail-response').textContent = log.response_body || '无响应内容';

            // 生成等效的curl命令
            const curlCommand = generateCurlCommand(log.method, log.target_host + log.target_path, log.request_headers, log.request_body);
            document.getElementById('detail-curl').textContent = curlCommand;

            document.getElementById('logDetailModal').style.display = 'block';
        }

        // 生成等效的curl命令
        function generateCurlCommand(method, target, requestHeaders, requestBody) {
            let curl = 'curl';

            // 添加方法（如果不是GET）
            if (method !== 'GET') {
                curl += ' -X ' + method;
            }

            // 添加所有请求头
            if (requestHeaders) {
                for (const [key, value] of Object.entries(requestHeaders)) {
                    // 转义引号
                    const escapedValue = value.replace(/"/g, '\\"');
                    curl += ' -H "' + key + ': ' + escapedValue + '"';
                }
            }

            // 添加请求体（如果有）
            if (requestBody && requestBody.trim()) {
                // 转义单引号和双引号
                const escapedBody = requestBody.replace(/'/g, "'\"'\"'").replace(/"/g, '\\"');
                curl += " -d '" + escapedBody + "'";
            }

            // 添加目标URL
            let fullUrl;
            if (target.startsWith('http://') || target.startsWith('https://')) {
                fullUrl = target;
            } else {
                // 如果target不包含协议，需要根据实际情况添加
                fullUrl = 'https://' + target;
            }
            curl += ' "' + fullUrl + '"';

            return curl;
        }

        // 关闭弹窗
        function closeModal() {
            document.getElementById('logDetailModal').style.display = 'none';
        }

        // 获取状态码样式类
        function getStatusClass(status) {
            if (status >= 200 && status < 300) return '2xx';
            if (status >= 300 && status < 400) return '3xx';
            if (status >= 400 && status < 500) return '4xx';
            if (status >= 500) return '5xx';
            return 'other';
        }

        // 点击弹窗外部关闭
        window.onclick = function(event) {
            const modal = document.getElementById('logDetailModal');
            if (event.target === modal) {
                closeModal();
            }
        }

        // ESC键关闭弹窗
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                closeModal();
            }
        });

        // 格式化日志时间
        function formatLogTime(timestamp) {
            const date = new Date(timestamp);
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const day = String(date.getDate()).padStart(2, '0');
            const hours = String(date.getHours()).padStart(2, '0');
            const minutes = String(date.getMinutes()).padStart(2, '0');
            return month + '-' + day + ' ' + hours + ':' + minutes;
        }

        // 复制curl命令到剪贴板
        function copyCurlCommand() {
            const curlElement = document.getElementById('detail-curl');
            const curlCommand = curlElement.textContent;

            // 使用现代的Clipboard API
            if (navigator.clipboard && window.isSecureContext) {
                navigator.clipboard.writeText(curlCommand).then(() => {
                    showCopySuccess();
                }).catch(err => {
                    console.error('复制失败:', err);
                    fallbackCopyTextToClipboard(curlCommand);
                });
            } else {
                // 降级方案
                fallbackCopyTextToClipboard(curlCommand);
            }
        }

        // 降级复制方案
        function fallbackCopyTextToClipboard(text) {
            const textArea = document.createElement("textarea");
            textArea.value = text;
            textArea.style.position = "fixed";
            textArea.style.left = "-999999px";
            textArea.style.top = "-999999px";
            document.body.appendChild(textArea);
            textArea.focus();
            textArea.select();

            try {
                const successful = document.execCommand('copy');
                if (successful) {
                    showCopySuccess();
                } else {
                    showCopyError();
                }
            } catch (err) {
                console.error('降级复制失败:', err);
                showCopyError();
            }

            document.body.removeChild(textArea);
        }

        // 显示复制成功提示
        function showCopySuccess() {
            const btn = document.querySelector('.copy-btn');
            const originalText = btn.textContent;
            btn.textContent = '✅';
            btn.style.background = '#28a745';
            setTimeout(() => {
                btn.textContent = originalText;
                btn.style.background = '#007bff';
            }, 1500);
        }

        // 显示复制失败提示
        function showCopyError() {
            const btn = document.querySelector('.copy-btn');
            const originalText = btn.textContent;
            btn.textContent = '❌';
            btn.style.background = '#dc3545';
            setTimeout(() => {
                btn.textContent = originalText;
                btn.style.background = '#007bff';
            }, 1500);
        }

        // 自动刷新功能
        function autoRefresh() {
            const refreshInterval = 30000; // 30秒
            setTimeout(() => {
                window.location.reload();
            }, refreshInterval);
        }

        // 如果没有筛选条件，启用自动刷新
        if (window.location.search.indexOf('domain=') === -1 &&
            window.location.search.indexOf('status=') === -1 &&
            window.location.search.indexOf('from=') === -1) {
            autoRefresh();
        }
    </script>
</body>
</html>`

// GetTemplate 获取模板
func GetTemplate() *template.Template {
	funcMap := template.FuncMap{
		"formatTime": func(timeStr string) string {
			if timeStr == "" {
				return "-"
			}
			if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
				return t.Format("01-02 15:04")
			}
			return timeStr
		},
		"formatLogTime": func(t time.Time) string {
			return t.Format("01-02 15:04:05")
		},
		"formatDateTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02T15:04")
		},
		"getStatusClass": func(status int) string {
			switch {
			case status >= 200 && status < 300:
				return "2xx"
			case status >= 300 && status < 400:
				return "3xx"
			case status >= 400 && status < 500:
				return "4xx"
			case status >= 500:
				return "5xx"
			default:
				return "other"
			}
		},
		"getTypeClass": func(requestType string) string {
			switch requestType {
			case "HTTP":
				return "HTTP"
			case "HTTPS":
				return "HTTPS"
			case "WebSocket":
				return "WebSocket"
			case "SSE":
				return "SSE"
			default:
				return "HTTP"
			}
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	tmpl := template.Must(template.New("logview").Funcs(funcMap).Parse(LogViewTemplate))
	return tmpl
}

// CreateTemplateData 创建模板数据
func CreateTemplateData(title string, logs []accesslog.AccessLog, filter *FilterParams, response *accesslog.LogResponse, stats *accesslog.StorageStats, logRecord200 bool) *TemplateData {
	data := &TemplateData{
		Title:        title,
		Logs:         logs,
		Filter:       filter,
		Stats:        stats,
		StatusGroups: GetStatusCodeGroups(),
		LogRecord200: logRecord200,
	}

	if response != nil {
		data.Pagination = &PaginationData{
			CurrentPage:  response.Page,
			TotalPages:   response.TotalPages,
			TotalItems:   response.Total,
			ItemsPerPage: response.Limit,
			HasPrev:      response.Page > 1,
			HasNext:      response.Page < response.TotalPages,
			PrevPage:     response.Page - 1,
			NextPage:     response.Page + 1,
		}

		// 生成查询字符串（不包含page参数）
		if filter != nil {
			builder := NewFilterBuilder()
			builder.params = filter
			queryString := builder.ToQueryString()
			// 移除page参数
			if strings.Contains(queryString, "page=") {
				parts := strings.Split(queryString, "&")
				var newParts []string
				for _, part := range parts {
					if !strings.HasPrefix(part, "page=") {
						newParts = append(newParts, part)
					}
				}
				queryString = strings.Join(newParts, "&")
			}
			data.Pagination.QueryString = queryString
		}
	}

	return data
}

// RenderError 渲染错误页面
func RenderError(title, errorMsg string) *TemplateData {
	return &TemplateData{
		Title: title,
		Error: errorMsg,
	}
}
