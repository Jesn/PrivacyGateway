package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

// HandleProxyConfigAPI 处理代理配置API请求
func HandleProxyConfigAPI(w http.ResponseWriter, r *http.Request, cfg *config.Config, log *logger.Logger, storage proxyconfig.Storage) {
	// 认证检查
	if !isAuthorizedForConfig(r, cfg.AdminSecret) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Log-Secret")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetConfigs(w, r, storage, log)
	case http.MethodPost:
		handleCreateConfig(w, r, storage, log)
	case http.MethodPut:
		handleUpdateConfig(w, r, storage, log)
	case http.MethodDelete:
		handleDeleteConfig(w, r, storage, log)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// isAuthorizedForConfig 检查配置管理权限
func isAuthorizedForConfig(r *http.Request, adminSecret string) bool {
	if adminSecret == "" {
		return false
	}

	// 检查请求头
	if secret := r.Header.Get("X-Log-Secret"); secret == adminSecret {
		return true
	}

	// 检查查询参数
	if secret := r.URL.Query().Get("secret"); secret == adminSecret {
		return true
	}

	return false
}

// handleGetConfigs 获取配置列表
func handleGetConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	// 解析查询参数
	filter := &proxyconfig.ConfigFilter{
		Search: r.URL.Query().Get("search"),
		Page:   1,
		Limit:  20,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	if enabledStr := r.URL.Query().Get("enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			filter.Enabled = &enabled
		}
	}

	// 获取配置列表
	response, err := storage.List(filter)
	if err != nil {
		log.Error("failed to get config list", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateConfig 创建配置
func handleCreateConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	var config proxyconfig.ProxyConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证配置
	if err := proxyconfig.ValidateConfig(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 添加配置
	if err := storage.Add(&config); err != nil {
		log.Error("failed to add config", "error", err)
		if err == proxyconfig.ErrDuplicateSubdomain {
			http.Error(w, "Subdomain already exists", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Info("config created", "id", config.ID, "name", config.Name, "subdomain", config.Subdomain)

	// 返回创建的配置
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// handleUpdateConfig 更新配置
func handleUpdateConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	configID := r.URL.Query().Get("id")
	if configID == "" {
		http.Error(w, "Config ID is required", http.StatusBadRequest)
		return
	}

	var config proxyconfig.ProxyConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证配置
	if err := proxyconfig.ValidateConfig(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 更新配置
	if err := storage.Update(configID, &config); err != nil {
		log.Error("failed to update config", "id", configID, "error", err)
		if err == proxyconfig.ErrConfigNotFound {
			http.Error(w, "Config not found", http.StatusNotFound)
		} else if err == proxyconfig.ErrDuplicateSubdomain {
			http.Error(w, "Subdomain already exists", http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Info("config updated", "id", configID, "name", config.Name)

	// 返回更新的配置
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleDeleteConfig 删除配置
func handleDeleteConfig(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	configID := r.URL.Query().Get("id")
	if configID == "" {
		http.Error(w, "Config ID is required", http.StatusBadRequest)
		return
	}

	// 删除配置
	if err := storage.Delete(configID); err != nil {
		log.Error("failed to delete config", "id", configID, "error", err)
		if err == proxyconfig.ErrConfigNotFound {
			http.Error(w, "Config not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Info("config deleted", "id", configID)

	w.WriteHeader(http.StatusNoContent)
}
