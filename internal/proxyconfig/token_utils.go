package proxyconfig

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateToken 生成安全的随机令牌
//
// 该函数使用crypto/rand生成32字节的随机数据，然后使用Base64 URL编码。
// 生成的令牌长度固定，避免包含特殊字符，适合在HTTP头部和URL中使用。
//
// 返回值:
//   - string: 生成的令牌字符串
//   - error: 如果随机数生成失败则返回错误
func GenerateToken() (string, error) {
	// 生成32字节的随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// 使用Base64 URL编码，避免特殊字符
	token := base64.URLEncoding.EncodeToString(bytes)

	// 确保令牌长度符合要求
	if len(token) > TokenLength {
		token = token[:TokenLength]
	}

	return token, nil
}

// HashToken 计算令牌的SHA-256哈希值
//
// 该函数使用SHA-256算法对令牌进行哈希处理，用于安全存储。
// 哈希后的值用于令牌验证，避免在存储中保存明文令牌。
//
// 参数:
//   - token: 要哈希的令牌字符串
//
// 返回值:
//   - string: 十六进制编码的哈希值
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// VerifyToken 验证令牌是否匹配哈希值
func VerifyToken(token, hash string) bool {
	return HashToken(token) == hash
}

// CreateAccessToken 创建新的访问令牌
func CreateAccessToken(req *TokenCreateRequest, createdBy string) (*AccessToken, string, error) {
	// 验证请求
	if err := ValidateCreateRequest(req); err != nil {
		return nil, "", err
	}

	// 生成令牌值
	tokenValue, err := GenerateToken()
	if err != nil {
		return nil, "", err
	}

	// 创建令牌对象
	now := time.Now()
	token := &AccessToken{
		ID:          uuid.New().String(),
		Name:        req.Name,
		TokenHash:   HashToken(tokenValue),
		TokenValue:  tokenValue, // 保存令牌值用于复制
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastUsed:    nil,
		UsageCount:  0,
		Enabled:     true,
		CreatedBy:   createdBy,
		Description: req.Description,
	}

	return token, tokenValue, nil
}

// UpdateAccessToken 更新访问令牌
func UpdateAccessToken(token *AccessToken, req *TokenUpdateRequest) error {
	// 验证请求
	if err := ValidateUpdateRequest(req); err != nil {
		return err
	}

	// 更新字段
	if req.Name != "" {
		token.Name = req.Name
	}
	if req.ExpiresAt != nil {
		token.ExpiresAt = req.ExpiresAt
	}
	if req.Description != "" {
		token.Description = req.Description
	}
	if req.Enabled != nil {
		token.Enabled = *req.Enabled
	}

	// 更新时间戳
	token.UpdatedAt = time.Now()

	return nil
}

// ValidateTokenAccess 验证令牌访问权限
func ValidateTokenAccess(token *AccessToken) error {
	if token == nil {
		return ErrTokenNotFound
	}

	if !token.Enabled {
		return ErrTokenDisabled
	}

	if token.IsExpired() {
		return ErrTokenExpired
	}

	return nil
}

// CalculateTokenStats 计算令牌统计信息
func CalculateTokenStats(tokens []AccessToken) *TokenStats {
	stats := &TokenStats{
		TotalTokens:   len(tokens),
		ActiveTokens:  0,
		TotalRequests: 0,
	}

	var lastUsed time.Time

	for _, token := range tokens {
		// 统计活跃令牌
		if token.IsActive() {
			stats.ActiveTokens++
		}

		// 累计请求数
		stats.TotalRequests += token.UsageCount

		// 找到最后使用时间
		if token.LastUsed != nil && token.LastUsed.After(lastUsed) {
			lastUsed = *token.LastUsed
		}
	}

	stats.LastTokenUsed = lastUsed

	return stats
}

// FilterActiveTokens 过滤活跃令牌
func FilterActiveTokens(tokens []AccessToken) []AccessToken {
	var activeTokens []AccessToken
	for _, token := range tokens {
		if token.IsActive() {
			activeTokens = append(activeTokens, token)
		}
	}
	return activeTokens
}

// FilterTokensByStatus 按状态过滤令牌
func FilterTokensByStatus(tokens []AccessToken, status string) []AccessToken {
	var filteredTokens []AccessToken
	for _, token := range tokens {
		if token.GetStatus() == status {
			filteredTokens = append(filteredTokens, token)
		}
	}
	return filteredTokens
}

// FindTokenByID 根据ID查找令牌
func FindTokenByID(tokens []AccessToken, tokenID string) (*AccessToken, int) {
	for i, token := range tokens {
		if token.ID == tokenID {
			return &tokens[i], i
		}
	}
	return nil, -1
}

// FindTokenByHash 根据哈希值查找令牌
func FindTokenByHash(tokens []AccessToken, tokenHash string) *AccessToken {
	for _, token := range tokens {
		if token.TokenHash == tokenHash {
			return &token
		}
	}
	return nil
}

// CleanupExpiredTokens 清理过期令牌
func CleanupExpiredTokens(tokens []AccessToken) []AccessToken {
	var validTokens []AccessToken
	for _, token := range tokens {
		if !token.IsExpired() {
			validTokens = append(validTokens, token)
		}
	}
	return validTokens
}

// ValidateTokenLimit 验证令牌数量限制
func ValidateTokenLimit(currentCount int) error {
	if currentCount >= MaxTokensPerConfig {
		return ErrMaxTokensExceeded
	}
	return nil
}

// SanitizeTokenForResponse 清理令牌数据用于响应（移除敏感信息）
func SanitizeTokenForResponse(token *AccessToken) AccessToken {
	sanitized := *token
	// 不返回令牌哈希值，但保留令牌值用于复制
	sanitized.TokenHash = ""
	return sanitized
}

// SanitizeTokensForResponse 批量清理令牌数据
func SanitizeTokensForResponse(tokens []AccessToken) []AccessToken {
	sanitized := make([]AccessToken, len(tokens))
	for i, token := range tokens {
		sanitized[i] = SanitizeTokenForResponse(&token)
	}
	return sanitized
}

// CreateTokenUsageRecord 创建令牌使用记录
func CreateTokenUsageRecord(tokenID, configID string, success bool, errorMsg, userAgent, remoteIP string) *TokenUsageRecord {
	return &TokenUsageRecord{
		TokenID:     tokenID,
		ConfigID:    configID,
		RequestTime: time.Now(),
		Success:     success,
		ErrorMsg:    errorMsg,
		UserAgent:   userAgent,
		RemoteIP:    remoteIP,
	}
}

// IsValidTokenFormat 检查令牌格式是否有效
func IsValidTokenFormat(token string) bool {
	// 检查长度
	if len(token) < 32 || len(token) > TokenLength {
		return false
	}

	// 检查是否为有效的Base64 URL编码
	_, err := base64.URLEncoding.DecodeString(token)
	return err == nil
}
