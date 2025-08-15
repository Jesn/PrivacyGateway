package logviewer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Secret string // 查看密钥
}

// AuthResult 认证结果
type AuthResult struct {
	Authenticated bool   // 是否认证成功
	Error         string // 错误信息
	Method        string // 认证方式
}

// Authenticator 认证器接口
type Authenticator interface {
	// Authenticate 验证请求是否有权限访问
	Authenticate(r *http.Request) *AuthResult

	// RequireAuth 中间件，要求认证
	RequireAuth(next http.HandlerFunc) http.HandlerFunc

	// IsEnabled 检查认证是否启用
	IsEnabled() bool
}

// SecretAuthenticator 基于密钥的认证器
type SecretAuthenticator struct {
	config *AuthConfig
}

// NewSecretAuthenticator 创建新的密钥认证器
func NewSecretAuthenticator(secret string) *SecretAuthenticator {
	return &SecretAuthenticator{
		config: &AuthConfig{
			Secret: secret,
		},
	}
}

// Authenticate 验证请求
func (sa *SecretAuthenticator) Authenticate(r *http.Request) *AuthResult {
	if sa.config.Secret == "" {
		return &AuthResult{
			Authenticated: false,
			Error:         "系统未配置访问密钥",
			Method:        "secret",
		}
	}

	// 从多个来源获取密钥
	secret := sa.extractSecret(r)
	if secret == "" {
		return &AuthResult{
			Authenticated: false,
			Error:         "请输入访问密钥",
			Method:        "secret",
		}
	}

	// 使用常量时间比较防止时序攻击
	if subtle.ConstantTimeCompare([]byte(secret), []byte(sa.config.Secret)) == 1 {
		return &AuthResult{
			Authenticated: true,
			Error:         "",
			Method:        "secret",
		}
	}

	return &AuthResult{
		Authenticated: false,
		Error:         "访问密钥错误，请重新输入",
		Method:        "secret",
	}
}

// RequireAuth 认证中间件
func (sa *SecretAuthenticator) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := sa.Authenticate(r)

		if !result.Authenticated {
			sa.handleAuthFailure(w, r, result)
			return
		}

		// 认证成功，设置安全Cookie（如果是通过表单登录）
		if r.FormValue("secret") != "" {
			sa.SetSecureCookie(w, r.FormValue("secret"))
		}

		// 认证成功，继续处理
		next(w, r)
	}
}

// extractSecret 从请求中提取密钥
func (sa *SecretAuthenticator) extractSecret(r *http.Request) string {
	// 1. 优先从自定义头获取 (推荐方式)
	if secret := r.Header.Get("X-Log-Secret"); secret != "" {
		return secret
	}

	// 2. 从 Authorization 头获取 (Bearer token)
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// 3. 从 Cookie 获取 (加密存储)
	if cookie, err := r.Cookie("log_secret"); err == nil {
		// 这里可以添加解密逻辑
		return sa.decryptSecret(cookie.Value)
	}

	// 4. 从表单数据获取 (仅用于初始登录)
	if secret := r.FormValue("secret"); secret != "" {
		return secret
	}

	// 5. 从查询参数获取 (向后兼容，不推荐)
	if secret := r.URL.Query().Get("secret"); secret != "" {
		return secret
	}

	return ""
}

// handleAuthFailure 处理认证失败
func (sa *SecretAuthenticator) handleAuthFailure(w http.ResponseWriter, r *http.Request, result *AuthResult) {
	// 检查是否是 API 请求
	if sa.isAPIRequest(r) {
		sa.handleAPIAuthFailure(w, result)
		return
	}

	// 处理 Web 页面认证失败
	sa.handleWebAuthFailure(w, r, result)
}

// isAPIRequest 判断是否是 API 请求
func (sa *SecretAuthenticator) isAPIRequest(r *http.Request) bool {
	// 检查 Accept 头
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	// 检查路径
	if strings.HasSuffix(r.URL.Path, "/api") || strings.Contains(r.URL.Path, "/api/") {
		return true
	}

	// 检查查询参数
	if r.URL.Query().Get("format") == "json" {
		return true
	}

	return false
}

// handleAPIAuthFailure 处理 API 认证失败
func (sa *SecretAuthenticator) handleAPIAuthFailure(w http.ResponseWriter, result *AuthResult) {
	// 立即清除服务器端的Cookie
	sa.clearServerCookie(w)

	w.Header().Set("Content-Type", "application/json")
	// 添加清除存储的指示头
	w.Header().Set("X-Clear-Auth-Storage", "true")
	w.WriteHeader(http.StatusUnauthorized)

	// 简单的 JSON 响应
	w.Write([]byte(`{"error":"` + result.Error + `","message":"Authentication required","method":"` + result.Method + `","clear_storage":true}`))
}

// handleWebAuthFailure 处理 Web 页面认证失败
func (sa *SecretAuthenticator) handleWebAuthFailure(w http.ResponseWriter, r *http.Request, result *AuthResult) {
	// 立即清除服务器端的Cookie
	sa.clearServerCookie(w)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 添加自定义头，指示前端清除localStorage
	w.Header().Set("X-Clear-Auth-Storage", "true")
	w.WriteHeader(http.StatusUnauthorized)

	// 美观的中文认证页面
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <title>访问日志 - 身份验证</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="clear-auth" content="true">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 15px 35px rgba(0, 0, 0, 0.1);
            width: 100%;
            max-width: 400px;
            text-align: center;
        }
        .logo {
            width: 60px;
            height: 60px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
            margin-bottom: 8px;
            font-size: 24px;
            font-weight: 600;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
            text-align: left;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="password"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e1e5e9;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s ease;
            background: #f8f9fa;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
            background: white;
        }
        .login-btn {
            width: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 12px 20px;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 500;
            cursor: pointer;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .login-btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.3);
        }
        .login-btn:active {
            transform: translateY(0);
        }
        .error {
            background: #fee;
            color: #c53030;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
            border-left: 4px solid #c53030;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e1e5e9;
            color: #666;
            font-size: 12px;
        }
        @media (max-width: 480px) {
            .login-container {
                margin: 20px;
                padding: 30px 20px;
            }
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">🔒</div>
        <h1>访问日志</h1>
        <p class="subtitle">请输入访问密钥以查看系统日志</p>

        ` + (func() string {
		if result.Error != "" {
			return `<div class="error">` + result.Error + `</div>`
		}
		return ""
	})() + `

        <form method="POST" onsubmit="handleLogin(event)">
            <div class="form-group">
                <label for="secret">访问密钥</label>
                <input type="password" id="secret" name="secret" placeholder="请输入访问密钥" required>
            </div>
            <button type="submit" class="login-btn">进入日志系统</button>
        </form>

        <div class="footer">
            <p>Privacy Gateway 日志查看系统</p>
        </div>
    </div>

    <script>
        function handleLogin(event) {
            event.preventDefault();
            const secret = document.getElementById('secret').value;

            if (!secret.trim()) {
                alert('请输入访问密钥');
                return;
            }

            // 验证密钥格式
            if (!validateSecret(secret)) {
                alert('密钥格式错误：密钥长度必须在8-256个字符之间');
                return;
            }

            // 保存密钥到localStorage
            const encoded = btoa(secret + ':' + Date.now());
            localStorage.setItem('log_viewer_secret', encoded);

            // 清除重试计数，因为这是用户手动输入的新密钥
            sessionStorage.removeItem('login_retry_count');

            // 使用POST方式提交到服务器，让服务器设置Cookie
            const form = document.createElement('form');
            form.method = 'POST';
            form.action = '/logs';

            const secretInput = document.createElement('input');
            secretInput.type = 'hidden';
            secretInput.name = 'secret';
            secretInput.value = secret;

            form.appendChild(secretInput);
            document.body.appendChild(form);
            form.submit();
        }

        // 验证密钥格式
        function validateSecret(secret) {
            if (!secret || secret.length < 8) {
                return false;
            }
            if (secret.length > 256) {
                return false;
            }
            return true;
        }

        // 立即检查并清除无效的认证信息（在页面加载时立即执行）
        (function() {
            // 立即清除所有认证信息，因为这是401页面
            // 清除localStorage
            if (localStorage.getItem('log_viewer_secret')) {
                localStorage.removeItem('log_viewer_secret');
            }

            // 清除sessionStorage
            if (sessionStorage.getItem('login_retry_count')) {
                sessionStorage.removeItem('login_retry_count');
            }

            // 清除URL参数，避免循环
            if (window.location.search) {
                window.history.replaceState({}, '', window.location.pathname);
            }
        })();

        // 页面加载时的初始化
        document.addEventListener('DOMContentLoaded', function() {
            // 聚焦到密钥输入框
            document.getElementById('secret').focus();

            // 检查URL中是否有密钥参数，如果有则清除（避免循环）
            const urlParams = new URLSearchParams(window.location.search);
            const urlSecret = urlParams.get('secret');
            if (urlSecret) {
                // 清除URL参数，避免循环
                window.history.replaceState({}, '', window.location.pathname);

                // 如果密钥无效，显示错误并清除localStorage
                if (!validateSecret(urlSecret)) {
                    localStorage.removeItem('log_viewer_secret');
                    alert('密钥格式错误：密钥长度必须在8-256个字符之间');
                }
                return;
            }

            // 检查localStorage是否已被清除（说明认证失败）
            const savedSecret = localStorage.getItem('log_viewer_secret');
            if (!savedSecret) {
                // localStorage已被清除，不进行自动登录
                return;
            }

            // 检查是否有重试计数，防止无限循环
            const retryCount = parseInt(sessionStorage.getItem('login_retry_count') || '0');
            const maxRetries = 3;

            if (retryCount >= maxRetries) {
                sessionStorage.removeItem('login_retry_count');
                localStorage.removeItem('log_viewer_secret');
                return;
            }

            // 检查是否有错误信息，如果有则不进行自动登录
            const hasError = document.querySelector('.error');
            if (hasError) {
                return;
            }

            // 只有在没有URL参数、没有错误、localStorage有效的情况下，才进行自动登录
            try {
                const decoded = atob(savedSecret);
                const secret = decoded.split(':')[0];
                if (secret && validateSecret(secret)) {
                    // 增加重试计数
                    sessionStorage.setItem('login_retry_count', (retryCount + 1).toString());

                    // 使用POST方式提交，避免URL参数
                    const form = document.createElement('form');
                    form.method = 'POST';
                    form.action = '/logs';

                    const secretInput = document.createElement('input');
                    secretInput.type = 'hidden';
                    secretInput.name = 'secret';
                    secretInput.value = secret;

                    form.appendChild(secretInput);
                    document.body.appendChild(form);
                    form.submit();
                    return;
                } else {
                    // 清除无效的密钥
                    localStorage.removeItem('log_viewer_secret');
                    sessionStorage.removeItem('login_retry_count');
                }
            } catch (e) {
                // 如果解码失败，清除无效的密钥
                localStorage.removeItem('log_viewer_secret');
                sessionStorage.removeItem('login_retry_count');
            }
        });

        // 支持回车键提交
        document.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                handleLogin(e);
            }
        });
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// IsEnabled 检查认证是否启用
func (sa *SecretAuthenticator) IsEnabled() bool {
	return sa.config.Secret != ""
}

// GetConfig 获取认证配置（不包含敏感信息）
func (sa *SecretAuthenticator) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": sa.IsEnabled(),
		"method":  "secret",
	}
}

// ValidateSecret 验证密钥格式
func ValidateSecret(secret string) error {
	if secret == "" {
		return nil // 空密钥表示不启用认证
	}

	if len(secret) < 8 {
		return &AuthError{
			Code:    "WEAK_SECRET",
			Message: "访问密钥长度不足，至少需要8个字符",
		}
	}

	if len(secret) > 256 {
		return &AuthError{
			Code:    "SECRET_TOO_LONG",
			Message: "访问密钥过长，不能超过256个字符",
		}
	}

	return nil
}

// AuthError 认证错误
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *AuthError) Error() string {
	return e.Message
}

// encryptSecret 加密密钥用于存储在Cookie中
func (sa *SecretAuthenticator) encryptSecret(plaintext string) string {
	// 使用配置的密钥作为加密密钥的基础
	key := sha256.Sum256([]byte(sa.config.Secret + "cookie-encryption"))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return plaintext // 加密失败时返回原文
	}

	// 生成随机IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return plaintext
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decryptSecret 解密Cookie中的密钥
func (sa *SecretAuthenticator) decryptSecret(ciphertext string) string {
	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "" // 解密失败返回空
	}

	key := sha256.Sum256([]byte(sa.config.Secret + "cookie-encryption"))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return ""
	}

	if len(data) < aes.BlockSize {
		return ""
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return string(data)
}

// SetSecureCookie 设置安全的Cookie
func (sa *SecretAuthenticator) SetSecureCookie(w http.ResponseWriter, secret string) {
	encryptedSecret := sa.encryptSecret(secret)

	cookie := &http.Cookie{
		Name:     "log_secret",
		Value:    encryptedSecret,
		Path:     "/logs",
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true (HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24小时
	}

	http.SetCookie(w, cookie)
}

// clearServerCookie 清除服务器端Cookie
func (sa *SecretAuthenticator) clearServerCookie(w http.ResponseWriter) {
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
}

// CreateAuthenticator 创建认证器的工厂函数
func CreateAuthenticator(secret string) (Authenticator, error) {
	if err := ValidateSecret(secret); err != nil {
		return nil, err
	}

	return NewSecretAuthenticator(secret), nil
}
