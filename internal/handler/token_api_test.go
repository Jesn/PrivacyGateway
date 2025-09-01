package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

func setupTokenAPITest() (*TokenAPIHandler, *proxyconfig.ProxyConfig) {
	storage := proxyconfig.NewMemoryStorage(100)
	log := logger.New()
	handler := NewTokenAPIHandler(storage, "test-secret", log)

	// 创建测试配置
	config := &proxyconfig.ProxyConfig{
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)

	return handler, config
}

func TestTokenAPIHandler_HandleListTokens(t *testing.T) {
	handler, config := setupTokenAPITest()

	// 添加测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Test Token",
		Description: "Test token for API",
	}
	token, _, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	handler.storage.AddToken(config.ID, token)

	// 创建请求
	req := httptest.NewRequest("GET", "/config/proxy/"+config.ID+"/tokens", nil)
	req.Header.Set("X-Log-Secret", "test-secret")
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleTokenAPI(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response TokenListAPIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if len(response.Data.Tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(response.Data.Tokens))
	}

	if response.Data.Tokens[0].Name != "Test Token" {
		t.Errorf("Expected token name 'Test Token', got %s", response.Data.Tokens[0].Name)
	}

	// 验证敏感信息已被清理
	if response.Data.Tokens[0].TokenHash != "" {
		t.Error("Token hash should be empty in response")
	}
}

func TestTokenAPIHandler_HandleCreateToken(t *testing.T) {
	handler, config := setupTokenAPITest()

	// 创建请求体
	createReq := proxyconfig.TokenCreateRequest{
		Name:        "New Token",
		Description: "New test token",
	}
	reqBody, _ := json.Marshal(createReq)

	// 创建请求
	req := httptest.NewRequest("POST", "/config/proxy/"+config.ID+"/tokens", bytes.NewReader(reqBody))
	req.Header.Set("X-Log-Secret", "test-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleTokenAPI(w, req)

	// 验证响应
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response TokenAPIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Data.Name != "New Token" {
		t.Errorf("Expected token name 'New Token', got %s", response.Data.Name)
	}

	// 验证返回了明文令牌值
	if response.Data.Token == "" {
		t.Error("Expected token value to be returned")
	}

	// 验证令牌已保存到存储
	tokens, err := handler.storage.GetTokens(config.ID)
	if err != nil {
		t.Fatalf("Failed to get tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token in storage, got %d", len(tokens))
	}
}

func TestTokenAPIHandler_HandleUpdateToken(t *testing.T) {
	handler, config := setupTokenAPITest()

	// 添加测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Original Token",
		Description: "Original description",
	}
	token, _, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	handler.storage.AddToken(config.ID, token)

	// 创建更新请求体
	updateReq := proxyconfig.TokenUpdateRequest{
		Name:        "Updated Token",
		Description: "Updated description",
		Enabled:     boolPtr(false),
	}
	reqBody, _ := json.Marshal(updateReq)

	// 创建请求
	req := httptest.NewRequest("PUT", "/config/proxy/"+config.ID+"/tokens/"+token.ID, bytes.NewReader(reqBody))
	req.Header.Set("X-Log-Secret", "test-secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleTokenAPI(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response TokenAPIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Data.Name != "Updated Token" {
		t.Errorf("Expected token name 'Updated Token', got %s", response.Data.Name)
	}

	// 验证令牌已更新
	updatedToken, err := handler.storage.GetTokenByID(config.ID, token.ID)
	if err != nil {
		t.Fatalf("Failed to get updated token: %v", err)
	}
	if updatedToken.Name != "Updated Token" {
		t.Errorf("Expected updated name 'Updated Token', got %s", updatedToken.Name)
	}
	if updatedToken.Enabled {
		t.Error("Expected token to be disabled")
	}
}

func TestTokenAPIHandler_HandleDeleteToken(t *testing.T) {
	handler, config := setupTokenAPITest()

	// 添加测试令牌
	tokenReq := &proxyconfig.TokenCreateRequest{
		Name:        "Token to Delete",
		Description: "This token will be deleted",
	}
	token, _, err := proxyconfig.CreateAccessToken(tokenReq, "admin")
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	handler.storage.AddToken(config.ID, token)

	// 创建删除请求
	req := httptest.NewRequest("DELETE", "/config/proxy/"+config.ID+"/tokens/"+token.ID, nil)
	req.Header.Set("X-Log-Secret", "test-secret")
	w := httptest.NewRecorder()

	// 执行请求
	handler.HandleTokenAPI(w, req)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	// 验证令牌已删除
	_, err = handler.storage.GetTokenByID(config.ID, token.ID)
	if err != proxyconfig.ErrTokenNotFound {
		t.Error("Expected token to be deleted")
	}
}

func TestTokenAPIHandler_Authentication(t *testing.T) {
	handler, config := setupTokenAPITest()

	tests := []struct {
		name           string
		secret         string
		expectedStatus int
	}{
		{
			name:           "valid admin secret",
			secret:         "test-secret",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid admin secret",
			secret:         "wrong-secret",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no admin secret",
			secret:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/config/proxy/"+config.ID+"/tokens", nil)
			if tt.secret != "" {
				req.Header.Set("X-Log-Secret", tt.secret)
			}
			w := httptest.NewRecorder()

			handler.HandleTokenAPI(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestTokenAPIHandler_InvalidRequests(t *testing.T) {
	handler, config := setupTokenAPITest()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "invalid config ID",
			method:         "GET",
			path:           "/config/proxy/nonexistent/tokens",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid JSON in create",
			method:         "POST",
			path:           "/config/proxy/" + config.ID + "/tokens",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing token name in create",
			method:         "POST",
			path:           "/config/proxy/" + config.ID + "/tokens",
			body:           `{"description":"test"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "nonexistent token ID",
			method:         "GET",
			path:           "/config/proxy/" + config.ID + "/tokens/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody *strings.Reader
			if tt.body != "" {
				reqBody = strings.NewReader(tt.body)
			} else {
				reqBody = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			req.Header.Set("X-Log-Secret", "test-secret")
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()

			handler.HandleTokenAPI(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestTokenAPIHandler_PathExtraction(t *testing.T) {
	handler, _ := setupTokenAPITest()

	tests := []struct {
		name           string
		path           string
		expectedConfig string
		expectedToken  string
	}{
		{
			name:           "tokens list path",
			path:           "/config/proxy/config123/tokens",
			expectedConfig: "config123",
			expectedToken:  "",
		},
		{
			name:           "single token path",
			path:           "/config/proxy/config456/tokens/token789",
			expectedConfig: "config456",
			expectedToken:  "token789",
		},
		{
			name:           "invalid path",
			path:           "/invalid/path",
			expectedConfig: "",
			expectedToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configID := handler.extractConfigIDFromPath(tt.path)
			if configID != tt.expectedConfig {
				t.Errorf("Expected config ID %s, got %s", tt.expectedConfig, configID)
			}

			tokenID := handler.extractTokenIDFromPath(tt.path)
			if tokenID != tt.expectedToken {
				t.Errorf("Expected token ID %s, got %s", tt.expectedToken, tokenID)
			}
		})
	}
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}
