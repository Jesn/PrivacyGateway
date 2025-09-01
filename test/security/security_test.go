package security

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
	"privacygateway/internal/router"
)

// SecurityTestSuite 安全测试套件
type SecurityTestSuite struct {
	server          *httptest.Server
	cfg             *config.Config
	log             *logger.Logger
	storage         proxyconfig.Storage
	adminSecret     string
	testConfigID    string
	validTokenID    string
	validTokenValue string
}

// SetupSecurityTest 设置安全测试环境
func SetupSecurityTest(t *testing.T) *SecurityTestSuite {
	// 创建测试配置
	cfg := &config.Config{
		AdminSecret: "security-test-secret-12345",
		Port:        "0",
	}

	// 创建日志器
	log := logger.New()

	// 创建存储
	storage := proxyconfig.NewMemoryStorage(100)

	// 创建新的ServeMux避免路由冲突
	mux := http.NewServeMux()

	// 创建路由器
	appRouter := router.NewRouter(cfg, log, nil, storage)

	// 手动设置路由到新的mux
	setupSecurityTestRoutes(mux, appRouter)

	// 创建测试服务器
	server := httptest.NewServer(mux)

	suite := &SecurityTestSuite{
		server:      server,
		cfg:         cfg,
		log:         log,
		storage:     storage,
		adminSecret: cfg.AdminSecret,
	}

	// 创建测试数据
	suite.createTestData(t)

	return suite
}

// setupSecurityTestRoutes 为安全测试设置路由
func setupSecurityTestRoutes(mux *http.ServeMux, appRouter *router.Router) {
	mux.HandleFunc("/", appRouter.HandleRoot)
	mux.HandleFunc("/proxy", appRouter.HandleHTTPProxy)
	mux.HandleFunc("/config/proxy", appRouter.HandleProxyConfigAPI)
	mux.HandleFunc("/config/proxy/", appRouter.HandleProxyConfigOrTokenAPI)
}

// TearDown 清理测试环境
func (suite *SecurityTestSuite) TearDown() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// createTestData 创建测试数据
func (suite *SecurityTestSuite) createTestData(t *testing.T) {
	// 创建测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Security Test Config",
		Subdomain: "sectest",
		TargetURL: "https://httpbin.org",
		Protocol:  "https",
		Enabled:   true,
	}
	suite.storage.Add(config)
	suite.testConfigID = config.ID

	// 创建测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Security Test Token",
		Description: "Token for security testing",
	}
	token, tokenValue, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	suite.storage.AddToken(config.ID, token)
	suite.validTokenID = token.ID
	suite.validTokenValue = tokenValue
}

// makeRequest 发送HTTP请求
func (suite *SecurityTestSuite) makeRequest(method, path string, body []byte, headers map[string]string) *http.Response {
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req, _ := http.NewRequest(method, suite.server.URL+path, reqBody)

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, _ := client.Do(req)
	return resp
}

// TestTokenSecurityValidation 测试令牌安全性验证
func TestTokenSecurityValidation(t *testing.T) {
	suite := SetupSecurityTest(t)
	defer suite.TearDown()

	t.Run("Token Format Validation", func(t *testing.T) {
		// 测试各种无效的令牌格式
		invalidTokens := []string{
			"",                              // 空令牌
			"short",                         // 太短的令牌
			"invalid-format",                // 无效格式
			"123456789012345",               // 纯数字
			"abcdefghijklmnop",              // 纯字母
			"../../../etc/passwd",           // 路径遍历
			"<script>alert('xss')</script>", // XSS尝试
			"'; DROP TABLE tokens; --",      // SQL注入尝试
			strings.Repeat("a", 1000),       // 超长令牌
		}

		for _, invalidToken := range invalidTokens {
			resp := suite.makeRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
				"X-Proxy-Token": invalidToken,
			})

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401 for invalid token '%s', got %d", invalidToken, resp.StatusCode)
			}
		}
	})

	t.Run("Token Timing Attack Protection", func(t *testing.T) {
		// 测试时序攻击防护
		validToken := suite.validTokenValue
		invalidToken := "invalid-token-same-length-as-valid-one-12345"

		// 测试多次请求的时间一致性
		var validTimes []time.Duration
		var invalidTimes []time.Duration

		for i := 0; i < 10; i++ {
			// 测试有效令牌
			start := time.Now()
			suite.makeRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
				"X-Proxy-Token": validToken,
			})
			validTimes = append(validTimes, time.Since(start))

			// 测试无效令牌
			start = time.Now()
			suite.makeRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
				"X-Proxy-Token": invalidToken,
			})
			invalidTimes = append(invalidTimes, time.Since(start))
		}

		// 计算平均时间
		var validAvg, invalidAvg time.Duration
		for i := 0; i < 10; i++ {
			validAvg += validTimes[i]
			invalidAvg += invalidTimes[i]
		}
		validAvg /= 10
		invalidAvg /= 10

		// 时间差不应该太大（防止时序攻击）
		timeDiff := validAvg - invalidAvg
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		t.Logf("Valid token avg time: %v, Invalid token avg time: %v, Diff: %v", validAvg, invalidAvg, timeDiff)

		// 时间差应该在合理范围内（小于100ms）
		if timeDiff > 100*time.Millisecond {
			t.Errorf("Timing difference too large: %v (potential timing attack vulnerability)", timeDiff)
		}
	})

	t.Run("Token Entropy Validation", func(t *testing.T) {
		// 测试令牌熵值
		tokens := make(map[string]bool)

		// 创建多个令牌并检查唯一性
		for i := 0; i < 100; i++ {
			tokenReq := map[string]interface{}{
				"name":        fmt.Sprintf("Entropy Test Token %d", i),
				"description": "Token for entropy testing",
			}

			reqBody, _ := json.Marshal(tokenReq)
			resp := suite.makeRequest("POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
				"X-Log-Secret": suite.adminSecret,
			})

			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("Failed to create token %d: status %d", i, resp.StatusCode)
			}

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)

			var tokenValue string
			if data, ok := response["data"].(map[string]interface{}); ok && data != nil {
				tokenValue = data["token"].(string)
			} else {
				tokenValue = response["token"].(string)
			}

			// 检查令牌唯一性
			if tokens[tokenValue] {
				t.Errorf("Duplicate token generated: %s", tokenValue)
			}
			tokens[tokenValue] = true

			// 检查令牌长度和格式
			if len(tokenValue) < 32 {
				t.Errorf("Token too short: %d characters", len(tokenValue))
			}

			// 检查令牌是否包含足够的随机性（简单检查）
			if strings.Contains(tokenValue, "000000") || strings.Contains(tokenValue, "111111") {
				t.Errorf("Token appears to have low entropy: %s", tokenValue)
			}
		}

		t.Logf("Generated %d unique tokens", len(tokens))
	})
}

// TestPermissionBypass 测试权限绕过
func TestPermissionBypass(t *testing.T) {
	suite := SetupSecurityTest(t)
	defer suite.TearDown()

	t.Run("Admin Secret Bypass Attempts", func(t *testing.T) {
		// 尝试各种管理员密钥绕过
		bypassAttempts := []string{
			"",                                 // 空密钥
			"admin",                            // 常见密钥
			"password",                         // 常见密钥
			"12345",                            // 简单密钥
			suite.adminSecret + "extra",        // 密钥后缀
			"prefix" + suite.adminSecret,       // 密钥前缀
			strings.ToUpper(suite.adminSecret), // 大写密钥
			strings.ToLower(suite.adminSecret), // 小写密钥
		}

		for _, attempt := range bypassAttempts {
			resp := suite.makeRequest("GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
				"X-Log-Secret": attempt,
			})

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Admin secret bypass succeeded with: '%s', status: %d", attempt, resp.StatusCode)
			}
		}
	})

	t.Run("Token Cross-Configuration Access", func(t *testing.T) {
		// 创建另一个配置
		config2 := &proxyconfig.ProxyConfig{
			Name:      "Security Test Config 2",
			Subdomain: "sectest2",
			TargetURL: "https://httpbin.org",
			Protocol:  "https",
			Enabled:   true,
		}
		suite.storage.Add(config2)

		// 尝试使用配置1的令牌访问配置2
		resp := suite.makeRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+config2.ID, nil, map[string]string{
			"X-Proxy-Token": suite.validTokenValue,
		})

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Cross-configuration access succeeded, status: %d", resp.StatusCode)
		}
	})

	t.Run("Header Injection Attempts", func(t *testing.T) {
		// 尝试各种头部注入攻击
		injectionAttempts := []map[string]string{
			{
				"X-Proxy-Token": suite.validTokenValue,
				"X-Log-Secret":  "injected-secret",
			},
			{
				"Authorization": "Bearer " + suite.validTokenValue,
				"X-Log-Secret":  suite.adminSecret,
			},
			{
				"X-Proxy-Token": suite.validTokenValue + "\r\nX-Log-Secret: " + suite.adminSecret,
			},
		}

		for i, headers := range injectionAttempts {
			resp := suite.makeRequest("GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, headers)

			// 应该只有管理员密钥能访问令牌管理API
			if resp.StatusCode == http.StatusOK {
				t.Errorf("Header injection attempt %d succeeded", i+1)
			}
		}
	})
}

// TestInjectionAttacks 测试注入攻击
func TestInjectionAttacks(t *testing.T) {
	suite := SetupSecurityTest(t)
	defer suite.TearDown()

	t.Run("SQL Injection in Token Creation", func(t *testing.T) {
		// 尝试SQL注入攻击
		sqlInjectionPayloads := []string{
			"'; DROP TABLE tokens; --",
			"' OR '1'='1",
			"'; UPDATE tokens SET enabled=false; --",
			"' UNION SELECT * FROM tokens --",
			"'; INSERT INTO tokens VALUES ('malicious'); --",
		}

		for _, payload := range sqlInjectionPayloads {
			tokenReq := map[string]interface{}{
				"name":        payload,
				"description": "SQL injection test",
			}

			reqBody, _ := json.Marshal(tokenReq)
			resp := suite.makeRequest("POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
				"X-Log-Secret": suite.adminSecret,
			})

			// 请求应该被正确处理（不应该导致SQL错误）
			if resp.StatusCode == http.StatusInternalServerError {
				t.Errorf("SQL injection payload caused server error: %s", payload)
			}
		}
	})

	t.Run("XSS in Token Fields", func(t *testing.T) {
		// 尝试XSS攻击
		xssPayloads := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"<svg onload=alert('xss')>",
			"';alert('xss');//",
		}

		for _, payload := range xssPayloads {
			tokenReq := map[string]interface{}{
				"name":        payload,
				"description": payload,
			}

			reqBody, _ := json.Marshal(tokenReq)
			resp := suite.makeRequest("POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
				"X-Log-Secret": suite.adminSecret,
			})

			if resp.StatusCode == http.StatusCreated {
				// 验证返回的数据是否被正确转义
				var response map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&response)

				// 检查响应中是否包含原始的恶意脚本
				responseStr := fmt.Sprintf("%v", response)
				if strings.Contains(responseStr, "<script>") || strings.Contains(responseStr, "javascript:") {
					t.Errorf("XSS payload not properly escaped: %s", payload)
				}
			}
		}
	})

	t.Run("Path Traversal in API Endpoints", func(t *testing.T) {
		// 尝试路径遍历攻击
		pathTraversalPayloads := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			"....//....//....//etc/passwd",
		}

		for _, payload := range pathTraversalPayloads {
			// 尝试在配置ID中使用路径遍历
			resp := suite.makeRequest("GET", "/config/proxy/"+payload+"/tokens", nil, map[string]string{
				"X-Log-Secret": suite.adminSecret,
			})

			// 应该返回404而不是500或其他错误
			if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Path traversal payload caused unexpected status: %s -> %d", payload, resp.StatusCode)
			}
		}
	})
}

// TestBruteForceProtection 测试暴力破解防护
func TestBruteForceProtection(t *testing.T) {
	suite := SetupSecurityTest(t)
	defer suite.TearDown()

	t.Run("Token Brute Force Protection", func(t *testing.T) {
		// 生成随机令牌进行暴力破解测试
		attempts := 100
		successCount := 0

		for i := 0; i < attempts; i++ {
			// 生成随机令牌
			randomBytes := make([]byte, 32)
			rand.Read(randomBytes)
			randomToken := hex.EncodeToString(randomBytes)

			resp := suite.makeRequest("GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
				"X-Proxy-Token": randomToken,
			})

			if resp.StatusCode == http.StatusOK {
				successCount++
			}

			// 短暂延迟避免过快请求
			time.Sleep(10 * time.Millisecond)
		}

		// 暴力破解成功率应该为0
		if successCount > 0 {
			t.Errorf("Brute force attack succeeded %d times out of %d attempts", successCount, attempts)
		}

		t.Logf("Brute force test: 0/%d successful attempts", attempts)
	})

	t.Run("Admin Secret Brute Force Protection", func(t *testing.T) {
		// 测试管理员密钥暴力破解
		commonSecrets := []string{
			"admin", "password", "123456", "admin123", "root",
			"secret", "12345678", "qwerty", "password123",
			"admin@123", "administrator", "pass", "test",
		}

		successCount := 0
		for _, secret := range commonSecrets {
			resp := suite.makeRequest("GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
				"X-Log-Secret": secret,
			})

			if resp.StatusCode == http.StatusOK {
				successCount++
			}

			time.Sleep(10 * time.Millisecond)
		}

		if successCount > 0 {
			t.Errorf("Admin secret brute force succeeded %d times", successCount)
		}
	})
}
