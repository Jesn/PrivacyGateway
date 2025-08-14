package proxy

import (
	"fmt"
	"net/url"
	"strings"

	"privacygateway/internal/config"
)

// Validate 验证代理配置的安全性
func Validate(proxyConfig *config.ProxyConfig, whitelist []string, allowPrivateIP bool) error {
	if proxyConfig == nil || proxyConfig.URL == "" {
		return nil
	}

	// 解析代理URL
	proxyURL, err := url.Parse(proxyConfig.URL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}

	// 检查协议
	if proxyURL.Scheme != "http" && proxyURL.Scheme != "https" && proxyURL.Scheme != "socks5" {
		return fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}

	// 检查主机名
	if proxyURL.Host == "" {
		return fmt.Errorf("proxy host cannot be empty")
	}

	// SSRF防护：禁止访问内网地址（除非明确允许）
	if !allowPrivateIP && isPrivateIP(proxyURL.Hostname()) {
		return fmt.Errorf("proxy cannot point to private IP addresses")
	}

	// 白名单检查
	if len(whitelist) > 0 {
		allowed := false
		for _, allowedHost := range whitelist {
			if strings.Contains(proxyURL.Host, allowedHost) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("proxy host not in whitelist")
		}
	}

	return nil
}

// isPrivateIP 检查是否为私有IP地址
func isPrivateIP(host string) bool {
	// 简单的私有IP检查
	privateRanges := []string{
		"127.", "10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.", "169.254.", "::1", "localhost",
	}

	for _, private := range privateRanges {
		if strings.HasPrefix(host, private) {
			return true
		}
	}
	return false
}
