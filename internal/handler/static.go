package handler

import (
	"net/http"
	"os"

	"privacygateway/internal/logger"
)

// Static 处理静态文件请求（主要是index.html）
func Static(w http.ResponseWriter, r *http.Request, log *logger.Logger) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
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
}
