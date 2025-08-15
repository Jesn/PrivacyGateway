package logviewer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/logger"
)

// TestImmediateLocalStorageClear 测试立即清除localStorage
func TestImmediateLocalStorageClear(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	// 创建使用正确密钥的handler
	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 模拟使用错误密钥的请求（模拟localStorage中的旧密钥）
	req := httptest.NewRequest("GET", "/logs", nil)
	// 添加错误的cookie来模拟localStorage中的旧密钥
	req.AddCookie(&http.Cookie{
		Name:  "log_secret",
		Value: "wrong_encrypted_secret",
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 检查响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// 检查是否设置了清除存储的响应头
	clearHeader := w.Header().Get("X-Clear-Auth-Storage")
	if clearHeader != "true" {
		t.Error("Expected X-Clear-Auth-Storage header to be set to 'true'")
	}

	body := w.Body.String()

	// 检查是否包含立即清除localStorage的逻辑
	expectedLogic := []string{
		"检测到认证失败，立即清除localStorage中的无效密钥",
		"localStorage.removeItem('log_viewer_secret')",
		"sessionStorage.removeItem('login_retry_count')",
		"localStorage已清除，跳过自动登录",
	}

	for _, logic := range expectedLogic {
		if !strings.Contains(body, logic) {
			t.Errorf("Expected immediate clear logic: %s", logic)
		}
	}

	t.Log("✓ 401响应包含立即清除localStorage的逻辑")
}

// TestLoginLoopFix 测试登录循环修复
func TestLoginLoopFix(t *testing.T) {
	// 创建测试用的recorder和logger
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	// 场景1：用密钥A登录成功
	secret1 := "12345678"
	handler1, err := NewHandler(recorder, secret1, log)
	if err != nil {
		t.Fatalf("Failed to create handler with secret1: %v", err)
	}

	// 模拟成功登录
	req1 := httptest.NewRequest("POST", "/logs", strings.NewReader("secret=12345678"))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()

	handler1.ServeHTTP(w1, req1)

	// 检查是否返回成功页面（包含清除重试计数的JavaScript）
	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200 for successful login, got %d", w1.Code)
	}

	body1 := w1.Body.String()
	if !strings.Contains(body1, "sessionStorage.removeItem('login_retry_count')") {
		t.Error("Expected login success page to contain retry count clearing JavaScript")
	}

	t.Log("✓ 场景1：成功登录返回清除重试计数的页面")

	// 场景2：服务器重启，使用不同的密钥
	secret2 := "mylogviewer123"
	handler2, err := NewHandler(recorder, secret2, log)
	if err != nil {
		t.Fatalf("Failed to create handler with secret2: %v", err)
	}

	// 模拟浏览器访问登录页面（GET请求）
	req2 := httptest.NewRequest("GET", "/logs", nil)
	w2 := httptest.NewRecorder()

	handler2.ServeHTTP(w2, req2)

	// 应该返回401和登录页面
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w2.Code)
	}

	body2 := w2.Body.String()

	// 检查是否包含错误检测和localStorage清除逻辑
	if !strings.Contains(body2, "const hasError = document.querySelector('.error')") {
		t.Error("Expected login page to contain error detection logic")
	}

	if !strings.Contains(body2, "localStorage.removeItem('log_viewer_secret')") {
		t.Error("Expected login page to contain localStorage clearing logic")
	}

	if !strings.Contains(body2, "login_retry_count") {
		t.Error("Expected login page to contain retry count logic")
	}

	t.Log("✓ 场景2：登录页面包含防循环逻辑")

	// 场景3：模拟使用错误密钥的POST请求（模拟JavaScript自动提交）
	req3 := httptest.NewRequest("POST", "/logs", strings.NewReader("secret=12345678"))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w3 := httptest.NewRecorder()

	handler2.ServeHTTP(w3, req3)

	// 应该返回401和包含错误信息的登录页面
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong secret, got %d", w3.Code)
	}

	body3 := w3.Body.String()
	if !strings.Contains(body3, "访问密钥错误") {
		t.Error("Expected error message in login page")
	}

	// 检查是否包含错误检测逻辑
	if !strings.Contains(body3, "const hasError = document.querySelector('.error')") {
		t.Error("Expected login page with error to contain error detection logic")
	}

	t.Log("✓ 场景3：错误密钥返回包含防循环逻辑的登录页面")
}

// TestRetryCountLogic 测试重试计数逻辑
func TestRetryCountLogic(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试GET请求返回的登录页面是否包含重试逻辑
	req := httptest.NewRequest("GET", "/logs", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()

	// 检查重试计数相关的JavaScript代码
	expectedJSSnippets := []string{
		"const retryCount = parseInt(sessionStorage.getItem('login_retry_count') || '0')",
		"const maxRetries = 3",
		"if (retryCount >= maxRetries)",
		"sessionStorage.setItem('login_retry_count', (retryCount + 1).toString())",
		"sessionStorage.removeItem('login_retry_count')",
	}

	for _, snippet := range expectedJSSnippets {
		if !strings.Contains(body, snippet) {
			t.Errorf("Expected login page to contain JavaScript snippet: %s", snippet)
		}
	}

	t.Log("✓ 登录页面包含完整的重试计数逻辑")
}

// TestErrorDetectionLogic 测试错误检测逻辑
func TestErrorDetectionLogic(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试使用错误密钥的POST请求
	req := httptest.NewRequest("POST", "/logs", strings.NewReader("secret=wrongsecret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	body := w.Body.String()

	// 检查是否包含错误信息
	if !strings.Contains(body, "访问密钥错误") {
		t.Error("Expected error message in response")
	}

	// 检查是否包含错误检测和清除localStorage的逻辑
	expectedLogic := []string{
		"const hasError = document.querySelector('.error')",
		"if (hasError || (isUnauthorized && localStorage.getItem('log_viewer_secret')))",
		"console.log('检测到认证失败，立即清除localStorage中的无效密钥')",
		"localStorage.removeItem('log_viewer_secret')",
	}

	for _, logic := range expectedLogic {
		if !strings.Contains(body, logic) {
			t.Errorf("Expected error detection logic: %s", logic)
		}
	}

	t.Log("✓ 错误页面包含正确的错误检测和清除逻辑")
}

// TestSuccessfulLoginFlow 测试成功登录流程
func TestSuccessfulLoginFlow(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试成功登录的POST请求
	req := httptest.NewRequest("POST", "/logs", strings.NewReader("secret=correctsecret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for successful login, got %d", w.Code)
	}

	body := w.Body.String()

	// 检查是否包含清除重试计数的逻辑
	if !strings.Contains(body, "sessionStorage.removeItem('login_retry_count')") {
		t.Error("Expected successful login page to clear retry count")
	}

	if !strings.Contains(body, "window.location.href = '/logs'") {
		t.Error("Expected successful login page to redirect to logs")
	}

	t.Log("✓ 成功登录返回正确的重定向页面")
}

// TestLogoutWithChangedSecret 测试密钥变更后的logout功能
func TestLogoutWithChangedSecret(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	// 场景：密钥已经变更，但用户仍然尝试logout
	handler, err := NewHandler(recorder, "newsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 模拟用户访问logout路径（可能带有旧的cookie）
	req := httptest.NewRequest("GET", "/logs/logout", nil)
	// 添加旧的cookie
	req.AddCookie(&http.Cookie{
		Name:  "log_secret",
		Value: "old_encrypted_secret",
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// logout应该成功，不应该返回401
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for logout, got %d", w.Code)
	}

	body := w.Body.String()

	// 检查是否包含清除localStorage的逻辑
	expectedLogic := []string{
		"localStorage.removeItem('log_viewer_secret')",
		"sessionStorage.removeItem('login_retry_count')",
		"已退出登录，清除本地存储",
		"window.location.href = '/logs'",
	}

	for _, logic := range expectedLogic {
		if !strings.Contains(body, logic) {
			t.Errorf("Expected logout page to contain: %s", logic)
		}
	}

	// 检查是否设置了清除cookie
	cookies := w.Result().Cookies()
	var logSecretCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "log_secret" {
			logSecretCookie = cookie
			break
		}
	}

	if logSecretCookie == nil {
		t.Error("Expected log_secret cookie to be set for clearing")
	} else if logSecretCookie.MaxAge != -1 {
		t.Errorf("Expected cookie MaxAge to be -1 for immediate expiry, got %d", logSecretCookie.MaxAge)
	}

	t.Log("✅ logout在密钥变更后仍能正常工作")
}

// TestLogoutAPI 测试API方式的logout
func TestLogoutAPI(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "testsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试API方式的logout
	req := httptest.NewRequest("POST", "/logs/logout", nil)
	req.Header.Set("Accept", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for API logout, got %d", w.Code)
	}

	// 检查JSON响应
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Error("Expected JSON content type for API logout")
	}

	body := w.Body.String()
	if !strings.Contains(body, "已退出登录") {
		t.Error("Expected logout success message in JSON response")
	}

	t.Log("✅ API logout正常工作")
}

// TestAutomaticCleanupOn401 测试401时的自动清理机制
func TestAutomaticCleanupOn401(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试带有无效cookie的GET请求
	req := httptest.NewRequest("GET", "/logs", nil)
	req.AddCookie(&http.Cookie{
		Name:  "log_secret",
		Value: "invalid_old_cookie",
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 验证401响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// 验证服务器自动清除了cookie
	cookies := w.Result().Cookies()
	var logSecretCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "log_secret" {
			logSecretCookie = cookie
			break
		}
	}

	if logSecretCookie == nil {
		t.Error("Expected server to set clearing cookie")
	} else {
		if logSecretCookie.MaxAge != -1 {
			t.Errorf("Expected cookie MaxAge to be -1 for clearing, got %d", logSecretCookie.MaxAge)
		}
		if logSecretCookie.Value != "" {
			t.Errorf("Expected cookie value to be empty for clearing, got %s", logSecretCookie.Value)
		}
	}

	// 验证响应头包含清除指示
	clearHeader := w.Header().Get("X-Clear-Auth-Storage")
	if clearHeader != "true" {
		t.Error("Expected X-Clear-Auth-Storage header to be 'true'")
	}

	// 验证页面包含清除meta标签
	body := w.Body.String()
	if !strings.Contains(body, `<meta name="clear-auth" content="true">`) {
		t.Error("Expected page to contain clear-auth meta tag")
	}

	// 验证页面包含立即清除的JavaScript
	expectedJS := []string{
		"检测到401认证失败页面，立即清除所有认证信息",
		"localStorage.removeItem('log_viewer_secret')",
		"sessionStorage.removeItem('login_retry_count')",
		"认证信息清除完成",
	}

	for _, js := range expectedJS {
		if !strings.Contains(body, js) {
			t.Errorf("Expected page to contain JavaScript: %s", js)
		}
	}

	t.Log("✅ 401时自动清理机制正常工作")
}

// TestAPIAutomaticCleanupOn401 测试API请求401时的自动清理
func TestAPIAutomaticCleanupOn401(t *testing.T) {
	cfg := &config.Config{
		LogMaxEntries:     100,
		LogMaxMemoryMB:    10,
		LogRetentionHours: 24,
		LogMaxBodySize:    1024,
	}
	log := logger.New()
	recorder, err := accesslog.NewRecorder(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Close()

	handler, err := NewHandler(recorder, "correctsecret", log)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// 测试API请求带有无效cookie
	req := httptest.NewRequest("GET", "/logs/api", nil)
	req.Header.Set("Accept", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "log_secret",
		Value: "invalid_old_cookie",
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 验证401响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// 验证服务器自动清除了cookie
	cookies := w.Result().Cookies()
	var logSecretCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "log_secret" {
			logSecretCookie = cookie
			break
		}
	}

	if logSecretCookie == nil {
		t.Error("Expected server to set clearing cookie for API request")
	} else if logSecretCookie.MaxAge != -1 {
		t.Errorf("Expected cookie MaxAge to be -1 for clearing, got %d", logSecretCookie.MaxAge)
	}

	// 验证响应头包含清除指示
	clearHeader := w.Header().Get("X-Clear-Auth-Storage")
	if clearHeader != "true" {
		t.Error("Expected X-Clear-Auth-Storage header to be 'true' for API request")
	}

	// 验证JSON响应包含清除指示
	body := w.Body.String()
	if !strings.Contains(body, `"clear_storage":true`) {
		t.Error("Expected JSON response to contain clear_storage flag")
	}

	t.Log("✅ API请求401时自动清理机制正常工作")
}
