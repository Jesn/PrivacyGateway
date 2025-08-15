package main

import (
	"net/http"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/handler"
	"privacygateway/internal/logger"
	"privacygateway/internal/logviewer"
)

func main() {
	// 初始化日志记录器
	log := logger.New()

	// 加载配置
	cfg := config.Load()

	// 创建访问日志记录器
	var recorder *accesslog.Recorder
	if cfg.LogViewSecret != "" {
		var err error
		recorder, err = accesslog.NewRecorder(cfg, log)
		if err != nil {
			log.Error("failed to create access log recorder", "error", err)
		} else {
			log.Info("access log recorder initialized")
		}
	}

	log.Info("starting Privacy Gateway", "port", cfg.Port)

	// 设置路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.Static(w, r, log)
	})

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		handler.HTTPProxy(w, r, cfg, log, recorder)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.WebSocket(w, r, cfg, log)
	})

	// 设置日志查看路由
	if recorder != nil {
		// 先验证密钥
		if _, err := logviewer.CreateAuthenticator(cfg.LogViewSecret); err != nil {
			log.Error("log viewer configuration error", "error", err.Error())
			log.Info("log viewer disabled", "reason", "invalid secret configuration")
			// 即使配置错误，也要注册路由来显示错误页面
			logHandler := logviewer.CreateLogViewHandler(recorder, cfg.LogViewSecret, log)
			http.HandleFunc("/logs", logHandler)
			http.HandleFunc("/logs/", logHandler)
		} else {
			logHandler := logviewer.CreateLogViewHandler(recorder, cfg.LogViewSecret, log)
			http.HandleFunc("/logs", logHandler)
			http.HandleFunc("/logs/", logHandler)
			log.Info("log viewer enabled", "path", "/logs")
		}
	} else {
		log.Info("log viewer disabled", "reason", "no secret configured")
	}

	// 启动服务器
	err := http.ListenAndServe(":"+cfg.Port, nil)
	if err != nil {
		log.Error("server failed to start", "error", err)
	}
}
