package proxyconfig

import (
	"errors"
	"time"
)

// AccessToken 访问令牌结构
type AccessToken struct {
	ID          string     `json:"id"`                    // 令牌唯一标识
	Name        string     `json:"name"`                  // 令牌名称
	TokenHash   string     `json:"token_hash"`            // 令牌哈希值(不存储明文)
	TokenValue  string     `json:"token_value,omitempty"` // 令牌值(用于复制)
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`  // 过期时间
	CreatedAt   time.Time  `json:"created_at"`            // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`            // 更新时间
	LastUsed    *time.Time `json:"last_used,omitempty"`   // 最后使用时间
	UsageCount  int64      `json:"usage_count"`           // 使用次数
	Enabled     bool       `json:"enabled"`               // 是否启用
	CreatedBy   string     `json:"created_by,omitempty"`  // 创建者
	Description string     `json:"description,omitempty"` // 描述信息
}

// TokenStats 令牌统计信息
type TokenStats struct {
	TotalTokens   int       `json:"total_tokens"`    // 总令牌数
	ActiveTokens  int       `json:"active_tokens"`   // 活跃令牌数
	TotalRequests int64     `json:"total_requests"`  // 总请求数
	LastTokenUsed time.Time `json:"last_token_used"` // 最后令牌使用时间
}

// TokenCreateRequest 创建令牌请求
type TokenCreateRequest struct {
	Name        string     `json:"name"`                  // 令牌名称
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`  // 过期时间
	Description string     `json:"description,omitempty"` // 描述信息
}

// TokenUpdateRequest 更新令牌请求
type TokenUpdateRequest struct {
	Name        string     `json:"name,omitempty"`        // 令牌名称
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`  // 过期时间
	Description string     `json:"description,omitempty"` // 描述信息
	Enabled     *bool      `json:"enabled,omitempty"`     // 是否启用
}

// TokenResponse 令牌响应（包含明文令牌，仅在创建时返回）
type TokenResponse struct {
	AccessToken
	Token string `json:"token,omitempty"` // 明文令牌值（仅在创建时返回）
}

// TokenListResponse 令牌列表响应
type TokenListResponse struct {
	Tokens []AccessToken `json:"tokens"`
	Stats  *TokenStats   `json:"stats,omitempty"`
}

// TokenUsageRecord 令牌使用记录
type TokenUsageRecord struct {
	TokenID     string    `json:"token_id"`
	ConfigID    string    `json:"config_id"`
	RequestTime time.Time `json:"request_time"`
	Success     bool      `json:"success"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	RemoteIP    string    `json:"remote_ip,omitempty"`
}

// TokenValidationResult 令牌验证结果
type TokenValidationResult struct {
	Valid     bool         `json:"valid"`
	Token     *AccessToken `json:"token,omitempty"`
	ConfigID  string       `json:"config_id,omitempty"`
	ErrorCode string       `json:"error_code,omitempty"`
	ErrorMsg  string       `json:"error_msg,omitempty"`
}

// 令牌相关错误定义
var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenDisabled     = errors.New("token disabled")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrTokenNameRequired = errors.New("token name is required")
	ErrTokenNameTooLong  = errors.New("token name too long (max 100 characters)")
	ErrDuplicateToken    = errors.New("token already exists")
	ErrMaxTokensExceeded = errors.New("maximum tokens per config exceeded")
)

// 令牌状态常量
const (
	TokenStatusActive   = "active"   // 活跃
	TokenStatusDisabled = "disabled" // 已禁用
	TokenStatusExpired  = "expired"  // 已过期
)

// 令牌相关配置常量
const (
	MaxTokensPerConfig = 50                   // 每个配置最大令牌数
	TokenLength        = 64                   // 令牌长度
	DefaultTokenTTL    = 365 * 24 * time.Hour // 默认令牌有效期（1年）
	MaxTokenNameLength = 100                  // 令牌名称最大长度
)

// IsExpired 检查令牌是否过期
func (t *AccessToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// IsActive 检查令牌是否活跃
func (t *AccessToken) IsActive() bool {
	return t.Enabled && !t.IsExpired()
}

// GetStatus 获取令牌状态
func (t *AccessToken) GetStatus() string {
	if !t.Enabled {
		return TokenStatusDisabled
	}
	if t.IsExpired() {
		return TokenStatusExpired
	}
	return TokenStatusActive
}

// UpdateUsage 更新令牌使用统计
func (t *AccessToken) UpdateUsage() {
	now := time.Now()
	t.LastUsed = &now
	t.UsageCount++
	t.UpdatedAt = now
}

// Validate 验证令牌数据
func (t *AccessToken) Validate() error {
	if t.Name == "" {
		return ErrTokenNameRequired
	}
	if len(t.Name) > MaxTokenNameLength {
		return ErrTokenNameTooLong
	}
	if t.TokenHash == "" {
		return ErrTokenInvalid
	}
	return nil
}

// ValidateCreateRequest 验证创建令牌请求
func ValidateCreateRequest(req *TokenCreateRequest) error {
	if req.Name == "" {
		return ErrTokenNameRequired
	}
	if len(req.Name) > MaxTokenNameLength {
		return ErrTokenNameTooLong
	}
	// 验证过期时间不能是过去的时间
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration time cannot be in the past")
	}
	return nil
}

// ValidateUpdateRequest 验证更新令牌请求
func ValidateUpdateRequest(req *TokenUpdateRequest) error {
	if req.Name != "" && len(req.Name) > MaxTokenNameLength {
		return ErrTokenNameTooLong
	}
	// 验证过期时间不能是过去的时间
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration time cannot be in the past")
	}
	return nil
}
