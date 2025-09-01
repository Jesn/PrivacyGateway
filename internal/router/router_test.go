package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

func setupRouterTest() *Router {
	cfg := &config.Config{
		AdminSecret: "test-secret",
		Port:        "10805",
	}
	log := logger.New()
	storage := proxyconfig.NewMemoryStorage(100)

	return NewRouter(cfg, log, nil, storage)
}

func TestRouter_SetupRoutes(t *testing.T) {
	router := setupRouterTest()

	// 设置路由
	router.SetupRoutes()

	// 验证路由信息
	routeInfo := router.GetRouteInfo()

	// 检查主要路由
	mainRoutes := routeInfo["routes"].(map[string]interface{})["main"].(map[string]string)
	if mainRoutes["/"] != "静态文件服务 / 子域名代理" {
		t.Error("Root route not properly configured")
	}
	if mainRoutes["/proxy"] != "HTTP代理服务" {
		t.Error("Proxy route not properly configured")
	}

	// 检查API路由
	apiRoutes := routeInfo["routes"].(map[string]interface{})["api"].(map[string]string)
	expectedAPIRoutes := []string{
		"/config/proxy",
		"/config/proxy/export",
		"/config/proxy/import",
		"/config/proxy/batch",
		"/config/proxy/{configID}/tokens",
		"/config/proxy/{configID}/tokens/{tokenID}",
	}

	for _, route := range expectedAPIRoutes {
		if _, exists := apiRoutes[route]; !exists {
			t.Errorf("API route %s not found", route)
		}
	}

	// 检查认证配置
	auth := routeInfo["authentication"].(map[string]interface{})
	adminAuth := auth["admin"].(map[string]string)
	if adminAuth["header"] != "X-Log-Secret" {
		t.Error("Admin auth header not properly configured")
	}

	tokenAuth := auth["token"].(map[string]string)
	if tokenAuth["header_primary"] != "X-Proxy-Token" {
		t.Error("Token auth header not properly configured")
	}

	// 检查CORS配置
	cors := routeInfo["cors"].(map[string]interface{})
	if cors["enabled"] != true {
		t.Error("CORS not enabled")
	}
	if cors["origins"] != "*" {
		t.Error("CORS origins not properly configured")
	}
}

func TestRouter_CORSHeaders(t *testing.T) {
	router := setupRouterTest()

	tests := []struct {
		name     string
		path     string
		method   string
		expected map[string]string
	}{
		{
			name:   "proxy endpoint CORS",
			path:   "/proxy",
			method: "OPTIONS",
			expected: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":       "86400",
			},
		},
		{
			name:   "API endpoint CORS",
			path:   "/config/proxy",
			method: "OPTIONS",
			expected: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Max-Age":       "86400",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// 根据路径调用相应的处理器
			switch {
			case tt.path == "/proxy":
				router.handleHTTPProxy(w, req)
			case strings.HasPrefix(tt.path, "/config/proxy"):
				router.handleProxyConfigAPI(w, req)
			}

			// 验证CORS头
			for header, expectedValue := range tt.expected {
				actualValue := w.Header().Get(header)
				if actualValue != expectedValue {
					t.Errorf("Expected %s: %s, got: %s", header, expectedValue, actualValue)
				}
			}

			// OPTIONS请求应该返回200
			if tt.method == "OPTIONS" && w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for OPTIONS request, got %d", w.Code)
			}
		})
	}
}

func TestRouter_TokenAPIRouting(t *testing.T) {
	router := setupRouterTest()

	// 创建测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	router.configStorage.Add(config)

	tests := []struct {
		name           string
		path           string
		method         string
		isTokenAPI     bool
		expectedStatus int
	}{
		{
			name:           "token list endpoint",
			path:           "/config/proxy/" + config.ID + "/tokens",
			method:         "GET",
			isTokenAPI:     true,
			expectedStatus: http.StatusUnauthorized, // 无认证信息
		},
		{
			name:           "token create endpoint",
			path:           "/config/proxy/" + config.ID + "/tokens",
			method:         "POST",
			isTokenAPI:     true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "config endpoint (not token API)",
			path:           "/config/proxy/" + config.ID,
			method:         "GET",
			isTokenAPI:     false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "config list endpoint",
			path:           "/config/proxy",
			method:         "GET",
			isTokenAPI:     false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// 调用通用的配置/令牌API处理器
			router.handleProxyConfigOrTokenAPI(w, req)

			// 验证状态码
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// 验证是否正确路由到令牌API
			if tt.isTokenAPI {
				// 令牌API应该返回JSON错误响应
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Expected JSON response for token API, got: %s", w.Body.String())
				}
			}
		})
	}
}

func TestRouter_AuthenticationHeaders(t *testing.T) {
	router := setupRouterTest()

	tests := []struct {
		name           string
		path           string
		method         string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:   "admin secret authentication",
			path:   "/config/proxy",
			method: "GET",
			headers: map[string]string{
				"X-Log-Secret": "test-secret",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid admin secret",
			path:   "/config/proxy",
			method: "GET",
			headers: map[string]string{
				"X-Log-Secret": "wrong-secret",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no authentication",
			path:           "/config/proxy",
			method:         "GET",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// 添加请求头
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.handleProxyConfigAPI(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRouter_RouteConflicts(t *testing.T) {
	router := setupRouterTest()

	// 验证路由信息中没有冲突
	routeInfo := router.GetRouteInfo()

	// 检查是否有重复的路由路径
	allRoutes := make(map[string]bool)

	routes := routeInfo["routes"].(map[string]interface{})
	for category, routeMap := range routes {
		if category == "main" || category == "api" || category == "logs" {
			for path := range routeMap.(map[string]string) {
				if allRoutes[path] {
					t.Errorf("Duplicate route found: %s", path)
				}
				allRoutes[path] = true
			}
		}
	}

	// 验证关键路由存在
	requiredRoutes := []string{
		"/",
		"/proxy",
		"/config/proxy",
		"/config/proxy/{configID}/tokens",
	}

	for _, route := range requiredRoutes {
		if !allRoutes[route] {
			t.Errorf("Required route missing: %s", route)
		}
	}
}

func TestRouter_PrintRoutes(t *testing.T) {
	router := setupRouterTest()

	// 这个测试主要验证PrintRoutes方法不会panic
	// 实际的输出验证需要更复杂的日志捕获机制
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintRoutes panicked: %v", r)
		}
	}()

	router.PrintRoutes()
}

func TestRouter_MiddlewareIntegration(t *testing.T) {
	router := setupRouterTest()

	// 测试中间件是否正确应用（主要是CORS）
	req := httptest.NewRequest("GET", "/proxy", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.handleHTTPProxy(w, req)

	// 验证CORS头是否被添加
	corsHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Expose-Headers",
		"Access-Control-Max-Age",
	}

	for _, header := range corsHeaders {
		if w.Header().Get(header) == "" {
			t.Errorf("CORS header %s not set", header)
		}
	}
}
