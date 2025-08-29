package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxy"
)

// HTTPProxy 处理HTTP代理请求
func HTTPProxy(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger, recorder *accesslog.Recorder) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Log-Secret")

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 认证检查 - 代理服务需要管理员权限
	if !isAuthorizedForProxy(r, cfg.AdminSecret) {
		log.Warn("unauthorized proxy request", "client_ip", getClientIP(r), "target", r.URL.Query().Get("target"))
		http.Error(w, "Unauthorized: Admin secret required", http.StatusUnauthorized)
		return
	}

	// 创建响应捕获器（如果有记录器）
	var capture *accesslog.ResponseCapture

	if recorder != nil {
		capture = accesslog.NewResponseCapture(w, true, cfg.LogMaxBodySize, cfg.LogRecord200)
		w = capture
	}

	// 确保在函数结束时记录日志
	defer func() {
		if recorder != nil && capture != nil {
			recorder.RecordFromCapture(r, capture, "/proxy")
		}
	}()

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

	// 读取请求体（如果有）
	var requestBody []byte
	if r.Body != nil {
		requestBody, err = io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read request body", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		r.Body.Close()
	}

	// 创建转发请求
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), bytes.NewReader(requestBody))
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

	// 记录请求信息到capture
	if capture != nil {
		// 记录实际发送给目标服务器的User-Agent
		actualUserAgent := proxyReq.Header.Get("User-Agent")
		if actualUserAgent == "" {
			actualUserAgent = "未设置"
		}
		capture.SetActualUserAgent(actualUserAgent)

		// 设置代理信息
		proxyInfoStr := "Privacy Gateway v1.0"
		if proxyConfig != nil && proxyConfig.URL != "" {
			proxyInfoStr += " (via " + proxyConfig.Type + " proxy)"
		}
		capture.SetProxyInfo(proxyInfoStr)

		// 捕获请求头信息（过滤敏感头）
		requestHeaders := make(map[string]string)
		for key, values := range proxyReq.Header {
			if !IsSensitiveHeader(key, cfg.SensitiveHeaders) && len(values) > 0 {
				requestHeaders[key] = values[0] // 只取第一个值
			}
		}
		capture.SetRequestHeaders(requestHeaders)

		// 捕获请求体（如果有且不是太大）
		if len(requestBody) > 0 && len(requestBody) <= cfg.LogMaxBodySize {
			capture.SetRequestBody(string(requestBody))
		}
	}

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

	// 将目标服务器的响应复制回客户端（过滤CORS头避免重复）
	for key, values := range resp.Header {
		// 跳过CORS相关的头，因为我们已经设置了
		if isCORSHeader(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
