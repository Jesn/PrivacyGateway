package proxyconfig

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMigrateConfigData_LegacyToNew(t *testing.T) {
	// 创建旧版本配置数据
	legacyConfigs := map[string]*LegacyProxyConfig{
		"config1": {
			ID:        "config1",
			Name:      "Test Config 1",
			Subdomain: "test1",
			TargetURL: "https://example1.com",
			Protocol:  "https",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Stats: &ConfigStats{
				RequestCount: 100,
				ErrorCount:   5,
			},
		},
		"config2": {
			ID:        "config2",
			Name:      "Test Config 2",
			Subdomain: "test2",
			TargetURL: "https://example2.com",
			Protocol:  "https",
			Enabled:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// 序列化旧版本数据
	legacyData, err := json.Marshal(legacyConfigs)
	if err != nil {
		t.Fatalf("Failed to marshal legacy data: %v", err)
	}

	// 执行迁移
	result, migratedData, err := MigrateConfigData(legacyData)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 验证迁移结果
	if !result.Success {
		t.Errorf("Migration should succeed, got errors: %v", result.Errors)
	}
	if result.MigratedCount != 2 {
		t.Errorf("Expected 2 migrated configs, got %d", result.MigratedCount)
	}
	if result.FromVersion != DataVersionV1_0 {
		t.Errorf("Expected from version %s, got %s", DataVersionV1_0, result.FromVersion)
	}
	if result.ToVersion != CurrentVersion {
		t.Errorf("Expected to version %s, got %s", CurrentVersion, result.ToVersion)
	}

	// 验证迁移后的数据
	var migratedConfigs map[string]*ProxyConfig
	err = json.Unmarshal(migratedData, &migratedConfigs)
	if err != nil {
		t.Fatalf("Failed to unmarshal migrated data: %v", err)
	}

	if len(migratedConfigs) != 2 {
		t.Errorf("Expected 2 migrated configs, got %d", len(migratedConfigs))
	}

	// 验证具体配置
	config1 := migratedConfigs["config1"]
	if config1 == nil {
		t.Fatal("Config1 should exist after migration")
	}
	if config1.Name != "Test Config 1" {
		t.Errorf("Expected name 'Test Config 1', got %s", config1.Name)
	}

	// 检查AccessTokens字段（nil slice和空slice都是可接受的）
	if len(config1.AccessTokens) != 0 {
		t.Errorf("Expected 0 access tokens, got %d", len(config1.AccessTokens))
	}

	// 检查TokenStats字段
	if config1.TokenStats == nil {
		t.Error("TokenStats should be initialized")
	}
}

func TestMigrateConfigData_AlreadyNewVersion(t *testing.T) {
	// 创建新版本配置数据
	newConfigs := map[string]*ProxyConfig{
		"config1": {
			ID:           "config1",
			Name:         "Test Config 1",
			Subdomain:    "test1",
			TargetURL:    "https://example1.com",
			Protocol:     "https",
			Enabled:      true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			AccessTokens: []AccessToken{},
			TokenStats:   &TokenStats{},
		},
	}

	// 序列化新版本数据
	newData, err := json.Marshal(newConfigs)
	if err != nil {
		t.Fatalf("Failed to marshal new data: %v", err)
	}

	// 执行迁移
	result, migratedData, err := MigrateConfigData(newData)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 验证迁移结果
	if !result.Success {
		t.Errorf("Migration should succeed, got errors: %v", result.Errors)
	}
	if result.FromVersion != CurrentVersion {
		t.Errorf("Expected from version %s, got %s", CurrentVersion, result.FromVersion)
	}
	if result.MigratedCount != 0 {
		t.Errorf("Expected 0 migrated configs (already new), got %d", result.MigratedCount)
	}

	// 数据应该保持不变
	if string(migratedData) != string(newData) {
		t.Error("Data should remain unchanged for current version")
	}
}

func TestMigrateConfigData_InvalidData(t *testing.T) {
	// 测试无效JSON数据
	invalidData := []byte("invalid json data")

	result, _, err := MigrateConfigData(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data")
	}
	if result.Success {
		t.Error("Migration should fail for invalid data")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors for invalid data")
	}
}

func TestMigrateConfigData_NilConfigs(t *testing.T) {
	// 创建包含nil配置的数据
	legacyConfigs := map[string]*LegacyProxyConfig{
		"config1": {
			ID:        "config1",
			Name:      "Valid Config",
			Subdomain: "test1",
			TargetURL: "https://example1.com",
			Protocol:  "https",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		"config2": nil, // nil配置
	}

	legacyData, err := json.Marshal(legacyConfigs)
	if err != nil {
		t.Fatalf("Failed to marshal legacy data: %v", err)
	}

	result, _, err := MigrateConfigData(legacyData)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 应该成功，但有跳过的配置
	if !result.Success {
		t.Errorf("Migration should succeed, got errors: %v", result.Errors)
	}
	if result.MigratedCount != 1 {
		t.Errorf("Expected 1 migrated config, got %d", result.MigratedCount)
	}
	if result.SkippedCount != 1 {
		t.Errorf("Expected 1 skipped config, got %d", result.SkippedCount)
	}
}

func TestValidateConfigIntegrity(t *testing.T) {
	configs := map[string]*ProxyConfig{
		"valid-config": {
			ID:           "valid-config",
			Name:         "Valid Config",
			Subdomain:    "valid",
			TargetURL:    "https://example.com",
			AccessTokens: []AccessToken{},
			TokenStats:   &TokenStats{},
		},
		"invalid-config": {
			ID:           "wrong-id", // ID不匹配
			Name:         "",         // 空名称
			Subdomain:    "invalid",
			TargetURL:    "",  // 空URL
			AccessTokens: nil, // nil AccessTokens
			TokenStats:   nil, // nil TokenStats
		},
		"nil-config": nil, // nil配置
	}

	errors := ValidateConfigIntegrity(configs)

	// 应该有多个错误
	if len(errors) == 0 {
		t.Error("Expected validation errors")
	}

	// 检查特定错误
	errorMessages := make(map[string]bool)
	for _, err := range errors {
		errorMessages[err] = true
	}

	expectedErrors := []string{
		"config invalid-config: ID mismatch",
		"config invalid-config: empty name",
		"config invalid-config: empty target URL",
		"config invalid-config: AccessTokens is nil",
		"config invalid-config: TokenStats is nil",
		"config nil-config is nil",
	}

	for _, expectedError := range expectedErrors {
		if !errorMessages[expectedError] {
			t.Errorf("Expected error not found: %s", expectedError)
		}
	}
}

func TestGetDataVersion(t *testing.T) {
	// 测试新版本数据
	newConfigs := map[string]*ProxyConfig{
		"config1": {
			ID:           "config1",
			AccessTokens: []AccessToken{},
			TokenStats:   &TokenStats{},
		},
	}
	newData, _ := json.Marshal(newConfigs)

	version, err := GetDataVersion(newData)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if version.Version != CurrentVersion {
		t.Errorf("Expected version %s, got %s", CurrentVersion, version.Version)
	}

	// 测试旧版本数据
	legacyConfigs := map[string]*LegacyProxyConfig{
		"config1": {
			ID:   "config1",
			Name: "Legacy Config",
		},
	}
	legacyData, _ := json.Marshal(legacyConfigs)

	version, err = GetDataVersion(legacyData)
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if version.Version != DataVersionV1_0 {
		t.Errorf("Expected version %s, got %s", DataVersionV1_0, version.Version)
	}

	// 测试无效数据
	invalidData := []byte("invalid")
	_, err = GetDataVersion(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

func TestNeedsMigration(t *testing.T) {
	// 测试新版本数据（不需要迁移）
	newConfigs := map[string]*ProxyConfig{
		"config1": {
			ID:           "config1",
			AccessTokens: []AccessToken{},
			TokenStats:   &TokenStats{},
		},
	}
	newData, _ := json.Marshal(newConfigs)

	needs, err := NeedsMigration(newData)
	if err != nil {
		t.Fatalf("Failed to check migration need: %v", err)
	}
	if needs {
		t.Error("New version data should not need migration")
	}

	// 测试旧版本数据（需要迁移）
	legacyConfigs := map[string]*LegacyProxyConfig{
		"config1": {
			ID:   "config1",
			Name: "Legacy Config",
		},
	}
	legacyData, _ := json.Marshal(legacyConfigs)

	needs, err = NeedsMigration(legacyData)
	if err != nil {
		t.Fatalf("Failed to check migration need: %v", err)
	}
	if !needs {
		t.Error("Legacy version data should need migration")
	}
}

func TestMigrateLegacyConfig(t *testing.T) {
	legacy := &LegacyProxyConfig{
		ID:        "test-config",
		Name:      "Test Config",
		Subdomain: "test",
		TargetURL: "https://example.com",
		Protocol:  "https",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Stats: &ConfigStats{
			RequestCount: 100,
		},
	}

	migrated := migrateLegacyConfig(legacy)

	// 验证基本字段
	if migrated.ID != legacy.ID {
		t.Errorf("Expected ID %s, got %s", legacy.ID, migrated.ID)
	}
	if migrated.Name != legacy.Name {
		t.Errorf("Expected name %s, got %s", legacy.Name, migrated.Name)
	}
	if migrated.Stats.RequestCount != legacy.Stats.RequestCount {
		t.Errorf("Expected request count %d, got %d", legacy.Stats.RequestCount, migrated.Stats.RequestCount)
	}

	// 验证新字段
	if migrated.AccessTokens == nil {
		t.Error("AccessTokens should be initialized")
	}
	if len(migrated.AccessTokens) != 0 {
		t.Errorf("Expected 0 access tokens, got %d", len(migrated.AccessTokens))
	}
	if migrated.TokenStats == nil {
		t.Error("TokenStats should be initialized")
	}
}
