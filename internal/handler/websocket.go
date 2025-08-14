package handler

import (
	"net/http"
	"net/url"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxy"

	"github.com/gorilla/websocket"
)

// upgrader is used to upgrade the HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{
	// Allow any origin, for simplicity. In a real-world scenario, you might want to restrict this.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket handles WebSocket proxying with optional upstream proxy support.
func WebSocket(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger) {
	targetURLStr := r.URL.Query().Get("target")
	if targetURLStr == "" {
		http.Error(w, "'target' query parameter is required for WebSocket proxy", http.StatusBadRequest)
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

	// 记录连接信息
	if proxyConfig != nil && proxyConfig.URL != "" {
		log.Info("connecting to target WebSocket via proxy", "target", targetURLStr, "proxy_type", proxyConfig.Type)
	} else {
		log.Info("connecting to target WebSocket", "target", targetURLStr)
	}

	// Prepare the request headers to be forwarded to the target server.
	// It's important to only forward necessary headers, not all of them.
	requestHeader := http.Header{}
	if origin := r.Header.Get("Origin"); origin != "" {
		requestHeader.Add("Origin", origin)
	}
	if protocol := r.Header.Get("Sec-WebSocket-Protocol"); protocol != "" {
		requestHeader.Add("Sec-WebSocket-Protocol", protocol)
	}
	if extensions := r.Header.Get("Sec-WebSocket-Extensions"); extensions != "" {
		requestHeader.Add("Sec-WebSocket-Extensions", extensions)
	}

	// 创建WebSocket拨号器，支持代理
	dialer := &websocket.Dialer{}

	// 如果有代理配置，设置代理
	if proxyConfig != nil && proxyConfig.URL != "" {
		proxyURL, err := url.Parse(proxyConfig.URL)
		if err != nil {
			log.Error("failed to parse proxy URL", "error", err)
			http.Error(w, "Invalid proxy URL", http.StatusBadRequest)
			return
		}

		switch proxyConfig.Type {
		case "http", "https":
			// HTTP代理
			dialer.Proxy = http.ProxyURL(proxyURL)
		case "socks5":
			// SOCKS5代理 - 暂时不支持，因为WebSocket的SOCKS5代理实现比较复杂
			log.Error("SOCKS5 proxy not yet supported for WebSocket", "proxy_url", proxyConfig.URL)
			http.Error(w, "SOCKS5 proxy not yet supported for WebSocket", http.StatusNotImplemented)
			return
		default:
			log.Error("unsupported proxy type for WebSocket", "type", proxyConfig.Type)
			http.Error(w, "Unsupported proxy type for WebSocket", http.StatusBadRequest)
			return
		}
	}

	// Connect to the target WebSocket server with the prepared headers.
	targetConn, _, err := dialer.Dial(targetURLStr, requestHeader)
	if err != nil {
		log.Error("failed to dial target WebSocket server", "error", err)
		http.Error(w, "could not connect to target WebSocket server", http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Upgrade the client's connection.
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade client connection", "error", err)
		return
	}
	defer clientConn.Close()

	log.Info("WebSocket connections established, starting proxying")

	// Create channels for error handling
	done := make(chan struct{})

	// Goroutine to copy messages from target to client.
	go func() {
		defer close(done)
		for {
			messageType, p, err := targetConn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Error("error reading from target", "error", err)
				}
				return
			}
			if err := clientConn.WriteMessage(messageType, p); err != nil {
				log.Error("error writing to client", "error", err)
				return
			}
		}
	}()

	// Copy messages from client to target in the main goroutine (blocking).
	for {
		select {
		case <-done:
			return
		default:
			messageType, p, err := clientConn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Error("error reading from client", "error", err)
				}
				return
			}
			if err := targetConn.WriteMessage(messageType, p); err != nil {
				log.Error("error writing to target", "error", err)
				return
			}
		}
	}
}
