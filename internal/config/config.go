package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Load 从环境变量加载配置
func Load() *Config {
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "10805"
	}

	sensitiveHeadersStr := os.Getenv("SENSITIVE_HEADERS")
	if sensitiveHeadersStr == "" {
		sensitiveHeadersStr = "cf-,x-forwarded,proxy,via,x-request-id,x-trace,x-correlation-id,x-country,x-region,x-city"
	}

	// 加载默认代理配置
	var defaultProxy *ProxyConfig
	if defaultProxyURL := os.Getenv("DEFAULT_PROXY"); defaultProxyURL != "" {
		if proxy, err := parseSimpleProxy(defaultProxyURL); err == nil {
			defaultProxy = proxy
		}
	}

	// 加载代理白名单
	var proxyWhitelist []string
	if whitelistStr := os.Getenv("PROXY_WHITELIST"); whitelistStr != "" {
		proxyWhitelist = strings.Split(whitelistStr, ",")
		// 清理空白字符
		for i, host := range proxyWhitelist {
			proxyWhitelist[i] = strings.TrimSpace(host)
		}
	}

	// 是否允许私有IP代理（用于开发测试）
	allowPrivateIP := os.Getenv("ALLOW_PRIVATE_PROXY") == "true"

	// 加载管理相关配置
	adminSecret := os.Getenv("ADMIN_SECRET")

	logMaxEntries := 1000
	if val := os.Getenv("LOG_MAX_ENTRIES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			logMaxEntries = parsed
		}
	}

	logMaxBodySize := 1024
	if val := os.Getenv("LOG_MAX_BODY_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			logMaxBodySize = parsed
		}
	}

	logRetentionHours := 24
	if val := os.Getenv("LOG_RETENTION_HOURS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			logRetentionHours = parsed
		}
	}

	logMaxMemoryMB := 50.0
	if val := os.Getenv("LOG_MAX_MEMORY_MB"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed > 0 {
			logMaxMemoryMB = parsed
		}
	}

	// 是否记录200状态码的详细信息（默认false，只记录非200状态码）
	logRecord200 := os.Getenv("LOG_RECORD_200") == "true"

	return &Config{
		Port:             port,
		SensitiveHeaders: strings.Split(strings.ToLower(sensitiveHeadersStr), ","),
		DefaultProxy:     defaultProxy,
		ProxyWhitelist:   proxyWhitelist,
		AllowPrivateIP:   allowPrivateIP,

		// 管理配置
		AdminSecret:       adminSecret,
		LogMaxEntries:     logMaxEntries,
		LogMaxBodySize:    logMaxBodySize,
		LogRetentionHours: logRetentionHours,
		LogMaxMemoryMB:    logMaxMemoryMB,
		LogRecord200:      logRecord200,
	}
}

// parseSimpleProxy 解析简单的代理URL（内部辅助函数）
func parseSimpleProxy(proxyURL string) (*ProxyConfig, error) {
	if proxyURL == "" {
		return nil, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	config := &ProxyConfig{
		URL:     proxyURL,
		Type:    u.Scheme,
		Timeout: 30, // 默认30秒超时
	}

	// 解析认证信息
	if u.User != nil {
		config.Auth = &ProxyAuth{
			Username: u.User.Username(),
		}
		if password, ok := u.User.Password(); ok {
			config.Auth.Password = password
		}
	}

	return config, nil
}
