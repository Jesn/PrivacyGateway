package proxyconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"privacygateway/internal/logger"
)

// PersistentStorage 持久化存储实现
type PersistentStorage struct {
	*MemoryStorage
	filePath     string
	saveInterval time.Duration
	autoSave     bool
	logger       *logger.Logger
	saveMutex    sync.Mutex
	stopChan     chan struct{}
}

// NewPersistentStorage 创建持久化存储实例
func NewPersistentStorage(filePath string, maxEntries int, autoSave bool, log *logger.Logger) *PersistentStorage {
	ps := &PersistentStorage{
		MemoryStorage: NewMemoryStorage(maxEntries),
		filePath:      filePath,
		saveInterval:  30 * time.Second,
		autoSave:      autoSave,
		logger:        log,
		stopChan:      make(chan struct{}),
	}

	// 启动时加载配置
	if err := ps.LoadFromFile(); err != nil {
		log.Error("failed to load configs from file", "error", err, "file", filePath)
	} else {
		log.Info("configs loaded from file", "file", filePath, "count", len(ps.configs))
	}

	// 启动自动保存
	if autoSave {
		ps.StartAutoSave()
		log.Info("auto save enabled", "interval", ps.saveInterval)
	}

	return ps
}

// SaveToFile 保存配置到文件
func (ps *PersistentStorage) SaveToFile() error {
	ps.saveMutex.Lock()
	defer ps.saveMutex.Unlock()

	ps.mutex.RLock()
	configsCopy := make(map[string]*ProxyConfig)
	for k, v := range ps.configs {
		configsCopy[k] = v
	}
	ps.mutex.RUnlock()

	data, err := json.MarshalIndent(configsCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configs: %w", err)
	}

	// 创建目录（如果不存在）
	if dir := filepath.Dir(ps.filePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 写入临时文件，然后原子性重命名
	tempFile := ps.filePath + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, ps.filePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	ps.logger.Debug("configs saved to file", "file", ps.filePath, "count", len(configsCopy))
	return nil
}

// LoadFromFile 从文件加载配置
func (ps *PersistentStorage) LoadFromFile() error {
	if _, err := os.Stat(ps.filePath); os.IsNotExist(err) {
		ps.logger.Info("config file does not exist, starting with empty storage", "file", ps.filePath)
		return nil // 文件不存在，跳过加载
	}

	data, err := ioutil.ReadFile(ps.filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs map[string]*ProxyConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.configs = configs
	ps.rebuildSubdomainIndex()

	return nil
}

// rebuildSubdomainIndex 重建子域名索引
func (ps *PersistentStorage) rebuildSubdomainIndex() {
	ps.subdomains = make(map[string]string)
	for id, config := range ps.configs {
		ps.subdomains[config.Subdomain] = id
	}
}

// StartAutoSave 启动自动保存
func (ps *PersistentStorage) StartAutoSave() {
	go func() {
		ticker := time.NewTicker(ps.saveInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ps.SaveToFile(); err != nil {
					ps.logger.Error("auto save failed", "error", err)
				}
			case <-ps.stopChan:
				ps.logger.Info("auto save stopped")
				return
			}
		}
	}()
}

// StopAutoSave 停止自动保存
func (ps *PersistentStorage) StopAutoSave() {
	close(ps.stopChan)
}

// Add 添加配置（重写以支持持久化）
func (ps *PersistentStorage) Add(config *ProxyConfig) error {
	if err := ps.MemoryStorage.Add(config); err != nil {
		return err
	}

	// 立即保存到文件
	if err := ps.SaveToFile(); err != nil {
		ps.logger.Error("failed to save after add", "error", err)
		// 不返回错误，因为内存操作已经成功
	}

	return nil
}

// Update 更新配置（重写以支持持久化）
func (ps *PersistentStorage) Update(id string, config *ProxyConfig) error {
	if err := ps.MemoryStorage.Update(id, config); err != nil {
		return err
	}

	// 立即保存到文件
	if err := ps.SaveToFile(); err != nil {
		ps.logger.Error("failed to save after update", "error", err)
		// 不返回错误，因为内存操作已经成功
	}

	return nil
}

// Delete 删除配置（重写以支持持久化）
func (ps *PersistentStorage) Delete(id string) error {
	if err := ps.MemoryStorage.Delete(id); err != nil {
		return err
	}

	// 立即保存到文件
	if err := ps.SaveToFile(); err != nil {
		ps.logger.Error("failed to save after delete", "error", err)
		// 不返回错误，因为内存操作已经成功
	}

	return nil
}

// Clear 清空所有配置（重写以支持持久化）
func (ps *PersistentStorage) Clear() {
	ps.MemoryStorage.Clear()

	// 立即保存到文件
	if err := ps.SaveToFile(); err != nil {
		ps.logger.Error("failed to save after clear", "error", err)
	}
}

// Shutdown 优雅关闭
func (ps *PersistentStorage) Shutdown() error {
	ps.StopAutoSave()

	// 最后保存一次
	if err := ps.SaveToFile(); err != nil {
		return fmt.Errorf("failed to save on shutdown: %w", err)
	}

	ps.logger.Info("persistent storage shutdown complete")
	return nil
}
