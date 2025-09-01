package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

// AuthResult 认证结果
type AuthResult struct {
	Authenticated    bool                               `json:"authenticated"`
	Method           string                             `json:"method"`            // "admin" 或 "token"
	ConfigID         string                             `json:"config_id"`         // 令牌认证时的配置ID
	Token            *proxyconfig.AccessToken           `json:"token"`             // 令牌信息
	ValidationResult *proxyconfig.TokenValidationResult `json:"validation_result"` // 令牌验证结果
	Error            string                             `json:"error"`
}

// ProxyAuthenticator 代理认证器
type ProxyAuthenticator struct {
	adminSecret string
	storage     proxyconfig.Storage
	logger      *logger.Logger
}

// NewProxyAuthenticator 创建代理认证器
func NewProxyAuthenticator(adminSecret string, storage proxyconfig.Storage, logger *logger.Logger) *ProxyAuthenticator {
	return &ProxyAuthenticator{
		adminSecret: adminSecret,
		storage:     storage,
		logger:      logger,
	}
}

// AuthenticateForProxy 代理请求认证
func (pa *ProxyAuthenticator) AuthenticateForProxy(r *http.Request, configID string) *AuthResult {
	startTime := time.Now()

	// 首先尝试管理员密钥认证
	if pa.authenticateAdmin(r) {
		pa.logger.Debug("admin authentication successful",
			"client_ip", getClientIP(r),
			"duration", time.Since(startTime))

		return &AuthResult{
			Authenticated: true,
			Method:        "admin",
			ConfigID:      configID,
		}
	}

	// 尝试令牌认证
	tokenValue := pa.extractToken(r)
	if tokenValue == "" {
		pa.logger.Debug("authentication failed: no token provided",
			"client_ip", getClientIP(r),
			"config_id", configID)

		return &AuthResult{
			Authenticated: false,
			Method:        "none",
			Error:         "Authentication required: admin secret or access token",
		}
	}

	// 如果没有配置ID，尝试通过令牌反向查找
	if configID == "" {
		foundConfigID, err := pa.storage.FindConfigByToken(tokenValue)
		if err != nil {
			pa.logger.Debug("authentication failed: unable to find config by token",
				"client_ip", getClientIP(r),
				"error", err)

			return &AuthResult{
				Authenticated: false,
				Method:        "token",
				Error:         "Invalid token or token not found",
			}
		}
		configID = foundConfigID
		pa.logger.Debug("config ID found by token",
			"client_ip", getClientIP(r),
			"config_id", configID)
	}

	// 验证令牌
	validationResult, err := pa.storage.ValidateToken(configID, tokenValue)
	if err != nil {
		pa.logger.Error("token validation error",
			"error", err,
			"client_ip", getClientIP(r),
			"config_id", configID,
			"duration", time.Since(startTime))

		return &AuthResult{
			Authenticated: false,
			Method:        "token",
			ConfigID:      configID,
			Error:         "Token validation failed",
		}
	}

	if !validationResult.Valid {
		pa.logger.Warn("token authentication failed",
			"client_ip", getClientIP(r),
			"config_id", configID,
			"error_code", validationResult.ErrorCode,
			"error_msg", validationResult.ErrorMsg,
			"duration", time.Since(startTime))

		return &AuthResult{
			Authenticated:    false,
			Method:           "token",
			ConfigID:         configID,
			ValidationResult: validationResult,
			Error:            validationResult.ErrorMsg,
		}
	}

	// 令牌认证成功，更新使用统计
	if err := pa.storage.UpdateTokenUsage(configID, tokenValue); err != nil {
		pa.logger.Error("failed to update token usage",
			"error", err,
			"config_id", configID,
			"token_id", validationResult.Token.ID)
	}

	pa.logger.Info("token authentication successful",
		"client_ip", getClientIP(r),
		"config_id", configID,
		"token_id", validationResult.Token.ID,
		"token_name", validationResult.Token.Name,
		"duration", time.Since(startTime))

	return &AuthResult{
		Authenticated:    true,
		Method:           "token",
		ConfigID:         configID,
		Token:            validationResult.Token,
		ValidationResult: validationResult,
	}
}

// AuthenticateForConfig 配置管理认证（仅支持管理员密钥）
func (pa *ProxyAuthenticator) AuthenticateForConfig(r *http.Request) *AuthResult {
	if pa.authenticateAdmin(r) {
		return &AuthResult{
			Authenticated: true,
			Method:        "admin",
		}
	}

	return &AuthResult{
		Authenticated: false,
		Method:        "none",
		Error:         "Admin authentication required for configuration management",
	}
}

// authenticateAdmin 管理员密钥认证
func (pa *ProxyAuthenticator) authenticateAdmin(r *http.Request) bool {
	if pa.adminSecret == "" {
		return false
	}

	// 检查请求头
	if secret := r.Header.Get("X-Log-Secret"); secret == pa.adminSecret {
		return true
	}

	// 检查查询参数（向后兼容）
	if secret := r.URL.Query().Get("secret"); secret == pa.adminSecret {
		return true
	}

	return false
}

// extractToken 从请求中提取令牌
func (pa *ProxyAuthenticator) extractToken(r *http.Request) string {
	// 优先从专用的令牌头获取
	if token := r.Header.Get("X-Proxy-Token"); token != "" {
		return token
	}

	// 从查询参数获取
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	return ""
}

// ExtractConfigID 从请求中提取配置ID
func ExtractConfigID(r *http.Request) string {
	// 从URL路径提取（如 /config/proxy/{id}/...）
	path := r.URL.Path

	// 处理令牌管理API路径
	if strings.HasPrefix(path, "/config/proxy/") {
		parts := strings.Split(strings.TrimPrefix(path, "/config/proxy/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// 从查询参数获取
	if configID := r.URL.Query().Get("config_id"); configID != "" {
		return configID
	}

	// 从请求头获取
	if configID := r.Header.Get("X-Config-ID"); configID != "" {
		return configID
	}

	return ""
}

// ExtractConfigIDFromSubdomain 从子域名请求中提取配置ID
func ExtractConfigIDFromSubdomain(r *http.Request, storage proxyconfig.Storage) (string, error) {
	subdomain := extractSubdomain(r.Host)
	if subdomain == "" {
		return "", fmt.Errorf("invalid subdomain")
	}

	config, err := storage.GetBySubdomain(subdomain)
	if err != nil {
		return "", fmt.Errorf("subdomain not configured: %v", err)
	}

	return config.ID, nil
}

// LogAuthFailure 记录认证失败日志
func (pa *ProxyAuthenticator) LogAuthFailure(r *http.Request, result *AuthResult, context string) {
	pa.logger.Warn("authentication failed",
		"context", context,
		"method", result.Method,
		"config_id", result.ConfigID,
		"error", result.Error,
		"client_ip", getClientIP(r),
		"user_agent", r.Header.Get("User-Agent"),
		"path", r.URL.Path)
}

// IsTokenAuthenticationEnabled 检查是否启用了令牌认证
func (pa *ProxyAuthenticator) IsTokenAuthenticationEnabled() bool {
	return pa.storage != nil
}

// GetAuthenticationMethods 获取支持的认证方法
func (pa *ProxyAuthenticator) GetAuthenticationMethods() []string {
	methods := []string{}

	if pa.adminSecret != "" {
		methods = append(methods, "admin")
	}

	if pa.IsTokenAuthenticationEnabled() {
		methods = append(methods, "token")
	}

	return methods
}
