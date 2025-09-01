package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
	"privacygateway/internal/router"
)

func main() {
	// 初始化日志记录器
	log := logger.New()

	// 加载配置
	cfg := config.Load()

	// 创建访问日志记录器
	var recorder *accesslog.Recorder
	if cfg.AdminSecret != "" {
		var err error
		recorder, err = accesslog.NewRecorder(cfg, log)
		if err != nil {
			log.Error("failed to create access log recorder", "error", err)
		} else {
			log.Info("access log recorder initialized")
		}
	}

	// 创建代理配置存储
	var configStorage proxyconfig.Storage

	// 检查是否禁用持久化存储（默认启用）
	persistDisabled := os.Getenv("PROXY_CONFIG_PERSIST") == "false"
	if persistDisabled {
		configStorage = proxyconfig.NewMemoryStorage(1000)
		log.Info("memory config storage initialized", "max_entries", 1000)
	} else {
		configFile := os.Getenv("PROXY_CONFIG_FILE")
		if configFile == "" {
			configFile = "data/proxy-configs.json"
		}
		autoSave := os.Getenv("PROXY_CONFIG_AUTO_SAVE") != "false"
		configStorage = proxyconfig.NewPersistentStorage(configFile, 1000, autoSave, log)
		log.Info("persistent config storage initialized", "file", configFile, "auto_save", autoSave)
	}

	// 创建并设置路由
	appRouter := router.NewRouter(cfg, log, recorder, configStorage)
	appRouter.SetupRoutes()

	// 打印路由信息
	appRouter.PrintRoutes()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      nil,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info("starting Privacy Gateway", "port", cfg.Port)

	// 在goroutine中启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// 创建关闭上下文，30秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	// 清理资源
	if recorder != nil {
		if err := recorder.Close(); err != nil {
			log.Error("failed to close access log recorder", "error", err)
		}
	}

	// 如果配置存储实现了Closer接口，也要关闭它
	if closer, ok := configStorage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Error("failed to close config storage", "error", err)
		}
	}

	log.Info("server exited gracefully")
}
