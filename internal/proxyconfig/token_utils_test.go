package proxyconfig

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	// 测试令牌生成
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 验证令牌不同
	if token1 == token2 {
		t.Error("Generated tokens should be different")
	}

	// 验证令牌长度
	if len(token1) == 0 || len(token1) > TokenLength {
		t.Errorf("Token length should be between 1 and %d, got %d", TokenLength, len(token1))
	}

	// 验证令牌格式
	if !IsValidTokenFormat(token1) {
		t.Error("Generated token should have valid format")
	}
}

func TestTokenUniqueness(t *testing.T) {
	// 生成多个令牌，验证唯一性
	tokens := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		token, err := GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}
		if tokens[token] {
			t.Errorf("Duplicate token generated: %s", token)
		}
		tokens[token] = true
	}
}

func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	// 相同输入应该产生相同哈希
	if hash1 != hash2 {
		t.Error("Same token should produce same hash")
	}

	// 不同输入应该产生不同哈希
	differentToken := "different-token"
	hash3 := HashToken(differentToken)
	if hash1 == hash3 {
		t.Error("Different tokens should produce different hashes")
	}

	// 验证哈希长度（SHA-256 hex编码应该是64字符）
	if len(hash1) != 64 {
		t.Errorf("Hash length should be 64, got %d", len(hash1))
	}
}

func TestVerifyToken(t *testing.T) {
	token := "test-token-123"
	hash := HashToken(token)

	// 正确的令牌应该验证成功
	if !VerifyToken(token, hash) {
		t.Error("Valid token should verify successfully")
	}

	// 错误的令牌应该验证失败
	if VerifyToken("wrong-token", hash) {
		t.Error("Invalid token should fail verification")
	}
}

func TestCreateAccessToken(t *testing.T) {
	req := &TokenCreateRequest{
		Name:        "test-token",
		Description: "test description",
	}

	token, tokenValue, err := CreateAccessToken(req, "admin")
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	// 验证令牌对象
	if token.Name != req.Name {
		t.Errorf("Token name = %v, want %v", token.Name, req.Name)
	}
	if token.Description != req.Description {
		t.Errorf("Token description = %v, want %v", token.Description, req.Description)
	}
	if token.CreatedBy != "admin" {
		t.Errorf("Token created by = %v, want admin", token.CreatedBy)
	}
	if !token.Enabled {
		t.Error("Token should be enabled by default")
	}
	if token.UsageCount != 0 {
		t.Error("Token usage count should be 0 initially")
	}

	// 验证令牌值
	if tokenValue == "" {
		t.Error("Token value should not be empty")
	}
	if !VerifyToken(tokenValue, token.TokenHash) {
		t.Error("Token value should match hash")
	}

	// 验证ID生成
	if token.ID == "" {
		t.Error("Token ID should be generated")
	}
}

func TestUpdateAccessToken(t *testing.T) {
	// 创建初始令牌
	token := &AccessToken{
		Name:        "original-name",
		Description: "original-description",
		Enabled:     true,
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	// 更新请求
	req := &TokenUpdateRequest{
		Name:        "updated-name",
		Description: "updated-description",
		Enabled:     boolPtr(false),
	}

	beforeUpdate := time.Now()
	err := UpdateAccessToken(token, req)
	afterUpdate := time.Now()

	if err != nil {
		t.Fatalf("UpdateAccessToken() error = %v", err)
	}

	// 验证更新
	if token.Name != req.Name {
		t.Errorf("Token name = %v, want %v", token.Name, req.Name)
	}
	if token.Description != req.Description {
		t.Errorf("Token description = %v, want %v", token.Description, req.Description)
	}
	if token.Enabled != *req.Enabled {
		t.Errorf("Token enabled = %v, want %v", token.Enabled, *req.Enabled)
	}

	// 验证更新时间
	if token.UpdatedAt.Before(beforeUpdate) || token.UpdatedAt.After(afterUpdate) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestValidateTokenAccess(t *testing.T) {
	tests := []struct {
		name    string
		token   *AccessToken
		wantErr error
	}{
		{
			name:    "nil token",
			token:   nil,
			wantErr: ErrTokenNotFound,
		},
		{
			name: "disabled token",
			token: &AccessToken{
				Enabled: false,
			},
			wantErr: ErrTokenDisabled,
		},
		{
			name: "expired token",
			token: &AccessToken{
				Enabled:   true,
				ExpiresAt: timePtr(time.Now().Add(-time.Hour)),
			},
			wantErr: ErrTokenExpired,
		},
		{
			name: "valid token",
			token: &AccessToken{
				Enabled:   true,
				ExpiresAt: timePtr(time.Now().Add(time.Hour)),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenAccess(tt.token)
			if err != tt.wantErr {
				t.Errorf("ValidateTokenAccess() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateTokenStats(t *testing.T) {
	now := time.Now()
	tokens := []AccessToken{
		{
			Enabled:    true,
			ExpiresAt:  timePtr(now.Add(time.Hour)),
			UsageCount: 10,
			LastUsed:   timePtr(now.Add(-time.Minute)),
		},
		{
			Enabled:    false,
			ExpiresAt:  timePtr(now.Add(time.Hour)),
			UsageCount: 5,
			LastUsed:   timePtr(now.Add(-time.Hour)),
		},
		{
			Enabled:    true,
			ExpiresAt:  timePtr(now.Add(-time.Hour)), // expired
			UsageCount: 3,
			LastUsed:   timePtr(now.Add(-2 * time.Hour)),
		},
	}

	stats := CalculateTokenStats(tokens)

	if stats.TotalTokens != 3 {
		t.Errorf("TotalTokens = %v, want 3", stats.TotalTokens)
	}
	if stats.ActiveTokens != 1 {
		t.Errorf("ActiveTokens = %v, want 1", stats.ActiveTokens)
	}
	if stats.TotalRequests != 18 {
		t.Errorf("TotalRequests = %v, want 18", stats.TotalRequests)
	}
}

func TestFilterActiveTokens(t *testing.T) {
	now := time.Now()
	tokens := []AccessToken{
		{Enabled: true, ExpiresAt: timePtr(now.Add(time.Hour))},
		{Enabled: false, ExpiresAt: timePtr(now.Add(time.Hour))},
		{Enabled: true, ExpiresAt: timePtr(now.Add(-time.Hour))},
	}

	activeTokens := FilterActiveTokens(tokens)

	if len(activeTokens) != 1 {
		t.Errorf("Expected 1 active token, got %d", len(activeTokens))
	}
}

func TestFindTokenByID(t *testing.T) {
	tokens := []AccessToken{
		{ID: "token1"},
		{ID: "token2"},
		{ID: "token3"},
	}

	token, index := FindTokenByID(tokens, "token2")
	if token == nil {
		t.Error("Expected to find token")
	}
	if index != 1 {
		t.Errorf("Expected index 1, got %d", index)
	}
	if token.ID != "token2" {
		t.Errorf("Expected token2, got %s", token.ID)
	}

	// 测试未找到的情况
	token, index = FindTokenByID(tokens, "nonexistent")
	if token != nil {
		t.Error("Expected not to find token")
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}
}

func TestIsValidTokenFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid token",
			token: "abcdefghijklmnopqrstuvwxyzABCDEF",
			want:  true,
		},
		{
			name:  "too short",
			token: "short",
			want:  false,
		},
		{
			name:  "too long",
			token: strings.Repeat("a", TokenLength+1),
			want:  false,
		},
		{
			name:  "empty token",
			token: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTokenFormat(tt.token); got != tt.want {
				t.Errorf("IsValidTokenFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterTokensByStatus(t *testing.T) {
	now := time.Now()
	tokens := []AccessToken{
		{Enabled: true, ExpiresAt: timePtr(now.Add(time.Hour))},  // active
		{Enabled: false, ExpiresAt: timePtr(now.Add(time.Hour))}, // disabled
		{Enabled: true, ExpiresAt: timePtr(now.Add(-time.Hour))}, // expired
	}

	activeTokens := FilterTokensByStatus(tokens, TokenStatusActive)
	if len(activeTokens) != 1 {
		t.Errorf("Expected 1 active token, got %d", len(activeTokens))
	}

	disabledTokens := FilterTokensByStatus(tokens, TokenStatusDisabled)
	if len(disabledTokens) != 1 {
		t.Errorf("Expected 1 disabled token, got %d", len(disabledTokens))
	}

	expiredTokens := FilterTokensByStatus(tokens, TokenStatusExpired)
	if len(expiredTokens) != 1 {
		t.Errorf("Expected 1 expired token, got %d", len(expiredTokens))
	}
}

func TestFindTokenByHash(t *testing.T) {
	tokens := []AccessToken{
		{TokenHash: "hash1"},
		{TokenHash: "hash2"},
		{TokenHash: "hash3"},
	}

	token := FindTokenByHash(tokens, "hash2")
	if token == nil {
		t.Error("Expected to find token")
	}
	if token.TokenHash != "hash2" {
		t.Errorf("Expected hash2, got %s", token.TokenHash)
	}

	// 测试未找到的情况
	token = FindTokenByHash(tokens, "nonexistent")
	if token != nil {
		t.Error("Expected not to find token")
	}
}

func TestCleanupExpiredTokens(t *testing.T) {
	now := time.Now()
	tokens := []AccessToken{
		{ExpiresAt: timePtr(now.Add(time.Hour))},  // valid
		{ExpiresAt: timePtr(now.Add(-time.Hour))}, // expired
		{ExpiresAt: nil}, // no expiration
	}

	validTokens := CleanupExpiredTokens(tokens)
	if len(validTokens) != 2 {
		t.Errorf("Expected 2 valid tokens, got %d", len(validTokens))
	}
}

func TestValidateTokenLimit(t *testing.T) {
	// 测试未超过限制
	err := ValidateTokenLimit(MaxTokensPerConfig - 1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 测试超过限制
	err = ValidateTokenLimit(MaxTokensPerConfig)
	if err != ErrMaxTokensExceeded {
		t.Errorf("Expected ErrMaxTokensExceeded, got %v", err)
	}
}

func TestSanitizeTokenForResponse(t *testing.T) {
	token := &AccessToken{
		ID:        "test-id",
		Name:      "test-token",
		TokenHash: "secret-hash",
	}

	sanitized := SanitizeTokenForResponse(token)
	if sanitized.TokenHash != "" {
		t.Error("TokenHash should be empty in sanitized response")
	}
	if sanitized.ID != token.ID {
		t.Error("Other fields should be preserved")
	}
}

func TestSanitizeTokensForResponse(t *testing.T) {
	tokens := []AccessToken{
		{ID: "1", TokenHash: "hash1"},
		{ID: "2", TokenHash: "hash2"},
	}

	sanitized := SanitizeTokensForResponse(tokens)
	if len(sanitized) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(sanitized))
	}
	for _, token := range sanitized {
		if token.TokenHash != "" {
			t.Error("TokenHash should be empty in sanitized response")
		}
	}
}

func TestCreateTokenUsageRecord(t *testing.T) {
	record := CreateTokenUsageRecord("token1", "config1", true, "", "test-agent", "127.0.0.1")

	if record.TokenID != "token1" {
		t.Errorf("Expected token1, got %s", record.TokenID)
	}
	if record.ConfigID != "config1" {
		t.Errorf("Expected config1, got %s", record.ConfigID)
	}
	if !record.Success {
		t.Error("Expected success to be true")
	}
	if record.UserAgent != "test-agent" {
		t.Errorf("Expected test-agent, got %s", record.UserAgent)
	}
}

func TestValidateCreateRequest_EdgeCases(t *testing.T) {
	// 测试过期时间验证的边界情况
	req := &TokenCreateRequest{
		Name:      "test",
		ExpiresAt: timePtr(time.Now().Add(-time.Nanosecond)), // 刚刚过期
	}

	err := ValidateCreateRequest(req)
	if err == nil {
		t.Error("Expected error for past expiration time")
	}
}

func TestValidateUpdateRequest_EdgeCases(t *testing.T) {
	// 测试过期时间验证的边界情况
	req := &TokenUpdateRequest{
		ExpiresAt: timePtr(time.Now().Add(-time.Nanosecond)), // 刚刚过期
	}

	err := ValidateUpdateRequest(req)
	if err == nil {
		t.Error("Expected error for past expiration time")
	}
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}
