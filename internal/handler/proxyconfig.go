package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	// 检查特殊路径
	path := r.URL.Path
	if path == "/config/proxy/export" {
		handleExportConfigs(w, r, storage, log)
		return
	}
	if path == "/config/proxy/import" {
		handleImportConfigs(w, r, storage, log)
		return
	}
	if path == "/config/proxy/batch" {
		handleBatchOperation(w, r, storage, log)
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

// handleExportConfigs 导出配置
func handleExportConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	exportData, err := storage.ExportAll()
	if err != nil {
		log.Error("failed to export configs", "error", err)
		http.Error(w, "Export failed", http.StatusInternalServerError)
		return
	}

	// 设置下载文件头
	filename := fmt.Sprintf("proxy-configs-%s.json", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	json.NewEncoder(w).Encode(exportData)
	log.Info("configs exported", "count", exportData.TotalCount, "filename", filename)
}

// handleImportConfigs 导入配置
func handleImportConfigs(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var importData struct {
		Configs []proxyconfig.ProxyConfig `json:"configs"`
		Mode    string                    `json:"mode"` // skip, replace, error
	}

	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 默认模式为error（遇到冲突时报错）
	if importData.Mode == "" {
		importData.Mode = "error"
	}

	result, err := storage.ImportConfigs(importData.Configs, importData.Mode)
	if err != nil {
		log.Error("failed to import configs", "error", err)
		http.Error(w, "Import failed", http.StatusInternalServerError)
		return
	}

	log.Info("configs imported", "imported", result.ImportedCount, "skipped", result.SkippedCount, "errors", result.ErrorCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleBatchOperation 批量操作
func handleBatchOperation(w http.ResponseWriter, r *http.Request, storage proxyconfig.Storage, log *logger.Logger) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proxyconfig.BatchOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 验证操作类型
	if req.Operation != "enable" && req.Operation != "disable" && req.Operation != "delete" {
		http.Error(w, "Invalid operation. Must be: enable, disable, or delete", http.StatusBadRequest)
		return
	}

	if len(req.ConfigIDs) == 0 {
		http.Error(w, "No config IDs provided", http.StatusBadRequest)
		return
	}

	result, err := storage.BatchOperation(req.Operation, req.ConfigIDs)
	if err != nil {
		log.Error("batch operation failed", "operation", req.Operation, "error", err)
		http.Error(w, "Batch operation failed", http.StatusInternalServerError)
		return
	}

	log.Info("batch operation completed", "operation", req.Operation, "total", result.TotalCount, "success", len(result.Success), "failed", result.FailedCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
