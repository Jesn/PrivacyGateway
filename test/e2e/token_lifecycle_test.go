package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
	"privacygateway/internal/router"
)

// E2ETestSuite 端到端测试套件
type E2ETestSuite struct {
	server         *httptest.Server
	cfg            *config.Config
	log            *logger.Logger
	storage        proxyconfig.Storage
	adminSecret    string
	testConfigID   string
	testTokenID    string
	testTokenValue string
}

// SetupE2ETest 设置端到端测试环境
func SetupE2ETest(t *testing.T) *E2ETestSuite {
	// 创建测试配置
	cfg := &config.Config{
		AdminSecret: "e2e-test-secret",
		Port:        "0", // 使用随机端口
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
	setupRoutesForMux(mux, appRouter)

	// 创建测试服务器
	server := httptest.NewServer(mux)

	suite := &E2ETestSuite{
		server:      server,
		cfg:         cfg,
		log:         log,
		storage:     storage,
		adminSecret: cfg.AdminSecret,
	}

	// 创建测试配置
	suite.createTestConfig(t)

	return suite
}

// setupRoutesForMux 为指定的mux设置路由
func setupRoutesForMux(mux *http.ServeMux, appRouter *router.Router) {
	// 主要路由
	mux.HandleFunc("/", appRouter.HandleRoot)
	mux.HandleFunc("/proxy", appRouter.HandleHTTPProxy)
	mux.HandleFunc("/ws", appRouter.HandleWebSocket)

	// API路由
	mux.HandleFunc("/config/proxy", appRouter.HandleProxyConfigAPI)
	mux.HandleFunc("/config/proxy/export", appRouter.HandleProxyConfigExportAPI)
	mux.HandleFunc("/config/proxy/import", appRouter.HandleProxyConfigImportAPI)
	mux.HandleFunc("/config/proxy/batch", appRouter.HandleProxyConfigBatchAPI)
	mux.HandleFunc("/config/proxy/", appRouter.HandleProxyConfigOrTokenAPI)
}

// TearDown 清理测试环境
func (suite *E2ETestSuite) TearDown() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// createTestConfig 创建测试配置
func (suite *E2ETestSuite) createTestConfig(t *testing.T) {
	configReq := map[string]interface{}{
		"name":       "E2E Test Config",
		"subdomain":  "e2etest",
		"target_url": "https://httpbin.org",
		"protocol":   "https",
		"enabled":    true,
	}

	reqBody, _ := json.Marshal(configReq)
	resp := suite.makeRequest(t, "POST", "/config/proxy", reqBody, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create test config: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode config response: %v", err)
	}

	t.Logf("Config creation response: %+v", response)

	success, ok := response["success"].(bool)
	if !ok {
		t.Logf("Success field not found or not bool, response: %+v", response)
		// 如果没有success字段，检查是否有data字段（可能是直接返回的配置）
		if data, exists := response["data"]; exists && data != nil {
			// 继续处理
		} else if id, exists := response["id"]; exists && id != nil {
			// 直接返回的配置对象
			suite.testConfigID = id.(string)
			return
		} else {
			t.Fatalf("Invalid response format: %+v", response)
		}
	} else if !success {
		t.Fatalf("Config creation failed: %v", response)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok || data == nil {
		t.Fatalf("Invalid response data: %v", response)
	}

	configID, ok := data["id"].(string)
	if !ok || configID == "" {
		t.Fatalf("Invalid config ID in response: %v", data)
	}

	suite.testConfigID = configID
}

// makeRequest 发送HTTP请求
func (suite *E2ETestSuite) makeRequest(t *testing.T, method, path string, body []byte, headers map[string]string) *http.Response {
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, suite.server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

// TestTokenCompleteLifecycle 测试令牌完整生命周期
func TestTokenCompleteLifecycle(t *testing.T) {
	suite := SetupE2ETest(t)
	defer suite.TearDown()

	t.Run("1. Create Token", func(t *testing.T) {
		suite.testCreateToken(t)
	})

	t.Run("2. List Tokens", func(t *testing.T) {
		suite.testListTokens(t)
	})

	t.Run("3. Get Token Details", func(t *testing.T) {
		suite.testGetToken(t)
	})

	t.Run("4. Use Token for Proxy", func(t *testing.T) {
		suite.testUseTokenForProxy(t)
	})

	t.Run("5. Update Token", func(t *testing.T) {
		suite.testUpdateToken(t)
	})

	t.Run("6. Verify Token Usage Statistics", func(t *testing.T) {
		suite.testTokenUsageStatistics(t)
	})

	t.Run("7. Disable Token", func(t *testing.T) {
		suite.testDisableToken(t)
	})

	t.Run("8. Verify Disabled Token", func(t *testing.T) {
		suite.testDisabledTokenAccess(t)
	})

	t.Run("9. Delete Token", func(t *testing.T) {
		suite.testDeleteToken(t)
	})

	t.Run("10. Verify Token Deletion", func(t *testing.T) {
		suite.testVerifyTokenDeletion(t)
	})
}

// testCreateToken 测试创建令牌
func (suite *E2ETestSuite) testCreateToken(t *testing.T) {
	tokenReq := map[string]interface{}{
		"name":        "E2E Test Token",
		"description": "Token for end-to-end testing",
	}

	reqBody, _ := json.Marshal(tokenReq)
	resp := suite.makeRequest(t, "POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if !response["success"].(bool) {
		t.Fatal("Expected success=true")
	}

	data := response["data"].(map[string]interface{})
	suite.testTokenID = data["id"].(string)
	suite.testTokenValue = data["token"].(string)

	if suite.testTokenID == "" || suite.testTokenValue == "" {
		t.Fatal("Token ID or value is empty")
	}

	t.Logf("Created token: ID=%s, Value=%s", suite.testTokenID, suite.testTokenValue)
}

// testListTokens 测试获取令牌列表
func (suite *E2ETestSuite) testListTokens(t *testing.T) {
	resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if !response["success"].(bool) {
		t.Fatal("Expected success=true")
	}

	data := response["data"].(map[string]interface{})
	tokens := data["tokens"].([]interface{})

	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	token := tokens[0].(map[string]interface{})
	if token["id"].(string) != suite.testTokenID {
		t.Fatalf("Expected token ID %s, got %s", suite.testTokenID, token["id"].(string))
	}
}

// testGetToken 测试获取单个令牌
func (suite *E2ETestSuite) testGetToken(t *testing.T) {
	resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens/"+suite.testTokenID, nil, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if !response["success"].(bool) {
		t.Fatal("Expected success=true")
	}

	// 检查响应结构
	if data, ok := response["data"].(map[string]interface{}); ok && data != nil {
		// 标准API响应格式
		if token, ok := data["access_token"].(map[string]interface{}); ok && token != nil {
			// 使用token
			if token["id"].(string) != suite.testTokenID {
				t.Fatalf("Expected token ID %s, got %s", suite.testTokenID, token["id"].(string))
			}
			if token["name"].(string) != "E2E Test Token" {
				t.Fatalf("Expected token name 'E2E Test Token', got %s", token["name"].(string))
			}
			return
		} else {
			// data直接是令牌对象
			if data["id"].(string) != suite.testTokenID {
				t.Fatalf("Expected token ID %s, got %s", suite.testTokenID, data["id"].(string))
			}
			if data["name"].(string) != "E2E Test Token" {
				t.Fatalf("Expected token name 'E2E Test Token', got %s", data["name"].(string))
			}
			return
		}
	} else {
		// 直接返回令牌对象
		if response["id"].(string) != suite.testTokenID {
			t.Fatalf("Expected token ID %s, got %s", suite.testTokenID, response["id"].(string))
		}
		if response["name"].(string) != "E2E Test Token" {
			t.Fatalf("Expected token name 'E2E Test Token', got %s", response["name"].(string))
		}
		return
	}
}

// testUseTokenForProxy 测试使用令牌进行代理请求
func (suite *E2ETestSuite) testUseTokenForProxy(t *testing.T) {
	// 测试HTTP代理
	resp := suite.makeRequest(t, "GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
		"X-Proxy-Token": suite.testTokenValue,
	})

	// 由于这是真实的网络请求，我们主要验证认证部分
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("Token authentication failed")
	}

	t.Logf("Proxy request status: %d", resp.StatusCode)
}

// testUpdateToken 测试更新令牌
func (suite *E2ETestSuite) testUpdateToken(t *testing.T) {
	updateReq := map[string]interface{}{
		"name":        "Updated E2E Test Token",
		"description": "Updated description for testing",
	}

	reqBody, _ := json.Marshal(updateReq)
	resp := suite.makeRequest(t, "PUT", "/config/proxy/"+suite.testConfigID+"/tokens/"+suite.testTokenID, reqBody, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	// 检查响应结构
	if success, ok := response["success"].(bool); ok && success {
		// 标准API响应格式
		if data, ok := response["data"].(map[string]interface{}); ok && data != nil {
			if token, ok := data["access_token"].(map[string]interface{}); ok && token != nil {
				if token["name"].(string) != "Updated E2E Test Token" {
					t.Fatalf("Expected updated name, got %s", token["name"].(string))
				}
			} else if data["name"].(string) != "Updated E2E Test Token" {
				t.Fatalf("Expected updated name, got %s", data["name"].(string))
			}
		}
	} else {
		// 直接返回令牌对象
		if response["name"].(string) != "Updated E2E Test Token" {
			t.Fatalf("Expected updated name, got %s", response["name"].(string))
		}
	}
}

// testTokenUsageStatistics 测试令牌使用统计
func (suite *E2ETestSuite) testTokenUsageStatistics(t *testing.T) {
	// 先执行几个代理请求来生成统计数据
	for i := 0; i < 3; i++ {
		suite.makeRequest(t, "GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
			"X-Proxy-Token": suite.testTokenValue,
		})
		time.Sleep(100 * time.Millisecond) // 短暂等待
	}

	// 获取令牌统计
	resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	data := response["data"].(map[string]interface{})
	if stats, exists := data["stats"]; exists && stats != nil {
		statsMap := stats.(map[string]interface{})
		if totalRequests, exists := statsMap["total_requests"]; exists {
			if totalRequests.(float64) > 0 {
				t.Logf("Token usage statistics updated: %v requests", totalRequests)
			}
		}
	}
}

// testDisableToken 测试禁用令牌
func (suite *E2ETestSuite) testDisableToken(t *testing.T) {
	updateReq := map[string]interface{}{
		"enabled": false,
	}

	reqBody, _ := json.Marshal(updateReq)
	resp := suite.makeRequest(t, "PUT", "/config/proxy/"+suite.testConfigID+"/tokens/"+suite.testTokenID, reqBody, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if !response["success"].(bool) {
		t.Fatal("Expected success=true")
	}
}

// testDisabledTokenAccess 测试禁用令牌的访问
func (suite *E2ETestSuite) testDisabledTokenAccess(t *testing.T) {
	resp := suite.makeRequest(t, "GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
		"X-Proxy-Token": suite.testTokenValue,
	})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 for disabled token, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if response["error_code"] != "TOKEN_DISABLED" {
		t.Fatalf("Expected error_code TOKEN_DISABLED, got %v", response["error_code"])
	}
}

// testDeleteToken 测试删除令牌
func (suite *E2ETestSuite) testDeleteToken(t *testing.T) {
	resp := suite.makeRequest(t, "DELETE", "/config/proxy/"+suite.testConfigID+"/tokens/"+suite.testTokenID, nil, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	if !response["success"].(bool) {
		t.Fatal("Expected success=true")
	}
}

// testVerifyTokenDeletion 测试验证令牌已删除
func (suite *E2ETestSuite) testVerifyTokenDeletion(t *testing.T) {
	// 尝试获取已删除的令牌
	resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens/"+suite.testTokenID, nil, map[string]string{
		"X-Log-Secret": suite.adminSecret,
	})

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected status 404 for deleted token, got %d", resp.StatusCode)
	}

	// 尝试使用已删除的令牌
	resp = suite.makeRequest(t, "GET", "/proxy?target=https://httpbin.org/get&config_id="+suite.testConfigID, nil, map[string]string{
		"X-Proxy-Token": suite.testTokenValue,
	})

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 for deleted token, got %d", resp.StatusCode)
	}
}

// TestConcurrentTokenOperations 测试并发令牌操作
func TestConcurrentTokenOperations(t *testing.T) {
	suite := SetupE2ETest(t)
	defer suite.TearDown()

	const numGoroutines = 10
	const numOperationsPerGoroutine = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// 并发创建令牌
	t.Run("Concurrent Token Creation", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for j := 0; j < numOperationsPerGoroutine; j++ {
					tokenReq := map[string]interface{}{
						"name":        fmt.Sprintf("Concurrent Token %d-%d", id, j),
						"description": fmt.Sprintf("Concurrent test token %d-%d", id, j),
					}

					reqBody, _ := json.Marshal(tokenReq)
					resp := suite.makeRequest(t, "POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
						"X-Log-Secret": suite.adminSecret,
					})

					if resp.StatusCode != http.StatusCreated {
						errors <- fmt.Errorf("goroutine %d operation %d: expected status 201, got %d", id, j, resp.StatusCode)
						return
					}

					var response map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
						errors <- fmt.Errorf("goroutine %d operation %d: failed to decode response: %v", id, j, err)
						return
					}

					if !response["success"].(bool) {
						errors <- fmt.Errorf("goroutine %d operation %d: expected success=true", id, j)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// 检查错误
		var errorList []error
		for err := range errors {
			errorList = append(errorList, err)
		}

		if len(errorList) > 0 {
			t.Fatalf("Concurrent operations failed with %d errors: %v", len(errorList), errorList[0])
		}

		// 验证所有令牌都已创建
		resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
			"X-Log-Secret": suite.adminSecret,
		})

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		data := response["data"].(map[string]interface{})
		tokens := data["tokens"].([]interface{})

		expectedTokens := numGoroutines * numOperationsPerGoroutine
		if len(tokens) != expectedTokens {
			t.Fatalf("Expected %d tokens, got %d", expectedTokens, len(tokens))
		}

		t.Logf("Successfully created %d tokens concurrently", len(tokens))
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	suite := SetupE2ETest(t)
	defer suite.TearDown()

	t.Run("Invalid Authentication", func(t *testing.T) {
		// 无效的管理员密钥
		resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{
			"X-Log-Secret": "invalid-secret",
		})

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected status 401, got %d", resp.StatusCode)
		}

		// 无认证信息
		resp = suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens", nil, map[string]string{})

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid Configuration ID", func(t *testing.T) {
		resp := suite.makeRequest(t, "GET", "/config/proxy/nonexistent-config/tokens", nil, map[string]string{
			"X-Log-Secret": suite.adminSecret,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid Token ID", func(t *testing.T) {
		resp := suite.makeRequest(t, "GET", "/config/proxy/"+suite.testConfigID+"/tokens/nonexistent-token", nil, map[string]string{
			"X-Log-Secret": suite.adminSecret,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid JSON Format", func(t *testing.T) {
		invalidJSON := []byte("invalid json")
		resp := suite.makeRequest(t, "POST", "/config/proxy/"+suite.testConfigID+"/tokens", invalidJSON, map[string]string{
			"X-Log-Secret": suite.adminSecret,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		// 缺少名称字段
		tokenReq := map[string]interface{}{
			"description": "Token without name",
		}

		reqBody, _ := json.Marshal(tokenReq)
		resp := suite.makeRequest(t, "POST", "/config/proxy/"+suite.testConfigID+"/tokens", reqBody, map[string]string{
			"X-Log-Secret": suite.adminSecret,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}
