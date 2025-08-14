package handler

import (
	"io"
	"net/http"
	"net/url"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxy"
)

// HTTPProxy 处理HTTP代理请求
func HTTPProxy(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger) {
	targetStr := r.URL.Query().Get("target")
	if targetStr == "" {
		http.Error(w, "'target' query parameter is required", http.StatusBadRequest)
		return
	}

	targetURL, err := url.Parse(targetStr)
	if err != nil || targetURL.Host == "" {
		log.Error("failed to parse target URL", "input", targetStr, "error", err)
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	// 获取代理配置
	proxyConfig, err := proxy.GetConfig(r, cfg.DefaultProxy)
	if err != nil {
		log.Error("failed to parse proxy config", "error", err)
		http.Error(w, "Invalid proxy configuration", http.StatusBadRequest)
		return
	}

	// 验证代理配置安全性
	if err := proxy.Validate(proxyConfig, cfg.ProxyWhitelist, cfg.AllowPrivateIP); err != nil {
		log.Error("proxy validation failed", "error", err)
		http.Error(w, "Proxy not allowed", http.StatusForbidden)
		return
	}

	// 记录请求信息（不泄露敏感代理信息）
	if proxyConfig != nil && proxyConfig.URL != "" {
		log.Info("forwarding request via proxy", "method", r.Method, "target", targetURL.String(), "proxy_type", proxyConfig.Type)
	} else {
		log.Info("forwarding request", "method", r.Method, "target", targetURL.String())
	}

	// 创建转发请求
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		log.Error("failed to create proxy request", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 复制并过滤头信息
	for key, values := range r.Header {
		if !IsSensitiveHeader(key, cfg.SensitiveHeaders) {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}
	// 设置正确的主机头
	proxyReq.Host = targetURL.Host

	// 创建HTTP客户端（支持代理）
	client, err := proxy.CreateHTTPClient(proxyConfig)
	if err != nil {
		log.Error("failed to create HTTP client", "error", err)
		http.Error(w, "Failed to create proxy client", http.StatusInternalServerError)
		return
	}

	// 执行请求
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Error("failed to execute proxy request", "error", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 将目标服务器的响应复制回客户端
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
