package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"privacygateway/internal/logger"
	"privacygateway/internal/proxyconfig"
)

// TokenAPIHandler 令牌管理API处理器
type TokenAPIHandler struct {
	storage       proxyconfig.Storage
	authenticator *ProxyAuthenticator
	logger        *logger.Logger
}

// NewTokenAPIHandler 创建令牌API处理器
func NewTokenAPIHandler(storage proxyconfig.Storage, adminSecret string, logger *logger.Logger) *TokenAPIHandler {
	return &TokenAPIHandler{
		storage:       storage,
		authenticator: NewProxyAuthenticator(adminSecret, storage, logger),
		logger:        logger,
	}
}

// APIResponse 标准API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Status  int         `json:"status"`
}

// TokenListAPIResponse 令牌列表API响应
type TokenListAPIResponse struct {
	Success bool                           `json:"success"`
	Data    *proxyconfig.TokenListResponse `json:"data,omitempty"`
	Error   string                         `json:"error,omitempty"`
	Status  int                            `json:"status"`
}

// TokenAPIResponse 单个令牌API响应
type TokenAPIResponse struct {
	Success bool                       `json:"success"`
	Data    *proxyconfig.TokenResponse `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
	Status  int                        `json:"status"`
}

// HandleTokenAPI 处理令牌API请求的主路由
func (h *TokenAPIHandler) HandleTokenAPI(w http.ResponseWriter, r *http.Request) {
	// 注意：CORS头部已在路由层设置，这里不再重复设置
	w.Header().Set("Content-Type", "application/json")

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 认证检查（仅支持管理员）
	authResult := h.authenticator.AuthenticateForConfig(r)
	if !authResult.Authenticated {
		h.logger.Warn("token API access denied",
			"client_ip", getClientIP(r),
			"path", r.URL.Path,
			"method", r.Method,
			"error", authResult.Error)

		h.sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 提取配置ID
	configID := h.extractConfigIDFromPath(r.URL.Path)
	if configID == "" {
		h.sendErrorResponse(w, "Configuration ID is required", http.StatusBadRequest)
		return
	}

	// 验证配置是否存在
	_, err := h.storage.GetByID(configID)
	if err != nil {
		h.logger.Debug("config not found for token API", "config_id", configID, "error", err)
		h.sendErrorResponse(w, "Configuration not found", http.StatusNotFound)
		return
	}

	// 记录API访问
	h.logger.Info("token API access",
		"config_id", configID,
		"method", r.Method,
		"path", r.URL.Path,
		"client_ip", getClientIP(r))

	// 路由到具体的处理方法
	switch r.Method {
	case http.MethodGet:
		if strings.HasSuffix(r.URL.Path, "/tokens") {
			h.handleListTokens(w, r, configID)
		} else {
			tokenID := h.extractTokenIDFromPath(r.URL.Path)
			if tokenID == "" {
				h.sendErrorResponse(w, "Token ID is required", http.StatusBadRequest)
				return
			}
			h.handleGetToken(w, r, configID, tokenID)
		}
	case http.MethodPost:
		h.handleCreateToken(w, r, configID)
	case http.MethodPut:
		tokenID := h.extractTokenIDFromPath(r.URL.Path)
		if tokenID == "" {
			h.sendErrorResponse(w, "Token ID is required", http.StatusBadRequest)
			return
		}
		h.handleUpdateToken(w, r, configID, tokenID)
	case http.MethodDelete:
		tokenID := h.extractTokenIDFromPath(r.URL.Path)
		if tokenID == "" {
			h.sendErrorResponse(w, "Token ID is required", http.StatusBadRequest)
			return
		}
		h.handleDeleteToken(w, r, configID, tokenID)
	default:
		h.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListTokens 处理获取令牌列表请求
func (h *TokenAPIHandler) handleListTokens(w http.ResponseWriter, r *http.Request, configID string) {
	tokens, err := h.storage.GetTokens(configID)
	if err != nil {
		h.logger.Error("failed to get tokens", "config_id", configID, "error", err)
		h.sendErrorResponse(w, "Failed to retrieve tokens", http.StatusInternalServerError)
		return
	}

	// 获取令牌统计
	stats, err := h.storage.GetTokenStats(configID)
	if err != nil {
		h.logger.Error("failed to get token stats", "config_id", configID, "error", err)
		// 继续处理，但不包含统计信息
		stats = nil
	}

	// 清理敏感信息
	sanitizedTokens := proxyconfig.SanitizeTokensForResponse(tokens)

	response := &TokenListAPIResponse{
		Success: true,
		Data: &proxyconfig.TokenListResponse{
			Tokens: sanitizedTokens,
			Stats:  stats,
		},
		Status: http.StatusOK,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// handleGetToken 处理获取单个令牌请求
func (h *TokenAPIHandler) handleGetToken(w http.ResponseWriter, r *http.Request, configID, tokenID string) {
	token, err := h.storage.GetTokenByID(configID, tokenID)
	if err != nil {
		if err == proxyconfig.ErrTokenNotFound {
			h.sendErrorResponse(w, "Token not found", http.StatusNotFound)
		} else {
			h.logger.Error("failed to get token", "config_id", configID, "token_id", tokenID, "error", err)
			h.sendErrorResponse(w, "Failed to retrieve token", http.StatusInternalServerError)
		}
		return
	}

	// 清理敏感信息
	sanitizedToken := proxyconfig.SanitizeTokenForResponse(token)

	response := &TokenAPIResponse{
		Success: true,
		Data: &proxyconfig.TokenResponse{
			AccessToken: sanitizedToken,
		},
		Status: http.StatusOK,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// handleCreateToken 处理创建令牌请求
func (h *TokenAPIHandler) handleCreateToken(w http.ResponseWriter, r *http.Request, configID string) {
	var req proxyconfig.TokenCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证请求
	if err := proxyconfig.ValidateCreateRequest(&req); err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 创建令牌
	token, tokenValue, err := proxyconfig.CreateAccessToken(&req, "admin")
	if err != nil {
		h.logger.Error("failed to create token", "config_id", configID, "error", err)
		h.sendErrorResponse(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	// 保存令牌
	if err := h.storage.AddToken(configID, token); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.sendErrorResponse(w, "Token name already exists", http.StatusConflict)
		} else if err == proxyconfig.ErrMaxTokensExceeded {
			h.sendErrorResponse(w, "Maximum tokens per configuration exceeded", http.StatusBadRequest)
		} else {
			h.logger.Error("failed to save token", "config_id", configID, "error", err)
			h.sendErrorResponse(w, "Failed to save token", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("token created",
		"config_id", configID,
		"token_id", token.ID,
		"token_name", token.Name,
		"client_ip", getClientIP(r))

	// 返回令牌（包含明文值，仅此一次）
	response := &TokenAPIResponse{
		Success: true,
		Data: &proxyconfig.TokenResponse{
			AccessToken: *token,
			Token:       tokenValue, // 明文令牌值
		},
		Status: http.StatusCreated,
	}

	h.sendJSONResponse(w, response, http.StatusCreated)
}

// handleUpdateToken 处理更新令牌请求
func (h *TokenAPIHandler) handleUpdateToken(w http.ResponseWriter, r *http.Request, configID, tokenID string) {
	var req proxyconfig.TokenUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证请求
	if err := proxyconfig.ValidateUpdateRequest(&req); err != nil {
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取现有令牌
	existingToken, err := h.storage.GetTokenByID(configID, tokenID)
	if err != nil {
		if err == proxyconfig.ErrTokenNotFound {
			h.sendErrorResponse(w, "Token not found", http.StatusNotFound)
		} else {
			h.logger.Error("failed to get existing token", "config_id", configID, "token_id", tokenID, "error", err)
			h.sendErrorResponse(w, "Failed to retrieve token", http.StatusInternalServerError)
		}
		return
	}

	// 更新令牌
	if err := proxyconfig.UpdateAccessToken(existingToken, &req); err != nil {
		h.logger.Error("failed to update token", "config_id", configID, "token_id", tokenID, "error", err)
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 保存更新
	if err := h.storage.UpdateToken(configID, tokenID, existingToken); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.sendErrorResponse(w, "Token name already exists", http.StatusConflict)
		} else {
			h.logger.Error("failed to save updated token", "config_id", configID, "token_id", tokenID, "error", err)
			h.sendErrorResponse(w, "Failed to save token", http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("token updated",
		"config_id", configID,
		"token_id", tokenID,
		"token_name", existingToken.Name,
		"client_ip", getClientIP(r))

	// 清理敏感信息并返回
	sanitizedToken := proxyconfig.SanitizeTokenForResponse(existingToken)

	response := &TokenAPIResponse{
		Success: true,
		Data: &proxyconfig.TokenResponse{
			AccessToken: sanitizedToken,
		},
		Status: http.StatusOK,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// handleDeleteToken 处理删除令牌请求
func (h *TokenAPIHandler) handleDeleteToken(w http.ResponseWriter, r *http.Request, configID, tokenID string) {
	// 检查令牌是否存在
	token, err := h.storage.GetTokenByID(configID, tokenID)
	if err != nil {
		if err == proxyconfig.ErrTokenNotFound {
			h.sendErrorResponse(w, "Token not found", http.StatusNotFound)
		} else {
			h.logger.Error("failed to get token for deletion", "config_id", configID, "token_id", tokenID, "error", err)
			h.sendErrorResponse(w, "Failed to retrieve token", http.StatusInternalServerError)
		}
		return
	}

	// 删除令牌
	if err := h.storage.DeleteToken(configID, tokenID); err != nil {
		h.logger.Error("failed to delete token", "config_id", configID, "token_id", tokenID, "error", err)
		h.sendErrorResponse(w, "Failed to delete token", http.StatusInternalServerError)
		return
	}

	h.logger.Info("token deleted",
		"config_id", configID,
		"token_id", tokenID,
		"token_name", token.Name,
		"client_ip", getClientIP(r))

	response := &APIResponse{
		Success: true,
		Message: "Token deleted successfully",
		Status:  http.StatusOK,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// 辅助方法

// extractConfigIDFromPath 从URL路径提取配置ID
func (h *TokenAPIHandler) extractConfigIDFromPath(path string) string {
	// 路径格式: /config/proxy/{configID}/tokens[/{tokenID}]
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "config" && parts[1] == "proxy" {
		return parts[2]
	}
	return ""
}

// extractTokenIDFromPath 从URL路径提取令牌ID
func (h *TokenAPIHandler) extractTokenIDFromPath(path string) string {
	// 路径格式: /config/proxy/{configID}/tokens/{tokenID}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 5 && parts[0] == "config" && parts[1] == "proxy" && parts[3] == "tokens" {
		return parts[4]
	}
	return ""
}

// sendErrorResponse 发送错误响应
func (h *TokenAPIHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := &APIResponse{
		Success: false,
		Error:   message,
		Status:  statusCode,
	}
	h.sendJSONResponse(w, response, statusCode)
}

// sendJSONResponse 发送JSON响应
func (h *TokenAPIHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}
