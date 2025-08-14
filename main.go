package main

import (
	"net/http"

	"privacygateway/internal/config"
	"privacygateway/internal/handler"
	"privacygateway/internal/logger"
)

func main() {
	// 初始化日志记录器
	log := logger.New()

	// 加载配置
	cfg := config.Load()

	log.Info("starting Privacy Gateway", "port", cfg.Port)

	// 设置路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.Static(w, r, log)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		handler.HTTPProxy(w, r, cfg, log)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.WebSocket(w, r, cfg, log)
	})

	// 启动服务器
	err := http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		log.Error("server failed to start", "error", err)
	}
}
