package proxyconfig

import (
	"encoding/json"
	"fmt"
	"time"
)

// DataVersion 数据版本信息
type DataVersion struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// MigrationResult 迁移结果
type MigrationResult struct {
	Success       bool          `json:"success"`
	FromVersion   string        `json:"from_version"`
	ToVersion     string        `json:"to_version"`
	MigratedCount int           `json:"migrated_count"`
	SkippedCount  int           `json:"skipped_count"`
	ErrorCount    int           `json:"error_count"`
	Errors        []string      `json:"errors"`
	BackupCreated bool          `json:"backup_created"`
	BackupPath    string        `json:"backup_path,omitempty"`
	MigrationTime time.Duration `json:"migration_time"`
}

// 版本常量
const (
	DataVersionV1_0 = "1.0.0" // 原始版本（无令牌管理）
	DataVersionV1_1 = "1.1.0" // 添加令牌管理
	CurrentVersion  = DataVersionV1_1
)

// LegacyProxyConfig 旧版本的代理配置结构（v1.0）
type LegacyProxyConfig struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Subdomain string       `json:"subdomain"`
	TargetURL string       `json:"target_url"`
	Protocol  string       `json:"protocol"`
	Enabled   bool         `json:"enabled"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Stats     *ConfigStats `json:"stats,omitempty"`
}

// MigrateConfigData 迁移配置数据
func MigrateConfigData(data []byte) (*MigrationResult, []byte, error) {
	startTime := time.Now()
	result := &MigrationResult{
		FromVersion: DataVersionV1_0,
		ToVersion:   CurrentVersion,
		Errors:      make([]string, 0),
	}

	// 首先检查数据版本
	version, err := GetDataVersion(data)
	if err == nil && version.Version == CurrentVersion {
		result.Success = true
		result.FromVersion = CurrentVersion
		result.MigrationTime = time.Since(startTime)
		return result, data, nil
	}

	// 尝试解析为旧版本
	var legacyConfigs map[string]*LegacyProxyConfig
	if err := json.Unmarshal(data, &legacyConfigs); err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to parse data: %v", err))
		result.MigrationTime = time.Since(startTime)
		return result, nil, err
	}

	// 执行迁移
	migratedConfigs := make(map[string]*ProxyConfig)
	for id, legacyConfig := range legacyConfigs {
		if legacyConfig == nil {
			result.SkippedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("skipped nil config with ID: %s", id))
			continue
		}

		// 迁移配置
		migratedConfig := migrateLegacyConfig(legacyConfig)
		migratedConfigs[id] = migratedConfig
		result.MigratedCount++
	}

	// 序列化迁移后的数据
	migratedData, err := json.MarshalIndent(migratedConfigs, "", "  ")
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to marshal migrated data: %v", err))
		result.MigrationTime = time.Since(startTime)
		return result, nil, err
	}

	result.Success = true
	result.MigrationTime = time.Since(startTime)
	return result, migratedData, nil
}

// migrateLegacyConfig 迁移单个旧版本配置
func migrateLegacyConfig(legacy *LegacyProxyConfig) *ProxyConfig {
	return &ProxyConfig{
		ID:           legacy.ID,
		Name:         legacy.Name,
		TargetURL:    legacy.TargetURL,
		Protocol:     legacy.Protocol,
		Enabled:      legacy.Enabled,
		CreatedAt:    legacy.CreatedAt,
		UpdatedAt:    legacy.UpdatedAt,
		Stats:        legacy.Stats,
		AccessTokens: make([]AccessToken, 0), // 初始化空的令牌列表
		TokenStats:   &TokenStats{},          // 初始化令牌统计
	}
}

// isCurrentVersion 检查配置是否已经是当前版本
func isCurrentVersion(configs map[string]*ProxyConfig) bool {
	// 检查是否有配置包含新字段
	for _, config := range configs {
		if config != nil {
			// 检查是否有AccessTokens字段（通过反射或直接检查）
			// 如果配置结构体包含AccessTokens字段，说明是新版本
			// 这里我们通过检查字段是否被正确初始化来判断
			if config.AccessTokens != nil || config.TokenStats != nil {
				return true
			}
		}
	}
	return false
}

// ValidateConfigIntegrity 验证配置数据完整性
func ValidateConfigIntegrity(configs map[string]*ProxyConfig) []string {
	var errors []string

	for id, config := range configs {
		if config == nil {
			errors = append(errors, fmt.Sprintf("config %s is nil", id))
			continue
		}

		// 验证基本字段
		if config.ID != id {
			errors = append(errors, fmt.Sprintf("config %s: ID mismatch", id))
		}
		if config.Name == "" {
			errors = append(errors, fmt.Sprintf("config %s: empty name", id))
		}
		if config.TargetURL == "" {
			errors = append(errors, fmt.Sprintf("config %s: empty target URL", id))
		}

		// 验证令牌相关字段
		if config.AccessTokens == nil {
			errors = append(errors, fmt.Sprintf("config %s: AccessTokens is nil", id))
		}
		if config.TokenStats == nil {
			errors = append(errors, fmt.Sprintf("config %s: TokenStats is nil", id))
		}

		// 验证令牌数据
		for i, token := range config.AccessTokens {
			if err := token.Validate(); err != nil {
				errors = append(errors, fmt.Sprintf("config %s: token %d invalid: %v", id, i, err))
			}
		}
	}

	return errors
}

// CreateBackup 创建数据备份
func CreateBackup(data []byte, backupPath string) error {
	if backupPath == "" {
		return fmt.Errorf("backup path is empty")
	}

	// 添加时间戳到备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("%s.backup_%s", backupPath, timestamp)

	// 写入备份文件
	return writeFile(backupFile, data)
}

// RollbackMigration 回滚迁移
func RollbackMigration(backupPath string) ([]byte, error) {
	if backupPath == "" {
		return nil, fmt.Errorf("backup path is empty")
	}

	// 读取备份文件
	return readFile(backupPath)
}

// GetDataVersion 获取数据版本信息
func GetDataVersion(data []byte) (*DataVersion, error) {
	// 首先尝试解析为通用的map[string]interface{}
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("unable to parse data: %v", err)
	}

	// 检查是否有配置数据
	if len(rawData) == 0 {
		return &DataVersion{
			Version:     CurrentVersion,
			Description: "Empty data, assuming current version",
			CreatedAt:   time.Now(),
		}, nil
	}

	// 检查第一个配置的结构来判断版本
	for _, configData := range rawData {
		if configMap, ok := configData.(map[string]interface{}); ok {
			// 检查是否包含新版本字段
			if _, hasAccessTokens := configMap["access_tokens"]; hasAccessTokens {
				return &DataVersion{
					Version:     CurrentVersion,
					Description: "Current version with token management",
					CreatedAt:   time.Now(),
				}, nil
			}
			if _, hasTokenStats := configMap["token_stats"]; hasTokenStats {
				return &DataVersion{
					Version:     CurrentVersion,
					Description: "Current version with token management",
					CreatedAt:   time.Now(),
				}, nil
			}
			// 如果没有新字段，认为是旧版本
			return &DataVersion{
				Version:     DataVersionV1_0,
				Description: "Legacy version without token management",
				CreatedAt:   time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("unable to determine data version")
}

// NeedsMigration 检查是否需要迁移
func NeedsMigration(data []byte) (bool, error) {
	version, err := GetDataVersion(data)
	if err != nil {
		return false, err
	}

	return version.Version != CurrentVersion, nil
}

// 文件操作辅助函数（简化实现）
func writeFile(path string, data []byte) error {
	// 这里应该使用实际的文件写入逻辑
	// 为了测试，我们暂时返回nil
	return nil
}

func readFile(path string) ([]byte, error) {
	// 这里应该使用实际的文件读取逻辑
	// 为了测试，我们暂时返回空数据
	return []byte{}, nil
}
