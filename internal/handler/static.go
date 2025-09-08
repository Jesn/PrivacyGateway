package handler

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"privacygateway/internal/logger"
)

// Static 处理静态文件请求（支持新的模块化前端结构）
func Static(w http.ResponseWriter, r *http.Request, log *logger.Logger) {
	// 注意：CORS头部已在路由层设置，这里不再重复设置

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 获取不带查询参数的路径
	path := r.URL.Path

	// 处理根路径
	if path == "/" {
		serveIndexHTML(w, log)
		return
	}

	// 处理favicon
	if path == "/favicon.ico" {
		serveFavicon(w, log)
		return
	}

	// 处理前端资源文件
	if strings.HasPrefix(path, "/assets/") {
		serveAsset(w, r, log)
		return
	}

	// 其他路径返回404
	http.NotFound(w, r)
}

// serveIndexHTML 提供主页面
func serveIndexHTML(w http.ResponseWriter, log *logger.Logger) {
	// 优先使用优化版本，然后是标准版本
	indexPaths := []string{
		"frontend/index-optimized.html", // Tailwind CSS优化版本
		"frontend/index.html",           // 标准版本
		"index.html",                    // 兼容旧版本
	}

	var indexFile []byte
	var err error

	for _, indexPath := range indexPaths {
		indexFile, err = os.ReadFile(indexPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Error("failed to read index.html", "error", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Privacy Gateway is running. Use /proxy for HTTP and /ws for WebSocket."))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // 开发时不缓存
	w.WriteHeader(http.StatusOK)
	w.Write(indexFile)
}

// serveFavicon 提供网站图标
func serveFavicon(w http.ResponseWriter, log *logger.Logger) {
	faviconPaths := []string{
		"frontend/favicon.ico",
		"favicon.ico", // 兼容旧版本
	}

	var faviconFile []byte
	var err error

	for _, faviconPath := range faviconPaths {
		faviconFile, err = os.ReadFile(faviconPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Error("failed to read favicon.ico", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 缓存1天
	w.WriteHeader(http.StatusOK)
	w.Write(faviconFile)
}

// serveAsset 提供前端资源文件
func serveAsset(w http.ResponseWriter, r *http.Request, log *logger.Logger) {
	// 移除 /assets 前缀，构建实际文件路径
	assetPath := strings.TrimPrefix(r.URL.Path, "/assets")
	filePath := filepath.Join("frontend/assets", assetPath)

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filePath, "..") {
		log.Warn("potential path traversal attempt", "path", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	// 读取文件
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Debug("asset file not found", "path", filePath, "error", err)
		http.NotFound(w, r)
		return
	}

	// 根据文件扩展名设置Content-Type
	ext := filepath.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		// 默认Content-Type
		switch ext {
		case ".js":
			contentType = "application/javascript; charset=utf-8"
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".json":
			contentType = "application/json; charset=utf-8"
		default:
			contentType = "text/plain; charset=utf-8"
		}
	}

	// 设置缓存策略
	cacheControl := "no-cache, no-store, must-revalidate" // 开发时不缓存
	if strings.Contains(r.Header.Get("User-Agent"), "production") {
		// 生产环境可以设置更长的缓存时间
		cacheControl = "public, max-age=3600" // 缓存1小时
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", cacheControl)
	w.WriteHeader(http.StatusOK)
	w.Write(fileContent)

	log.Debug("served asset", "path", filePath, "size", len(fileContent))
}
