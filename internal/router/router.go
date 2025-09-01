package router

import (
	"net/http"
	"strings"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/handler"
	"privacygateway/internal/logger"
	"privacygateway/internal/logviewer"
	"privacygateway/internal/proxyconfig"
)

// Router 路由器结构
type Router struct {
	cfg           *config.Config
	log           *logger.Logger
	recorder      *accesslog.Recorder
	configStorage proxyconfig.Storage
	tokenHandler  *handler.TokenAPIHandler
}

// NewRouter 创建新的路由器
func NewRouter(cfg *config.Config, log *logger.Logger, recorder *accesslog.Recorder, configStorage proxyconfig.Storage) *Router {
	tokenHandler := handler.NewTokenAPIHandler(configStorage, cfg.AdminSecret, log)

	return &Router{
		cfg:           cfg,
		log:           log,
		recorder:      recorder,
		configStorage: configStorage,
		tokenHandler:  tokenHandler,
	}
}

// SetupRoutes 设置所有路由
func (r *Router) SetupRoutes() {
	// 设置中间件
	r.setupMiddleware()

	// 设置主路由
	r.setupMainRoutes()

	// 设置API路由
	r.setupAPIRoutes()

	// 设置日志查看路由
	r.setupLogRoutes()
}

// setupMiddleware 设置全局中间件
func (r *Router) setupMiddleware() {
	// 这里可以添加全局中间件，如日志记录、CORS等
	// 由于使用的是标准库的http包，中间件需要在每个处理器中单独处理
}

// setupMainRoutes 设置主要路由
func (r *Router) setupMainRoutes() {
	// 根路径处理器 - 支持静态文件和子域名代理
	http.HandleFunc("/", r.HandleRoot)

	// HTTP代理路由
	http.HandleFunc("/proxy", r.HandleHTTPProxy)

	// WebSocket路由
	http.HandleFunc("/ws", r.HandleWebSocket)
}

// setupAPIRoutes 设置API路由
func (r *Router) setupAPIRoutes() {
	// 代理配置管理API
	http.HandleFunc("/config/proxy", r.HandleProxyConfigAPI)

	// 配置导入导出API
	http.HandleFunc("/config/proxy/export", r.HandleProxyConfigExportAPI)
	http.HandleFunc("/config/proxy/import", r.HandleProxyConfigImportAPI)

	// 批量操作API
	http.HandleFunc("/config/proxy/batch", r.HandleProxyConfigBatchAPI)

	// 令牌管理API（通用路由）
	http.HandleFunc("/config/proxy/", r.HandleProxyConfigOrTokenAPI)
}

// setupLogRoutes 设置日志查看路由
func (r *Router) setupLogRoutes() {
	if r.recorder != nil {
		// 验证日志查看器配置
		if _, err := logviewer.CreateAuthenticator(r.cfg.AdminSecret); err != nil {
			r.log.Error("log viewer configuration error", "error", err.Error())
			r.log.Info("log viewer disabled", "reason", "invalid secret configuration")
		}

		// 注册日志查看路由
		logHandler := logviewer.CreateLogViewHandler(r.recorder, r.cfg.AdminSecret, r.log)
		http.HandleFunc("/logs", logHandler)
		http.HandleFunc("/logs/", logHandler)
	}
}

// 路由处理器

// HandleRoot 处理根路径请求
func (r *Router) HandleRoot(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 处理静态文件
	handler.Static(w, req, r.log)
}

// HandleHTTPProxy 处理HTTP代理请求
func (r *Router) HandleHTTPProxy(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 使用支持令牌认证的HTTP代理处理器
	handler.HTTPProxyWithTokenAuth(w, req, r.cfg, r.log, r.recorder, r.configStorage)
}

// HandleWebSocket 处理WebSocket请求
func (r *Router) HandleWebSocket(w http.ResponseWriter, req *http.Request) {
	handler.WebSocket(w, req, r.cfg, r.log, r.recorder)
}

// HandleProxyConfigAPI 处理代理配置API请求
func (r *Router) HandleProxyConfigAPI(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	handler.HandleProxyConfigAPI(w, req, r.cfg, r.log, r.configStorage)
}

// HandleProxyConfigExportAPI 处理配置导出API请求
func (r *Router) HandleProxyConfigExportAPI(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	handler.HandleProxyConfigAPI(w, req, r.cfg, r.log, r.configStorage)
}

// HandleProxyConfigImportAPI 处理配置导入API请求
func (r *Router) HandleProxyConfigImportAPI(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	handler.HandleProxyConfigAPI(w, req, r.cfg, r.log, r.configStorage)
}

// HandleProxyConfigBatchAPI 处理批量操作API请求
func (r *Router) HandleProxyConfigBatchAPI(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	handler.HandleProxyConfigAPI(w, req, r.cfg, r.log, r.configStorage)
}

// HandleProxyConfigOrTokenAPI 处理代理配置或令牌API请求
func (r *Router) HandleProxyConfigOrTokenAPI(w http.ResponseWriter, req *http.Request) {
	// 添加CORS支持
	r.addCORSHeaders(w, req)

	// 处理预检请求
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 检查是否是令牌管理API请求
	if strings.Contains(req.URL.Path, "/tokens") {
		r.tokenHandler.HandleTokenAPI(w, req)
		return
	}

	// 否则交给配置管理API处理
	handler.HandleProxyConfigAPI(w, req, r.cfg, r.log, r.configStorage)
}

// addCORSHeaders 添加CORS头
func (r *Router) addCORSHeaders(w http.ResponseWriter, req *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Log-Secret, X-Proxy-Token, X-Config-ID")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24小时
}

// GetRouteInfo 获取路由信息
func (r *Router) GetRouteInfo() map[string]interface{} {
	return map[string]interface{}{
		"routes": map[string]interface{}{
			"main": map[string]string{
				"/":      "静态文件服务 / 子域名代理",
				"/proxy": "HTTP代理服务",
			},
			"api": map[string]string{
				"/config/proxy":                             "代理配置管理API",
				"/config/proxy/export":                      "配置导出API",
				"/config/proxy/import":                      "配置导入API",
				"/config/proxy/batch":                       "批量操作API",
				"/config/proxy/{configID}/tokens":           "令牌管理API - 列表/创建",
				"/config/proxy/{configID}/tokens/{tokenID}": "令牌管理API - 获取/更新/删除",
			},
			"logs": map[string]string{
				"/logs":  "访问日志查看",
				"/logs/": "访问日志详情",
			},
		},
		"authentication": map[string]interface{}{
			"admin": map[string]string{
				"header": "X-Log-Secret",
				"query":  "secret",
			},
			"token": map[string]string{
				"header":          "X-Proxy-Token",
				"query_parameter": "token",
			},
		},
		"cors": map[string]interface{}{
			"enabled": true,
			"origins": "*",
			"methods": []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			"headers": []string{
				"Content-Type",
				"Authorization",
				"X-Requested-With",
				"X-Log-Secret",
				"X-Proxy-Token",
				"X-Config-ID",
			},
		},
	}
}

// PrintRoutes 打印路由信息
func (r *Router) PrintRoutes() {
	r.log.Info("=== Privacy Gateway 路由配置 ===")

	r.log.Info("主要服务:")
	r.log.Info("  /           - 静态文件服务 / 子域名代理")
	r.log.Info("  /proxy      - HTTP代理服务")

	r.log.Info("API端点:")
	r.log.Info("  /config/proxy                              - 代理配置管理")
	r.log.Info("  /config/proxy/export                       - 配置导出")
	r.log.Info("  /config/proxy/import                       - 配置导入")
	r.log.Info("  /config/proxy/batch                        - 批量操作")
	r.log.Info("  /config/proxy/{configID}/tokens           - 令牌列表/创建")
	r.log.Info("  /config/proxy/{configID}/tokens/{tokenID} - 令牌操作")

	if r.recorder != nil {
		r.log.Info("日志服务:")
		r.log.Info("  /logs       - 访问日志查看")
		r.log.Info("  /logs/      - 访问日志详情")
	}

	r.log.Info("认证方式:")
	r.log.Info("  管理员密钥: X-Log-Secret 请求头 或 ?secret= 查询参数")
	r.log.Info("  令牌认证:   X-Proxy-Token 请求头 或 ?token= 查询参数")

	r.log.Info("CORS支持: 已启用，允许所有来源")
	r.log.Info("================================")
}
