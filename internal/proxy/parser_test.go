package proxy

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"privacygateway/internal/config"
)

func TestParseSimple(t *testing.T) {
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
			proxyURL:    "http://proxy.example.com:10805",
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
			proxyURL:    "http://user:pass@proxy.example.com:10805",
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
			proxyConfig, err := ParseSimple(tc.proxyURL)

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
				if proxyConfig != nil {
					t.Error("Expected nil config for empty URL")
				}
				return
			}

			if proxyConfig == nil {
				t.Error("Expected non-nil config")
				return
			}

			if proxyConfig.Type != tc.expectType {
				t.Errorf("Expected type %s, got %s", tc.expectType, proxyConfig.Type)
			}

			if tc.expectAuth {
				if proxyConfig.Auth == nil {
					t.Error("Expected auth info")
				} else if proxyConfig.Auth.Username != "user" || proxyConfig.Auth.Password != "pass" {
					t.Errorf("Expected user:pass, got %s:%s", proxyConfig.Auth.Username, proxyConfig.Auth.Password)
				}
			} else {
				if proxyConfig.Auth != nil {
					t.Error("Expected no auth info")
				}
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	t.Run("Empty Header", func(t *testing.T) {
		config, err := ParseHeader("")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if config != nil {
			t.Error("Expected nil config for empty header")
		}
	})

	t.Run("Valid JSON Config", func(t *testing.T) {
		proxyConfig := config.ProxyConfig{
			URL:     "http://proxy.example.com:10805",
			Type:    "http",
			Timeout: 60,
			Auth: &config.ProxyAuth{
				Username: "testuser",
				Password: "testpass",
			},
		}

		jsonData, _ := json.Marshal(proxyConfig)
		encoded := base64.StdEncoding.EncodeToString(jsonData)

		result, err := ParseHeader(encoded)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
			return
		}

		if result.URL != proxyConfig.URL {
			t.Errorf("Expected URL %s, got %s", proxyConfig.URL, result.URL)
		}

		if result.Type != proxyConfig.Type {
			t.Errorf("Expected type %s, got %s", proxyConfig.Type, result.Type)
		}

		if result.Timeout != proxyConfig.Timeout {
			t.Errorf("Expected timeout %d, got %d", proxyConfig.Timeout, result.Timeout)
		}

		if result.Auth == nil {
			t.Error("Expected auth info")
		} else {
			if result.Auth.Username != proxyConfig.Auth.Username {
				t.Errorf("Expected username %s, got %s", proxyConfig.Auth.Username, result.Auth.Username)
			}
			if result.Auth.Password != proxyConfig.Auth.Password {
				t.Errorf("Expected password %s, got %s", proxyConfig.Auth.Password, result.Auth.Password)
			}
		}
	})

	t.Run("Invalid Base64", func(t *testing.T) {
		_, err := ParseHeader("invalid-base64!")
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("invalid-json"))
		_, err := ParseHeader(encoded)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestGetConfig(t *testing.T) {
	defaultProxy := &config.ProxyConfig{
		URL:  "http://default.proxy.com:10805",
		Type: "http",
	}

	t.Run("Header Priority", func(t *testing.T) {
		proxyConfig := config.ProxyConfig{
			URL:  "http://header.proxy.com:10805",
			Type: "http",
		}
		jsonData, _ := json.Marshal(proxyConfig)
		encoded := base64.StdEncoding.EncodeToString(jsonData)

		req := &http.Request{
			Header: http.Header{
				"X-Proxy-Config": []string{encoded},
			},
			URL: &url.URL{
				RawQuery: "proxy=http://query.proxy.com:10805",
			},
		}

		result, err := GetConfig(req, defaultProxy)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil || result.URL != proxyConfig.URL {
			t.Error("Expected header config to take priority")
		}
	})

	t.Run("Query Parameter Fallback", func(t *testing.T) {
		req := &http.Request{
			Header: http.Header{},
			URL: &url.URL{
				RawQuery: "proxy=http://query.proxy.com:10805",
			},
		}

		result, err := GetConfig(req, defaultProxy)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil || result.URL != "http://query.proxy.com:10805" {
			t.Error("Expected query parameter config")
		}
	})

	t.Run("Default Fallback", func(t *testing.T) {
		req := &http.Request{
			Header: http.Header{},
			URL:    &url.URL{},
		}

		result, err := GetConfig(req, defaultProxy)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != defaultProxy {
			t.Error("Expected default config")
		}
	})
}
