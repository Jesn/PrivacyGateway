package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

func TestProxyAuthenticator_AuthenticateForProxy_AdminSecret(t *testing.T) {
	// 创建测试存储
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试管理员密钥认证（请求头）
	req := httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Log-Secret", "test-secret")

	result := authenticator.AuthenticateForProxy(req, "")
	if !result.Authenticated {
		t.Errorf("Expected admin authentication to succeed, got: %s", result.Error)
	}
	if result.Method != "admin" {
		t.Errorf("Expected method 'admin', got: %s", result.Method)
	}

	// 测试管理员密钥认证（查询参数）
	req = httptest.NewRequest("GET", "/proxy?target=https://example.com&secret=test-secret", nil)
	result = authenticator.AuthenticateForProxy(req, "")
	if !result.Authenticated {
		t.Errorf("Expected admin authentication to succeed, got: %s", result.Error)
	}

	// 测试错误的管理员密钥
	req = httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Log-Secret", "wrong-secret")
	result = authenticator.AuthenticateForProxy(req, "")
	if result.Authenticated {
		t.Error("Expected admin authentication to fail with wrong secret")
	}
}

func TestProxyAuthenticator_AuthenticateForProxy_Token(t *testing.T) {
	// 创建测试存储和配置
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()

	// 添加测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)

	// 创建测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Test Token",
		Description: "Test token for authentication",
	}
	token, tokenValue, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// 添加令牌到配置
	storage.AddToken(config.ID, token)

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试令牌认证（X-Proxy-Token头）
	req := httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Proxy-Token", tokenValue)

	result := authenticator.AuthenticateForProxy(req, config.ID)
	if !result.Authenticated {
		t.Errorf("Expected token authentication to succeed, got: %s", result.Error)
	}
	if result.Method != "token" {
		t.Errorf("Expected method 'token', got: %s", result.Method)
	}
	if result.ConfigID != config.ID {
		t.Errorf("Expected config ID %s, got: %s", config.ID, result.ConfigID)
	}

	// 测试令牌认证（Authorization Bearer头）
	req = httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("Authorization", "Bearer "+tokenValue)

	result = authenticator.AuthenticateForProxy(req, config.ID)
	if !result.Authenticated {
		t.Errorf("Expected Bearer token authentication to succeed, got: %s", result.Error)
	}

	// 测试令牌认证（查询参数）
	req = httptest.NewRequest("GET", "/proxy?target=https://example.com&token="+tokenValue, nil)
	result = authenticator.AuthenticateForProxy(req, config.ID)
	if !result.Authenticated {
		t.Errorf("Expected query token authentication to succeed, got: %s", result.Error)
	}

	// 测试无效令牌
	req = httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Proxy-Token", "invalid-token")

	result = authenticator.AuthenticateForProxy(req, config.ID)
	if result.Authenticated {
		t.Error("Expected invalid token authentication to fail")
	}
	if result.Method != "token" {
		t.Errorf("Expected method 'token', got: %s", result.Method)
	}
}

func TestProxyAuthenticator_AuthenticateForProxy_DisabledToken(t *testing.T) {
	// 创建测试存储和配置
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()

	// 添加测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)

	// 创建禁用的令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Disabled Token",
		Description: "Disabled test token",
	}
	token, tokenValue, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// 禁用令牌
	token.Enabled = false
	storage.AddToken(config.ID, token)

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试禁用令牌认证
	req := httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Proxy-Token", tokenValue)

	result := authenticator.AuthenticateForProxy(req, config.ID)
	if result.Authenticated {
		t.Error("Expected disabled token authentication to fail")
	}
	if result.ValidationResult == nil || result.ValidationResult.ErrorCode != "TOKEN_DISABLED" {
		t.Errorf("Expected TOKEN_DISABLED error, got: %v", result.ValidationResult)
	}
}

func TestProxyAuthenticator_AuthenticateForProxy_ExpiredToken(t *testing.T) {
	// 创建测试存储和配置
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()

	// 添加测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)

	// 创建令牌，然后手动设置为过期
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Expired Token",
		Description: "Expired test token",
	}
	token, tokenValue, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// 手动设置过期时间为过去
	expiredTime := time.Now().Add(-time.Hour)
	token.ExpiresAt = &expiredTime

	storage.AddToken(config.ID, token)

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试过期令牌认证
	req := httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Proxy-Token", tokenValue)

	result := authenticator.AuthenticateForProxy(req, config.ID)
	if result.Authenticated {
		t.Error("Expected expired token authentication to fail")
	}
	if result.ValidationResult == nil || result.ValidationResult.ErrorCode != "TOKEN_EXPIRED" {
		t.Errorf("Expected TOKEN_EXPIRED error, got: %v", result.ValidationResult)
	}
}

func TestProxyAuthenticator_AuthenticateForConfig(t *testing.T) {
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试管理员认证成功
	req := httptest.NewRequest("GET", "/config/proxy", nil)
	req.Header.Set("X-Log-Secret", "test-secret")

	result := authenticator.AuthenticateForConfig(req)
	if !result.Authenticated {
		t.Errorf("Expected config authentication to succeed, got: %s", result.Error)
	}
	if result.Method != "admin" {
		t.Errorf("Expected method 'admin', got: %s", result.Method)
	}

	// 测试认证失败
	req = httptest.NewRequest("GET", "/config/proxy", nil)
	req.Header.Set("X-Log-Secret", "wrong-secret")

	result = authenticator.AuthenticateForConfig(req)
	if result.Authenticated {
		t.Error("Expected config authentication to fail with wrong secret")
	}
}

func TestExtractConfigID(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		query    string
		header   string
		expected string
	}{
		{
			name:     "from URL path",
			path:     "/config/proxy/test-config-123/tokens",
			expected: "test-config-123",
		},
		{
			name:     "from query parameter",
			path:     "/proxy",
			query:    "config_id=query-config-456",
			expected: "query-config-456",
		},
		{
			name:     "from header",
			path:     "/proxy",
			header:   "header-config-789",
			expected: "header-config-789",
		},
		{
			name:     "no config ID",
			path:     "/proxy",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.query != "" {
				req.URL.RawQuery = tt.query
			}
			if tt.header != "" {
				req.Header.Set("X-Config-ID", tt.header)
			}

			result := ExtractConfigID(req)
			if result != tt.expected {
				t.Errorf("Expected config ID %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestProxyAuthenticator_Performance(t *testing.T) {
	// 创建测试存储和配置
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New() // 减少日志输出

	// 添加测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)

	authenticator := NewProxyAuthenticator("test-secret", storage, log)

	// 测试管理员认证性能
	req := httptest.NewRequest("GET", "/proxy?target=https://example.com", nil)
	req.Header.Set("X-Log-Secret", "test-secret")

	start := time.Now()
	result := authenticator.AuthenticateForProxy(req, config.ID)
	duration := time.Since(start)

	if !result.Authenticated {
		t.Errorf("Expected authentication to succeed, got: %s", result.Error)
	}

	// 验证认证时间小于10ms
	if duration > 10*time.Millisecond {
		t.Errorf("Authentication took too long: %v (expected < 10ms)", duration)
	}

	t.Logf("Admin authentication took: %v", duration)
}
