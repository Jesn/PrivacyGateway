package proxyconfig

import (
	"fmt"
	"testing"
	"time"
)

// 辅助函数：创建测试配置
func createTestConfig(storage *MemoryStorage, subdomain string) *ProxyConfig {
	config := &ProxyConfig{
		Name:      "Test Config",
		Subdomain: subdomain,
		TargetURL: "https://example.com",
		Enabled:   true,
	}
	storage.Add(config)
	return config
}

func TestMemoryStorage_AddToken(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 创建测试令牌
	token := &AccessToken{
		ID:        "test-token",
		Name:      "Test Token",
		TokenHash: "hash123",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试添加令牌
	err := storage.AddToken(config.ID, token)
	if err != nil {
		t.Fatalf("Failed to add token: %v", err)
	}

	// 验证令牌已添加
	tokens, err := storage.GetTokens(config.ID)
	if err != nil {
		t.Fatalf("Failed to get tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(tokens))
	}
	if tokens[0].ID != "test-token" {
		t.Errorf("Expected token ID 'test-token', got %s", tokens[0].ID)
	}
}

func TestMemoryStorage_AddToken_ConfigNotFound(t *testing.T) {
	storage := NewMemoryStorage(100)
	
	token := &AccessToken{
		ID:        "test-token",
		Name:      "Test Token",
		TokenHash: "hash123",
		Enabled:   true,
	}

	err := storage.AddToken("nonexistent-config", token)
	if err != ErrConfigNotFound {
		t.Errorf("Expected ErrConfigNotFound, got %v", err)
	}
}

func TestMemoryStorage_UpdateToken(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 添加令牌
	token := &AccessToken{
		ID:          "test-token",
		Name:        "Original Name",
		TokenHash:   "hash123",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Description: "Original description",
	}
	storage.AddToken(config.ID, token)

	// 更新令牌
	updatedToken := &AccessToken{
		ID:          "test-token",
		Name:        "Updated Name",
		TokenHash:   "hash123",
		Enabled:     false,
		CreatedAt:   token.CreatedAt,
		UpdatedAt:   time.Now(),
		Description: "Updated description",
	}

	err := storage.UpdateToken(config.ID, "test-token", updatedToken)
	if err != nil {
		t.Fatalf("Failed to update token: %v", err)
	}

	// 验证更新
	retrievedToken, err := storage.GetTokenByID(config.ID, "test-token")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if retrievedToken.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrievedToken.Name)
	}
	if retrievedToken.Enabled {
		t.Error("Expected token to be disabled")
	}
}

func TestMemoryStorage_DeleteToken(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 添加令牌
	token := &AccessToken{
		ID:        "test-token",
		Name:      "Test Token",
		TokenHash: "hash123",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	storage.AddToken(config.ID, token)

	// 删除令牌
	err := storage.DeleteToken(config.ID, "test-token")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// 验证删除
	tokens, err := storage.GetTokens(config.ID)
	if err != nil {
		t.Fatalf("Failed to get tokens: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("Expected 0 tokens after deletion, got %d", len(tokens))
	}
}

func TestMemoryStorage_ValidateToken(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 添加有效令牌
	tokenValue := "test-token-value"
	token := &AccessToken{
		ID:        "test-token",
		Name:      "Test Token",
		TokenHash: HashToken(tokenValue),
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	storage.AddToken(config.ID, token)

	// 测试有效令牌验证
	result, err := storage.ValidateToken(config.ID, tokenValue)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if !result.Valid {
		t.Error("Expected token to be valid")
	}
	if result.Token.ID != "test-token" {
		t.Errorf("Expected token ID 'test-token', got %s", result.Token.ID)
	}

	// 测试无效令牌验证
	result, err = storage.ValidateToken(config.ID, "invalid-token")
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if result.Valid {
		t.Error("Expected token to be invalid")
	}
	if result.ErrorCode != "TOKEN_NOT_FOUND" {
		t.Errorf("Expected error code 'TOKEN_NOT_FOUND', got %s", result.ErrorCode)
	}
}

func TestMemoryStorage_UpdateTokenUsage(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	tokenValue := "test-token-value"
	token := &AccessToken{
		ID:         "test-token",
		Name:       "Test Token",
		TokenHash:  HashToken(tokenValue),
		Enabled:    true,
		UsageCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	storage.AddToken(config.ID, token)

	// 更新使用统计
	err := storage.UpdateTokenUsage(config.ID, tokenValue)
	if err != nil {
		t.Fatalf("Failed to update token usage: %v", err)
	}

	// 验证使用统计更新
	retrievedToken, err := storage.GetTokenByID(config.ID, "test-token")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if retrievedToken.UsageCount != 1 {
		t.Errorf("Expected usage count 1, got %d", retrievedToken.UsageCount)
	}
	if retrievedToken.LastUsed == nil {
		t.Error("Expected LastUsed to be set")
	}
}

func TestMemoryStorage_GetTokenStats(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 添加多个令牌
	now := time.Now()
	tokens := []*AccessToken{
		{
			ID:         "token1",
			Name:       "Token 1",
			TokenHash:  "hash1",
			Enabled:    true,
			UsageCount: 10,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "token2",
			Name:       "Token 2",
			TokenHash:  "hash2",
			Enabled:    false,
			UsageCount: 5,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:        "token3",
			Name:      "Token 3",
			TokenHash: "hash3",
			Enabled:   true,
			ExpiresAt: timePtr(now.Add(-time.Hour)), // 已过期
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, token := range tokens {
		storage.AddToken(config.ID, token)
	}

	// 获取统计信息
	stats, err := storage.GetTokenStats(config.ID)
	if err != nil {
		t.Fatalf("Failed to get token stats: %v", err)
	}

	if stats.TotalTokens != 3 {
		t.Errorf("Expected 3 total tokens, got %d", stats.TotalTokens)
	}
	if stats.ActiveTokens != 1 {
		t.Errorf("Expected 1 active token, got %d", stats.ActiveTokens)
	}
	if stats.TotalRequests != 15 {
		t.Errorf("Expected 15 total requests, got %d", stats.TotalRequests)
	}
}

// 并发安全性测试
func TestMemoryStorage_ConcurrentTokenOperations(t *testing.T) {
	storage := NewMemoryStorage(100)
	config := createTestConfig(storage, "test")

	// 并发添加令牌
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			token := &AccessToken{
				ID:        fmt.Sprintf("token-%d", index),
				Name:      fmt.Sprintf("Token %d", index),
				TokenHash: fmt.Sprintf("hash%d", index),
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			storage.AddToken(config.ID, token)
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证结果
	tokens, err := storage.GetTokens(config.ID)
	if err != nil {
		t.Fatalf("Failed to get tokens: %v", err)
	}
	if len(tokens) != 10 {
		t.Errorf("Expected 10 tokens, got %d", len(tokens))
	}
}
