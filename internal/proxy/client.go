package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"privacygateway/internal/config"

	"golang.org/x/net/proxy"
)

// CreateHTTPClient 根据代理配置创建HTTP客户端
func CreateHTTPClient(proxyConfig *config.ProxyConfig) (*http.Client, error) {
	// 默认客户端配置
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果没有代理配置，返回默认客户端
	if proxyConfig == nil || proxyConfig.URL == "" {
		return client, nil
	}

	// 设置超时
	if proxyConfig.Timeout > 0 {
		client.Timeout = time.Duration(proxyConfig.Timeout) * time.Second
	}

	// 解析代理URL
	proxyURL, err := url.Parse(proxyConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %v", err)
	}

	// 根据代理类型创建传输层
	switch proxyConfig.Type {
	case "http", "https":
		// HTTP代理
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport

	case "socks5":
		// SOCKS5代理
		var auth *proxy.Auth
		if proxyConfig.Auth != nil {
			auth = &proxy.Auth{
				User:     proxyConfig.Auth.Username,
				Password: proxyConfig.Auth.Password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 proxy: %v", err)
		}

		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		client.Transport = transport

	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxyConfig.Type)
	}

	return client, nil
}
