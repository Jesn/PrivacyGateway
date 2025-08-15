package logviewer

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/logger"
)

// TestLoginLoopIssue 测试登录循环问题
func TestLoginLoopIssue(t *testing.T) {
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

	// 场景1：用密钥A登录
	secret1 := "12345678"
	handler1, err := NewHandler(recorder, secret1, log)
	if err != nil {
		t.Fatalf("Failed to create handler with secret1: %v", err)
	}

	// 模拟用户登录
	req1 := httptest.NewRequest("POST", "/logs", strings.NewReader("secret=12345678"))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()

	handler1.ServeHTTP(w1, req1)

	// 检查是否设置了Cookie
	cookies := w1.Result().Cookies()
	var logSecretCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "log_secret" {
			logSecretCookie = cookie
			break
		}
	}

	if logSecretCookie == nil {
		t.Fatal("Expected log_secret cookie to be set after successful login")
	}

	t.Logf("场景1完成：用密钥 %s 登录成功，Cookie已设置", secret1)

	// 场景2：服务器重启，使用不同的密钥
	secret2 := "mylogviewer123"
	handler2, err := NewHandler(recorder, secret2, log)
	if err != nil {
		t.Fatalf("Failed to create handler with secret2: %v", err)
	}

	// 模拟浏览器使用旧的Cookie访问
	req2 := httptest.NewRequest("GET", "/logs", nil)
	req2.AddCookie(logSecretCookie) // 使用旧的Cookie
	w2 := httptest.NewRecorder()

	handler2.ServeHTTP(w2, req2)

	// 应该返回401和登录页面
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w2.Code)
	}

	body := w2.Body.String()
	if !strings.Contains(body, "访问密钥错误") {
		t.Error("Expected login page with error message")
	}

	t.Logf("场景2完成：密钥变更后访问返回401状态码")

	// 场景3：模拟JavaScript自动登录循环
	// 这里我们模拟多次请求来检测是否会出现循环
	loopCount := 0
	maxLoops := 5

	for loopCount < maxLoops {
		// 模拟JavaScript使用localStorage中的旧密钥自动提交
		formData := url.Values{}
		formData.Set("secret", secret1) // 使用旧密钥

		req3 := httptest.NewRequest("POST", "/logs", strings.NewReader(formData.Encode()))
		req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w3 := httptest.NewRecorder()

		handler2.ServeHTTP(w3, req3)

		// 检查响应
		if w3.Code == http.StatusUnauthorized {
			loopCount++
			t.Logf("循环 %d: 使用旧密钥 %s 认证失败，返回401", loopCount, secret1)

			// 检查响应体是否包含登录页面
			body := w3.Body.String()
			if strings.Contains(body, "访问密钥错误") {
				t.Logf("循环 %d: 返回登录页面，包含错误信息", loopCount)
			}
		} else {
			break
		}
	}

	if loopCount >= maxLoops {
		t.Errorf("检测到潜在的循环问题：连续 %d 次使用旧密钥都返回401", loopCount)
	}
}

// TestAuthFailureHandling 测试认证失败处理
func TestAuthFailureHandling(t *testing.T) {
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

	tests := []struct {
		name           string
		method         string
		secret         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Wrong secret via form",
			method:         "POST",
			secret:         "wrongsecret",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "访问密钥错误",
		},
		{
			name:           "Empty secret via form",
			method:         "POST",
			secret:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "请输入访问密钥",
		},
		{
			name:           "Wrong secret via cookie",
			method:         "GET",
			secret:         "wrongsecret",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "访问密钥错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.method == "POST" {
				formData := url.Values{}
				if tt.secret != "" {
					formData.Set("secret", tt.secret)
				}
				req = httptest.NewRequest("POST", "/logs", strings.NewReader(formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest("GET", "/logs", nil)
				if tt.secret != "" {
					// 模拟加密的cookie（这里简化处理）
					cookie := &http.Cookie{
						Name:  "log_secret",
						Value: tt.secret,
					}
					req.AddCookie(cookie)
				}
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.expectedError) {
				t.Errorf("Expected error message '%s' in response body", tt.expectedError)
			}
		})
	}
}

// TestCookieEncryption 测试Cookie加密解密
func TestCookieEncryption(t *testing.T) {
	auth := NewSecretAuthenticator("testsecret123")

	originalSecret := "mypassword"
	encrypted := auth.encryptSecret(originalSecret)
	decrypted := auth.decryptSecret(encrypted)

	if decrypted != originalSecret {
		t.Errorf("Cookie encryption/decryption failed. Original: %s, Decrypted: %s", originalSecret, decrypted)
	}

	// 测试无效的加密数据
	invalidDecrypted := auth.decryptSecret("invalid-base64-data")
	if invalidDecrypted != "" {
		t.Error("Expected empty string for invalid encrypted data")
	}
}
