package logviewer

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/logger"
)

// Handler 日志查看处理器
type Handler struct {
	recorder      *accesslog.Recorder
	authenticator Authenticator
	logger        *logger.Logger
	template      *template.Template
}

// NewHandler 创建新的日志查看处理器
func NewHandler(recorder *accesslog.Recorder, secret string, log *logger.Logger) (*Handler, error) {
	// 创建认证器
	auth, err := CreateAuthenticator(secret)
	if err != nil {
		return nil, err
	}

	return &Handler{
		recorder:      recorder,
		authenticator: auth,
		logger:        log,
		template:      GetTemplate(),
	}, nil
}

// ServeHTTP 处理HTTP请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 认证检查
	if !h.authenticator.IsEnabled() {
		h.handleError(w, r, "Log viewing is not enabled", http.StatusServiceUnavailable)
		return
	}

	// 路由处理
	path := strings.TrimPrefix(r.URL.Path, "/logs")

	// logout路径不需要认证，直接处理
	if path == "/logout" {
		h.handleLogout(w, r)
		return
	}

	// 其他路径需要认证
	authResult := h.authenticator.Authenticate(r)
	if !authResult.Authenticated {
		if secretAuth, ok := h.authenticator.(*SecretAuthenticator); ok {
			secretAuth.handleAuthFailure(w, r, authResult)
		} else {
			h.handleAPIError(w, authResult.Error, http.StatusUnauthorized)
		}
		return
	}

	// 认证成功，设置安全Cookie（如果是通过表单或URL参数登录）
	if secret := r.FormValue("secret"); secret != "" {
		if secretAuth, ok := h.authenticator.(*SecretAuthenticator); ok {
			secretAuth.SetSecureCookie(w, secret)
		}
	} else if secret := r.URL.Query().Get("secret"); secret != "" {
		if secretAuth, ok := h.authenticator.(*SecretAuthenticator); ok {
			secretAuth.SetSecureCookie(w, secret)
		}
	}

	// 处理需要认证的路由
	switch {
	case path == "" || path == "/":
		h.handleLogView(w, r)
	case path == "/api" || strings.HasPrefix(path, "/api/"):
		h.handleAPI(w, r)
	case path == "/stats":
		h.handleStats(w, r)
	default:
		h.handleError(w, r, "Not found", http.StatusNotFound)
	}
}

// handleLogView 处理日志查看页面
func (h *Handler) handleLogView(w http.ResponseWriter, r *http.Request) {
	// 如果是POST请求，重定向到GET请求，避免重复提交
	if r.Method == "POST" {
		// 添加一个特殊的重定向页面，用于清除重试计数
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html>
<head>
    <title>登录成功</title>
    <meta charset="utf-8">
</head>
<body>
    <script>
        // 清除重试计数，表示登录成功
        sessionStorage.removeItem('login_retry_count');
        // 立即重定向到日志页面
        window.location.href = '/logs';
    </script>
    <p>登录成功，正在跳转...</p>
</body>
</html>`
		w.Write([]byte(html))
		return
	}

	// 构建筛选器
	filterBuilder := NewFilterBuilder().FromRequest(r)
	filter := filterBuilder.Build()

	// 验证筛选参数
	if err := ValidateFilter(filterBuilder.GetParams()); err != nil {
		h.renderErrorPage(w, "参数错误", err.Error())
		return
	}

	// 查询日志
	response, err := h.recorder.Query(filter)
	if err != nil {
		h.logger.Error("failed to query logs", "error", err)
		h.renderErrorPage(w, "查询失败", "无法获取日志数据")
		return
	}

	// 获取统计信息
	stats := h.recorder.GetStats()

	// 准备模板数据
	templateData := CreateTemplateData(
		"访问日志",
		response.Logs,
		filterBuilder.GetParams(),
		response,
		&stats.StorageStats,
		h.recorder.IsLogRecord200Enabled(),
	)

	// 渲染模板
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, templateData); err != nil {
		h.logger.Error("failed to render template", "error", err)
		h.handleError(w, r, "Template rendering failed", http.StatusInternalServerError)
	}
}

// handleAPI 处理API请求
func (h *Handler) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/logs/api")

	switch {
	case path == "" || path == "/":
		h.handleAPILogs(w, r)
	case path == "/stats":
		h.handleAPIStats(w, r)
	default:
		h.handleAPIError(w, "Not found", http.StatusNotFound)
	}
}

// handleAPILogs 处理API日志查询
func (h *Handler) handleAPILogs(w http.ResponseWriter, r *http.Request) {
	// 检查是否是按ID查询
	logID := r.URL.Query().Get("id")
	if logID != "" {
		// 按ID查询特定日志
		h.handleAPILogByID(w, r, logID)
		return
	}

	// 构建筛选器
	filterBuilder := NewFilterBuilder().FromRequest(r)
	filter := filterBuilder.Build()

	// 验证筛选参数
	if err := ValidateFilter(filterBuilder.GetParams()); err != nil {
		h.handleAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 查询日志
	response, err := h.recorder.Query(filter)
	if err != nil {
		h.logger.Error("failed to query logs via API", "error", err)
		h.handleAPIError(w, "Query failed", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode API response", "error", err)
		h.handleAPIError(w, "Encoding failed", http.StatusInternalServerError)
	}
}

// handleAPILogByID 处理按ID查询单个日志
func (h *Handler) handleAPILogByID(w http.ResponseWriter, r *http.Request, logID string) {
	// 创建一个大范围的查询来获取所有日志
	filter := &accesslog.LogFilter{
		Page:  1,
		Limit: 1000, // 查询足够多的日志
	}

	// 查询日志
	response, err := h.recorder.Query(filter)
	if err != nil {
		h.logger.Error("failed to query logs by ID", "error", err, "id", logID)
		h.handleAPIError(w, "Query failed", http.StatusInternalServerError)
		return
	}

	// 在结果中查找指定ID的日志
	var targetLog *accesslog.AccessLog
	for _, log := range response.Logs {
		if log.ID == logID {
			targetLog = &log
			break
		}
	}

	if targetLog == nil {
		h.handleAPIError(w, "Log not found", http.StatusNotFound)
		return
	}

	// 返回包含单个日志的响应
	singleLogResponse := &accesslog.LogResponse{
		Logs:       []accesslog.AccessLog{*targetLog},
		Total:      1,
		Page:       1,
		Limit:      1,
		TotalPages: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(singleLogResponse); err != nil {
		h.logger.Error("failed to encode API response", "error", err)
		h.handleAPIError(w, "Encoding failed", http.StatusInternalServerError)
	}
}

// handleAPIStats 处理API统计查询
func (h *Handler) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats := h.recorder.GetStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode stats response", "error", err)
		h.handleAPIError(w, "Encoding failed", http.StatusInternalServerError)
	}
}

// handleStats 处理统计页面
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := h.recorder.GetStats()

	// 检查是否请求JSON格式
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
		return
	}

	// 渲染HTML页面
	templateData := &TemplateData{
		Title: "系统统计",
		Stats: &stats.StorageStats,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, templateData); err != nil {
		h.logger.Error("failed to render stats template", "error", err)
		h.handleError(w, r, "Template rendering failed", http.StatusInternalServerError)
	}
}

// handleError 处理错误
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	// 检查是否是API请求
	if h.isAPIRequest(r) {
		h.handleAPIError(w, message, statusCode)
		return
	}

	// 渲染错误页面
	h.renderErrorPage(w, "错误", message)
}

// handleAPIError 处理API错误
func (h *Handler) handleAPIError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"success": false,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// renderErrorPage 渲染错误页面
func (h *Handler) renderErrorPage(w http.ResponseWriter, title, message string) {
	templateData := RenderError(title, message)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	if err := h.template.Execute(w, templateData); err != nil {
		h.logger.Error("failed to render error template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// isAPIRequest 判断是否是API请求
func (h *Handler) isAPIRequest(r *http.Request) bool {
	// 检查路径
	if strings.Contains(r.URL.Path, "/api") {
		return true
	}

	// 检查Accept头
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	// 检查查询参数
	if r.URL.Query().Get("format") == "json" {
		return true
	}

	return false
}

// GetRecorder 获取记录器（用于测试）
func (h *Handler) GetRecorder() *accesslog.Recorder {
	return h.recorder
}

// IsAuthEnabled 检查认证是否启用
func (h *Handler) IsAuthEnabled() bool {
	return h.authenticator.IsEnabled()
}

// CreateLogViewHandler 创建日志查看处理器的便捷函数
func CreateLogViewHandler(recorder *accesslog.Recorder, secret string, log *logger.Logger) http.HandlerFunc {
	handler, err := NewHandler(recorder, secret, log)
	if err != nil {
		log.Error("failed to create log view handler", "error", err)
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查是否是API请求
			isAPI := strings.Contains(r.Header.Get("Accept"), "application/json") ||
				strings.Contains(r.URL.Path, "/api")

			if isAPI {
				// API请求返回JSON错误
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				errorMsg := "Log viewer not available"
				if authErr, ok := err.(*AuthError); ok {
					errorMsg = authErr.Message
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   errorMsg,
					"success": false,
					"code":    "SERVICE_UNAVAILABLE",
				})
			} else {
				// Web请求返回HTML错误页面
				renderConfigErrorPage(w, err)
			}
		}
	}

	return handler.ServeHTTP
}

// renderConfigErrorPage 渲染配置错误页面
func renderConfigErrorPage(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)

	errorMsg := "日志查看器配置错误"
	detailMsg := "请检查服务器配置"

	if authErr, ok := err.(*AuthError); ok {
		detailMsg = authErr.Message
	}

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <title>配置错误 - Privacy Gateway</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .error-container {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
            width: 100%;
            max-width: 500px;
            text-align: center;
        }
        .error-icon {
            width: 60px;
            height: 60px;
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
            border-radius: 50%;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 24px;
            font-weight: bold;
        }
        h1 {
            color: #333;
            margin-bottom: 15px;
            font-size: 24px;
            font-weight: 600;
        }
        .error-detail {
            color: #666;
            margin-bottom: 30px;
            font-size: 16px;
            line-height: 1.5;
            background: #f8f9fa;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #ff6b6b;
        }
        .solution {
            background: #e3f2fd;
            padding: 20px;
            border-radius: 6px;
            border-left: 4px solid #2196f3;
            text-align: left;
            margin-bottom: 20px;
        }
        .solution h3 {
            color: #1976d2;
            margin-bottom: 10px;
            font-size: 16px;
        }
        .solution p {
            color: #555;
            font-size: 14px;
            line-height: 1.4;
            margin-bottom: 8px;
        }
        .solution code {
            background: #f5f5f5;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
            color: #d63384;
        }
        .footer {
            margin-top: 20px;
            padding-top: 20px;
            border-top: 1px solid #e1e5e9;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-icon">⚠️</div>
        <h1>` + errorMsg + `</h1>
        <div class="error-detail">` + detailMsg + `</div>

        <div class="solution">
            <h3>解决方案：</h3>
            <p>请确保环境变量 <code>ADMIN_SECRET</code> 设置正确：</p>
            <p>• 密钥长度至少8个字符</p>
            <p>• 密钥长度不超过256个字符</p>
            <p>• 示例：<code>ADMIN_SECRET=myadminsecret123</code></p>
        </div>

        <div class="footer">
            <p>Privacy Gateway 日志查看系统</p>
        </div>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

// handleLogout 处理退出登录
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// 清除Cookie
	cookie := &http.Cookie{
		Name:     "log_secret",
		Value:    "",
		Path:     "/logs",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // 立即过期
	}
	http.SetCookie(w, cookie)

	// 检查是否是API请求
	if h.isAPIRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "已退出登录",
		})
		return
	}

	// 返回包含清除localStorage的页面，然后重定向到登录页面
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>退出登录</title>
    <meta charset="utf-8">
</head>
<body>
    <script>
        // 清除所有本地存储
        localStorage.removeItem('log_viewer_secret');
        sessionStorage.removeItem('login_retry_count');

        // 立即重定向到登录页面
        window.location.href = '/logs';
    </script>
    <p>正在退出登录...</p>
</body>
</html>`
	w.Write([]byte(html))
}

// HealthCheck 健康检查处理器
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":             "ok",
		"auth_enabled":       h.authenticator.IsEnabled(),
		"recorder_available": h.recorder != nil,
	}

	if h.recorder != nil {
		stats := h.recorder.GetStats()
		health["log_count"] = stats.StorageStats.CurrentEntries
		health["memory_usage_mb"] = stats.StorageStats.MemoryUsageMB
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
