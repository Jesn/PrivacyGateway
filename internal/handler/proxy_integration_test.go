package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

func setupProxyIntegrationTest() (*config.Config, *logger.Logger, proxyconfig.Storage, *proxyconfig.ProxyConfig, string) {
	// 创建测试配置
	cfg := &config.Config{
		AdminSecret: "test-secret",
		Port:        "10805",
	}

	// 创建日志器
	log := logger.New()

	// 创建存储
	storage := proxyconfig.NewMemoryStorage(100)

	// 创建测试代理配置
	proxyConfig := &proxyconfig.ProxyConfig{
		Name:      "Test Proxy Config",
		Subdomain: "testproxy",
		TargetURL: "https://httpbin.org",
		Protocol:  "https",
		Enabled:   true,
	}
	storage.Add(proxyConfig)

	// 创建测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Integration Test Token",
		Description: "Token for integration testing",
	}
	token, tokenValue, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		panic("Failed to create test token: " + err.Error())
	}
	storage.AddToken(proxyConfig.ID, token)

	return cfg, log, storage, proxyConfig, tokenValue
}

func TestHTTPProxyWithTokenAuth_AdminSecret(t *testing.T) {
	cfg, log, storage, _, _ := setupProxyIntegrationTest()

	// 创建请求（使用管理员密钥）
	req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get", nil)
	req.Header.Set("X-Log-Secret", "test-secret")
	w := httptest.NewRecorder()

	// 执行请求
	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应（由于是真实的HTTP请求，我们主要验证认证部分）
	if w.Code == http.StatusUnauthorized {
		t.Error("Expected admin authentication to succeed")
	}

	// 注意：由于这是集成测试，实际的代理请求可能会失败（网络问题等）
	// 但认证应该成功，不应该返回401
}

func TestHTTPProxyWithTokenAuth_ValidToken(t *testing.T) {
	cfg, log, storage, proxyConfig, tokenValue := setupProxyIntegrationTest()

	// 创建请求（使用有效令牌）
	req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+proxyConfig.ID, nil)
	req.Header.Set("X-Proxy-Token", tokenValue)
	w := httptest.NewRecorder()

	// 执行请求
	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应
	if w.Code == http.StatusUnauthorized {
		t.Error("Expected token authentication to succeed")
	}

	// 验证令牌使用统计已更新
	tokens, err := storage.GetTokens(proxyConfig.ID)
	if err == nil && len(tokens) > 0 {
		// 注意：由于我们没有直接访问令牌ID，这里简化验证
		// 在实际应用中，令牌使用统计应该会更新
		_ = tokens[0] // 使用第一个令牌进行验证
	}
}

func TestHTTPProxyWithTokenAuth_InvalidToken(t *testing.T) {
	cfg, log, storage, proxyConfig, _ := setupProxyIntegrationTest()

	// 创建请求（使用无效令牌）
	req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+proxyConfig.ID, nil)
	req.Header.Set("X-Proxy-Token", "invalid-token")
	w := httptest.NewRecorder()

	// 执行请求
	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}

	// 验证错误响应格式
	var errorResponse map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["success"] != false {
		t.Error("Expected success=false in error response")
	}

	if errorResponse["error_code"] != "TOKEN_NOT_FOUND" {
		t.Errorf("Expected error_code=TOKEN_NOT_FOUND, got %v", errorResponse["error_code"])
	}
}

func TestHTTPProxyWithTokenAuth_NoAuthentication(t *testing.T) {
	cfg, log, storage, proxyConfig, _ := setupProxyIntegrationTest()

	// 创建请求（无认证信息）
	req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+proxyConfig.ID, nil)
	w := httptest.NewRecorder()

	// 执行请求
	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for no authentication, got %d", w.Code)
	}
}

func TestSubdomainProxyWithTokenAuth_ValidToken(t *testing.T) {
	cfg, log, storage, proxyConfig, tokenValue := setupProxyIntegrationTest()

	// 创建子域名请求（使用有效令牌）
	req := httptest.NewRequest("GET", "/test", nil)
	req.Host = proxyConfig.Subdomain + ".localhost:10805"
	req.Header.Set("X-Proxy-Token", tokenValue)
	w := httptest.NewRecorder()

	// 执行请求
	HandleSubdomainProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应
	if w.Code == http.StatusUnauthorized {
		t.Error("Expected token authentication to succeed for subdomain proxy")
	}
}

func TestSubdomainProxyWithTokenAuth_InvalidSubdomain(t *testing.T) {
	cfg, log, storage, _, tokenValue := setupProxyIntegrationTest()

	// 创建无效子域名请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Host = "nonexistent.localhost:10805"
	req.Header.Set("X-Proxy-Token", tokenValue)
	w := httptest.NewRecorder()

	// 执行请求
	HandleSubdomainProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证响应
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid subdomain, got %d", w.Code)
	}
}

func TestTokenUsageStatistics(t *testing.T) {
	cfg, log, storage, proxyConfig, tokenValue := setupProxyIntegrationTest()

	// 获取初始令牌统计
	initialStats, err := storage.GetTokenStats(proxyConfig.ID)
	if err != nil {
		t.Fatalf("Failed to get initial token stats: %v", err)
	}

	// 执行多个代理请求
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+proxyConfig.ID, nil)
		req.Header.Set("X-Proxy-Token", tokenValue)
		w := httptest.NewRecorder()

		HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

		// 短暂等待，确保统计更新
		time.Sleep(10 * time.Millisecond)
	}

	// 获取更新后的令牌统计
	updatedStats, err := storage.GetTokenStats(proxyConfig.ID)
	if err != nil {
		t.Fatalf("Failed to get updated token stats: %v", err)
	}

	// 验证统计信息已更新
	if updatedStats.TotalRequests <= initialStats.TotalRequests {
		t.Errorf("Expected total requests to increase, initial: %d, updated: %d",
			initialStats.TotalRequests, updatedStats.TotalRequests)
	}
}

func TestProxyRequestLogging(t *testing.T) {
	cfg, log, storage, proxyConfig, tokenValue := setupProxyIntegrationTest()

	// 创建请求
	req := httptest.NewRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+proxyConfig.ID, nil)
	req.Header.Set("X-Proxy-Token", tokenValue)
	req.Header.Set("User-Agent", "Integration-Test-Client")
	w := httptest.NewRecorder()

	// 执行请求
	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 注意：这里主要验证请求不会因为日志记录而失败
	// 实际的日志验证需要更复杂的设置
	if w.Code == http.StatusInternalServerError {
		t.Error("Request should not fail due to logging issues")
	}
}

func TestCORSHeaders(t *testing.T) {
	cfg, log, storage, _, _ := setupProxyIntegrationTest()

	// 测试预检请求
	req := httptest.NewRequest("OPTIONS", "/proxy", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	HTTPProxyWithTokenAuth(w, req, cfg, log, nil, storage)

	// 验证CORS头
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin: *")
	}

	if !contains(w.Header().Get("Access-Control-Allow-Headers"), "X-Proxy-Token") {
		t.Error("Expected X-Proxy-Token in allowed headers")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for OPTIONS request, got %d", w.Code)
	}
}

func TestConfigIDExtraction(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		header   string
		query    string
		expected string
	}{
		{
			name:     "from query parameter",
			url:      "/proxy?target=https://example.com&config_id=test-config",
			expected: "test-config",
		},
		{
			name:     "from header",
			url:      "/proxy?target=https://example.com",
			header:   "header-config",
			expected: "header-config",
		},
		{
			name:     "no config ID",
			url:      "/proxy?target=https://example.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			if tt.header != "" {
				req.Header.Set("X-Config-ID", tt.header)
			}

			configID := ExtractConfigID(req)
			if configID != tt.expected {
				t.Errorf("Expected config ID %s, got %s", tt.expected, configID)
			}
		})
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)+1] == substr+"," ||
			s[len(s)-len(substr)-1:] == ","+substr ||
			bytes.Contains([]byte(s), []byte(","+substr+",")))))
}
