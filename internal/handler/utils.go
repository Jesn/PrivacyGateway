package handler

import (
	"net/http"
	"strings"
)

// IsSensitiveHeader 检查一个头信息是否是敏感的（不区分大小写）
func IsSensitiveHeader(headerKey string, sensitiveList []string) bool {
	lowerHeaderKey := strings.ToLower(headerKey)
	for _, sensitive := range sensitiveList {
		if strings.Contains(lowerHeaderKey, sensitive) {
			return true
		}
	}
	return false
}

// isCORSHeader 检查是否是CORS相关的头信息
func isCORSHeader(headerKey string) bool {
	lowerKey := strings.ToLower(headerKey)
	corsHeaders := []string{
		"access-control-allow-origin",
		"access-control-allow-methods",
		"access-control-allow-headers",
		"access-control-allow-credentials",
		"access-control-max-age",
		"access-control-expose-headers",
	}

	for _, corsHeader := range corsHeaders {
		if lowerKey == corsHeader {
			return true
		}
	}
	return false
}

// isAuthorizedForProxy 检查代理访问权限（保留向后兼容）
// 注意：这个函数保留用于向后兼容，新代码应该使用 ProxyAuthenticator
func isAuthorizedForProxy(r *http.Request, adminSecret string) bool {
	if adminSecret == "" {
		return false
	}

	// 检查请求头
	if secret := r.Header.Get("X-Log-Secret"); secret == adminSecret {
		return true
	}

	// 检查查询参数（向后兼容）
	if secret := r.URL.Query().Get("secret"); secret == adminSecret {
		return true
	}

	return false
}

// isAuthorizedForProxyWithToken 检查代理访问权限（支持令牌认证）
func isAuthorizedForProxyWithToken(r *http.Request, configID string, authenticator *ProxyAuthenticator, adminSecret string) *AuthResult {
	if authenticator == nil {
		// 回退到旧的认证方式
		if isAuthorizedForProxy(r, adminSecret) {
			return &AuthResult{
				Authenticated: true,
				Method:        "admin",
				ConfigID:      configID,
			}
		}
		return &AuthResult{
			Authenticated: false,
			Method:        "none",
			Error:         "Authentication required",
		}
	}

	return authenticator.AuthenticateForProxy(r, configID)
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头（代理环境）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP（客户端真实IP）
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 使用RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}
