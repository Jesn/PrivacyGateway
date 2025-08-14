package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 保存原始环境变量
	originalPort := os.Getenv("GATEWAY_PORT")
	originalHeaders := os.Getenv("SENSITIVE_HEADERS")
	originalProxy := os.Getenv("DEFAULT_PROXY")
	originalWhitelist := os.Getenv("PROXY_WHITELIST")
	originalAllowPrivate := os.Getenv("ALLOW_PRIVATE_PROXY")

	// 清理环境变量
	defer func() {
		os.Setenv("GATEWAY_PORT", originalPort)
		os.Setenv("SENSITIVE_HEADERS", originalHeaders)
		os.Setenv("DEFAULT_PROXY", originalProxy)
		os.Setenv("PROXY_WHITELIST", originalWhitelist)
		os.Setenv("ALLOW_PRIVATE_PROXY", originalAllowPrivate)
	}()

	t.Run("Default Values", func(t *testing.T) {
		// 清理所有环境变量
		os.Unsetenv("GATEWAY_PORT")
		os.Unsetenv("SENSITIVE_HEADERS")
		os.Unsetenv("DEFAULT_PROXY")
		os.Unsetenv("PROXY_WHITELIST")
		os.Unsetenv("ALLOW_PRIVATE_PROXY")

		config := Load()

		if config.Port != "8080" {
			t.Errorf("Expected default port 8080, got %s", config.Port)
		}

		if len(config.SensitiveHeaders) == 0 {
			t.Error("Expected sensitive headers to be loaded")
		}

		if config.DefaultProxy != nil {
			t.Error("Expected no default proxy")
		}

		if len(config.ProxyWhitelist) != 0 {
			t.Error("Expected empty proxy whitelist")
		}

		if config.AllowPrivateIP {
			t.Error("Expected AllowPrivateIP to be false by default")
		}
	})

	t.Run("Custom Values", func(t *testing.T) {
		os.Setenv("GATEWAY_PORT", "9090")
		os.Setenv("SENSITIVE_HEADERS", "test1,test2")
		os.Setenv("DEFAULT_PROXY", "http://proxy.example.com:8080")
		os.Setenv("PROXY_WHITELIST", "proxy1.com, proxy2.com")
		os.Setenv("ALLOW_PRIVATE_PROXY", "true")

		config := Load()

		if config.Port != "9090" {
			t.Errorf("Expected port 9090, got %s", config.Port)
		}

		if len(config.SensitiveHeaders) != 2 {
			t.Errorf("Expected 2 sensitive headers, got %d", len(config.SensitiveHeaders))
		}

		if config.DefaultProxy == nil {
			t.Error("Expected default proxy to be set")
		} else if config.DefaultProxy.URL != "http://proxy.example.com:8080" {
			t.Errorf("Expected proxy URL http://proxy.example.com:8080, got %s", config.DefaultProxy.URL)
		}

		if len(config.ProxyWhitelist) != 2 {
			t.Errorf("Expected 2 whitelist entries, got %d", len(config.ProxyWhitelist))
		}

		if !config.AllowPrivateIP {
			t.Error("Expected AllowPrivateIP to be true")
		}
	})
}

func TestParseSimpleProxy(t *testing.T) {
	testCases := []struct {
		name        string
		proxyURL    string
		expectError bool
		expectType  string
		expectAuth  bool
	}{
		{
			name:        "Empty URL",
			proxyURL:    "",
			expectError: false,
			expectType:  "",
			expectAuth:  false,
		},
		{
			name:        "HTTP Proxy",
			proxyURL:    "http://proxy.example.com:8080",
			expectError: false,
			expectType:  "http",
			expectAuth:  false,
		},
		{
			name:        "SOCKS5 Proxy",
			proxyURL:    "socks5://proxy.example.com:1080",
			expectError: false,
			expectType:  "socks5",
			expectAuth:  false,
		},
		{
			name:        "Proxy with Auth",
			proxyURL:    "http://user:pass@proxy.example.com:8080",
			expectError: false,
			expectType:  "http",
			expectAuth:  true,
		},
		{
			name:        "Invalid URL",
			proxyURL:    "://invalid-url",
			expectError: true,
			expectType:  "",
			expectAuth:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := parseSimpleProxy(tc.proxyURL)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.proxyURL == "" {
				if config != nil {
					t.Error("Expected nil config for empty URL")
				}
				return
			}

			if config == nil {
				t.Error("Expected non-nil config")
				return
			}

			if config.Type != tc.expectType {
				t.Errorf("Expected type %s, got %s", tc.expectType, config.Type)
			}

			if tc.expectAuth {
				if config.Auth == nil {
					t.Error("Expected auth info")
				} else if config.Auth.Username != "user" || config.Auth.Password != "pass" {
					t.Errorf("Expected user:pass, got %s:%s", config.Auth.Username, config.Auth.Password)
				}
			} else {
				if config.Auth != nil {
					t.Error("Expected no auth info")
				}
			}
		})
	}
}
