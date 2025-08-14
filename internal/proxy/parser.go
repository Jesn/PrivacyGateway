package proxy

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"privacygateway/internal/config"
)

// ParseSimple 解析简单的代理URL（来自查询参数）
func ParseSimple(proxyURL string) (*config.ProxyConfig, error) {
	if proxyURL == "" {
		return nil, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	proxyConfig := &config.ProxyConfig{
		URL:     proxyURL,
		Type:    u.Scheme,
		Timeout: 30, // 默认30秒超时
	}

	// 解析认证信息
	if u.User != nil {
		proxyConfig.Auth = &config.ProxyAuth{
			Username: u.User.Username(),
		}
		if password, ok := u.User.Password(); ok {
			proxyConfig.Auth.Password = password
		}
	}

	return proxyConfig, nil
}

// ParseHeader 解析请求头中的代理配置（Base64编码的JSON）
func ParseHeader(headerValue string) (*config.ProxyConfig, error) {
	if headerValue == "" {
		return nil, nil
	}

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(headerValue)
	if err != nil {
		return nil, err
	}

	// JSON解析
	var proxyConfig config.ProxyConfig
	if err := json.Unmarshal(decoded, &proxyConfig); err != nil {
		return nil, err
	}

	// 设置默认值
	if proxyConfig.Timeout == 0 {
		proxyConfig.Timeout = 30
	}
	if proxyConfig.Type == "" && proxyConfig.URL != "" {
		if u, err := url.Parse(proxyConfig.URL); err == nil {
			proxyConfig.Type = u.Scheme
		}
	}

	return &proxyConfig, nil
}

// GetConfig 获取请求的代理配置（优先级：请求头 > 查询参数 > 默认配置）
func GetConfig(r *http.Request, defaultProxy *config.ProxyConfig) (*config.ProxyConfig, error) {
	// 优先检查请求头
	if headerConfig := r.Header.Get("X-Proxy-Config"); headerConfig != "" {
		return ParseHeader(headerConfig)
	}

	// 检查查询参数
	if queryProxy := r.URL.Query().Get("proxy"); queryProxy != "" {
		return ParseSimple(queryProxy)
	}

	// 返回默认配置
	return defaultProxy, nil
}
