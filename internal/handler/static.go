package handler

import (
	"net/http"
	"os"

	"privacygateway/internal/logger"
)

// Static 处理静态文件请求（主要是index.html和favicon）
func Static(w http.ResponseWriter, r *http.Request, log *logger.Logger) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 获取不带查询参数的路径
	path := r.URL.Path
	switch path {
	case "/":
		// 读取并返回 index.html 文件
		indexFile, err := os.ReadFile("index.html")
		if err != nil {
			log.Error("failed to read index.html", "error", err)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Privacy Gateway is running. Use /proxy for HTTP and /ws for WebSocket."))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexFile)

	case "/favicon.ico":
		// 读取并返回 favicon.ico 文件
		faviconFile, err := os.ReadFile("favicon.ico")
		if err != nil {
			log.Error("failed to read favicon.ico", "error", err)
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=86400") // 缓存1天
		w.WriteHeader(http.StatusOK)
		w.Write(faviconFile)

	default:
		http.NotFound(w, r)
	}
}
