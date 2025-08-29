package handler

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"privacygateway/internal/accesslog"
	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

// HandleSubdomainProxy 处理子域名代理请求
func HandleSubdomainProxy(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger, recorder *accesslog.Recorder, configStorage proxyconfig.Storage) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Log-Secret")

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 认证检查 - 子域名代理服务需要管理员权限
	if !isAuthorizedForProxy(r, cfg.AdminSecret) {
		log.Warn("unauthorized subdomain proxy request", "client_ip", getClientIP(r), "host", r.Host)
		http.Error(w, "Unauthorized: Admin secret required", http.StatusUnauthorized)
		return
	}

	// 提取子域名
	subdomain := extractSubdomain(r.Host)
	if subdomain == "" {
		http.Error(w, "Invalid subdomain", http.StatusBadRequest)
		return
	}

	// 查找配置
	config, err := configStorage.GetBySubdomain(subdomain)
	if err != nil {
		log.Debug("subdomain config not found", "subdomain", subdomain, "error", err)
		http.Error(w, "Subdomain not configured", http.StatusNotFound)
		return
	}

	// 构建目标URL
	targetURL := buildTargetURL(config, r)

	// 创建新的请求，模拟查询参数方式
	newReq := r.Clone(r.Context())
	newReq.URL.RawQuery = "target=" + url.QueryEscape(targetURL)

	log.Info("subdomain proxy request", "subdomain", subdomain, "target", targetURL, "path", r.URL.Path)

	// 记录开始时间用于统计
	startTime := time.Now()

	// 调用现有的代理处理逻辑
	HTTPProxy(w, newReq, cfg, log, recorder)

	// 更新统计信息
	responseTime := time.Since(startTime)
	success := true   // 这里简化处理，实际应该根据响应状态判断
	bytes := int64(0) // 这里简化处理，实际应该统计传输字节数

	if err := configStorage.UpdateStats(config.ID, responseTime, success, bytes); err != nil {
		log.Error("failed to update stats", "config_id", config.ID, "error", err)
	}
}

// IsSubdomainProxy 检查是否是子域名代理请求
func IsSubdomainProxy(host string) bool {
	// 解析主机名
	parts := strings.Split(host, ":")
	hostname := parts[0]

	// 检查是否有子域名
	domainParts := strings.Split(hostname, ".")
	if len(domainParts) < 2 {
		return false
	}

	// 检查是否是localhost的子域名
	if len(domainParts) >= 2 && domainParts[len(domainParts)-1] == "localhost" {
		return domainParts[0] != "localhost" // 有子域名且不是直接的localhost
	}

	// 检查是否是其他域名的子域名
	if len(domainParts) >= 3 {
		return true
	}

	return false
}

// extractSubdomain 提取子域名
func extractSubdomain(host string) string {
	// 移除端口号
	parts := strings.Split(host, ":")
	hostname := parts[0]

	// 分割域名
	domainParts := strings.Split(hostname, ".")
	if len(domainParts) < 2 {
		return ""
	}

	// 对于localhost，返回第一部分作为子域名
	if len(domainParts) >= 2 && domainParts[len(domainParts)-1] == "localhost" {
		if domainParts[0] != "localhost" {
			return domainParts[0]
		}
		return ""
	}

	// 对于其他域名，返回第一部分作为子域名
	if len(domainParts) >= 3 {
		return domainParts[0]
	}

	return ""
}

// buildTargetURL 构建目标URL
func buildTargetURL(config *proxyconfig.ProxyConfig, r *http.Request) string {
	targetURL := config.TargetURL

	// 添加请求路径
	if r.URL.Path != "/" {
		targetURL = strings.TrimSuffix(targetURL, "/") + r.URL.Path
	}

	// 添加查询参数
	if r.URL.RawQuery != "" {
		if strings.Contains(targetURL, "?") {
			targetURL += "&" + r.URL.RawQuery
		} else {
			targetURL += "?" + r.URL.RawQuery
		}
	}

	return targetURL
}

// validateSubdomainRequest 验证子域名请求
func validateSubdomainRequest(subdomain string) error {
	if subdomain == "" {
		return proxyconfig.ErrInvalidSubdomain
	}

	// 使用现有的验证逻辑
	return proxyconfig.ValidateSubdomain(subdomain)
}
