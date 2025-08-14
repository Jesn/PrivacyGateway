package proxy

import (
	"testing"

	"privacygateway/internal/config"
)

func TestValidate(t *testing.T) {
	t.Run("Nil Config", func(t *testing.T) {
		err := Validate(nil, nil, false)
		if err != nil {
			t.Errorf("Expected no error for nil config, got: %v", err)
		}
	})

	t.Run("Empty URL", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{URL: ""}
		err := Validate(proxyConfig, nil, false)
		if err != nil {
			t.Errorf("Expected no error for empty URL, got: %v", err)
		}
	})

	t.Run("Valid HTTP Proxy", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://proxy.example.com:8080",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, false)
		if err != nil {
			t.Errorf("Expected no error for valid HTTP proxy, got: %v", err)
		}
	})

	t.Run("Valid SOCKS5 Proxy", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "socks5://proxy.example.com:1080",
			Type: "socks5",
		}
		err := Validate(proxyConfig, nil, false)
		if err != nil {
			t.Errorf("Expected no error for valid SOCKS5 proxy, got: %v", err)
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "invalid-url",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, false)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("Unsupported Scheme", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "ftp://proxy.example.com:21",
			Type: "ftp",
		}
		err := Validate(proxyConfig, nil, false)
		if err == nil {
			t.Error("Expected error for unsupported scheme")
		}
	})

	t.Run("Empty Host", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, false)
		if err == nil {
			t.Error("Expected error for empty host")
		}
	})

	t.Run("Private IP Blocked", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://127.0.0.1:8080",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, false)
		if err == nil {
			t.Error("Expected error for private IP when not allowed")
		}
	})

	t.Run("Private IP Allowed", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://127.0.0.1:8080",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, true)
		if err != nil {
			t.Errorf("Expected no error for private IP when allowed, got: %v", err)
		}
	})

	t.Run("Whitelist Check - Allowed", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://allowed.proxy.com:8080",
			Type: "http",
		}
		whitelist := []string{"allowed.proxy.com", "other.proxy.com"}
		err := Validate(proxyConfig, whitelist, false)
		if err != nil {
			t.Errorf("Expected no error for whitelisted proxy, got: %v", err)
		}
	})

	t.Run("Whitelist Check - Blocked", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://blocked.proxy.com:8080",
			Type: "http",
		}
		whitelist := []string{"allowed.proxy.com", "other.proxy.com"}
		err := Validate(proxyConfig, whitelist, false)
		if err == nil {
			t.Error("Expected error for non-whitelisted proxy")
		}
	})

	t.Run("No Whitelist", func(t *testing.T) {
		proxyConfig := &config.ProxyConfig{
			URL:  "http://any.proxy.com:8080",
			Type: "http",
		}
		err := Validate(proxyConfig, nil, false)
		if err != nil {
			t.Errorf("Expected no error when no whitelist is set, got: %v", err)
		}
	})
}

func TestIsPrivateIP(t *testing.T) {
	testCases := []struct {
		host     string
		expected bool
	}{
		{"127.0.0.1", true},
		{"localhost", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"169.254.1.1", true},
		{"::1", true},
		{"8.8.8.8", false},
		{"google.com", false},
		{"172.15.0.1", false},
		{"172.32.0.1", false},
		{"193.168.1.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.host, func(t *testing.T) {
			result := isPrivateIP(tc.host)
			if result != tc.expected {
				t.Errorf("For host %s, expected %v, got %v", tc.host, tc.expected, result)
			}
		})
	}
}
