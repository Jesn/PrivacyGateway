package accesslog

import (
	"sync"
	"time"
)

// Storage 日志存储接口
type Storage interface {
	// Add 添加日志记录
	Add(log *AccessLog) error
	
	// Query 查询日志记录
	Query(filter *LogFilter) (*LogResponse, error)
	
	// GetStats 获取存储统计信息
	GetStats() *StorageStats
	
	// Clear 清空所有日志
	Clear()
	
	// Close 关闭存储
	Close() error
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	logs         []AccessLog   // 日志存储数组
	head         int           // 写入位置
	size         int           // 当前大小
	maxEntries   int           // 最大条数
	maxMemoryMB  float64       // 最大内存使用（MB）
	retentionHours int         // 保留时间（小时）
	maxBodySize  int           // 响应体最大大小
	
	mutex        sync.RWMutex  // 读写锁
	cleanupCount int64         // 清理次数
	lastCleanup  time.Time     // 最后清理时间
	
	// 清理相关
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewMemoryStorage 创建新的内存存储
func NewMemoryStorage(maxEntries int, maxMemoryMB float64, retentionHours int, maxBodySize int) *MemoryStorage {
	storage := &MemoryStorage{
		logs:           make([]AccessLog, maxEntries),
		head:           0,
		size:           0,
		maxEntries:     maxEntries,
		maxMemoryMB:    maxMemoryMB,
		retentionHours: retentionHours,
		maxBodySize:    maxBodySize,
		cleanupCount:   0,
		lastCleanup:    time.Now(),
		stopCleanup:    make(chan struct{}),
	}
	
	// 启动定期清理
	storage.startCleanup()
	
	return storage
}

// Add 添加日志记录
func (s *MemoryStorage) Add(log *AccessLog) error {
	if log == nil {
		return ErrInvalidLogID
	}
	
	// 验证日志记录
	if err := log.Validate(); err != nil {
		return err
	}
	
	// 截断响应体
	if len(log.ResponseBody) > s.maxBodySize {
		log.ResponseBody = TruncateBody([]byte(log.ResponseBody), s.maxBodySize)
	}
	
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// 检查内存使用
	if s.isMemoryLimitExceeded() {
		s.forceCleanup()
	}
	
	// 添加到环形缓冲区
	s.logs[s.head] = *log
	s.head = (s.head + 1) % s.maxEntries
	
	if s.size < s.maxEntries {
		s.size++
	}
	
	return nil
}

// Query 查询日志记录
func (s *MemoryStorage) Query(filter *LogFilter) (*LogResponse, error) {
	if filter == nil {
		filter = &LogFilter{}
	}
	
	// 验证和设置默认值
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	filter.SetDefaults()
	
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// 收集匹配的日志
	var matchedLogs []AccessLog
	
	for i := 0; i < s.size; i++ {
		// 计算实际索引（从最老的开始）
		var idx int
		if s.size < s.maxEntries {
			idx = i
		} else {
			idx = (s.head + i) % s.maxEntries
		}
		
		log := s.logs[idx]
		
		// 应用筛选条件
		if s.matchesFilter(&log, filter) {
			matchedLogs = append(matchedLogs, log)
		}
	}
	
	// 按时间倒序排列（最新的在前）
	for i, j := 0, len(matchedLogs)-1; i < j; i, j = i+1, j-1 {
		matchedLogs[i], matchedLogs[j] = matchedLogs[j], matchedLogs[i]
	}
	
	// 分页处理
	total := len(matchedLogs)
	totalPages := (total + filter.Limit - 1) / filter.Limit
	
	start := (filter.Page - 1) * filter.Limit
	end := start + filter.Limit
	
	if start >= total {
		matchedLogs = []AccessLog{}
	} else {
		if end > total {
			end = total
		}
		matchedLogs = matchedLogs[start:end]
	}
	
	return &LogResponse{
		Logs:       matchedLogs,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetStats 获取存储统计信息
func (s *MemoryStorage) GetStats() *StorageStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	stats := &StorageStats{
		CurrentEntries: s.size,
		MaxEntries:     s.maxEntries,
		MemoryUsageMB:  s.calculateMemoryUsage(),
		CleanupCount:   s.cleanupCount,
		LastCleanup:    s.lastCleanup.Format(time.RFC3339),
	}
	
	// 获取最老和最新的日志时间
	if s.size > 0 {
		var oldestIdx, newestIdx int
		if s.size < s.maxEntries {
			oldestIdx = 0
			newestIdx = s.size - 1
		} else {
			oldestIdx = s.head
			newestIdx = (s.head - 1 + s.maxEntries) % s.maxEntries
		}
		
		stats.OldestEntry = s.logs[oldestIdx].Timestamp.Format(time.RFC3339)
		stats.NewestEntry = s.logs[newestIdx].Timestamp.Format(time.RFC3339)
	}
	
	return stats
}

// Clear 清空所有日志
func (s *MemoryStorage) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.head = 0
	s.size = 0
	s.cleanupCount++
	s.lastCleanup = time.Now()
}

// Close 关闭存储
func (s *MemoryStorage) Close() error {
	close(s.stopCleanup)
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	return nil
}

// matchesFilter 检查日志是否匹配筛选条件
func (s *MemoryStorage) matchesFilter(log *AccessLog, filter *LogFilter) bool {
	// 域名筛选
	if !MatchesDomain(log.TargetHost, filter.Domain) {
		return false
	}
	
	// 状态码筛选
	if !ContainsStatusCode(filter.StatusCode, log.StatusCode) {
		return false
	}
	
	// 时间范围筛选
	if !IsWithinTimeRange(log.Timestamp, filter.FromTime, filter.ToTime) {
		return false
	}
	
	return true
}

// calculateMemoryUsage 计算当前内存使用量（MB）
func (s *MemoryStorage) calculateMemoryUsage() float64 {
	var totalBytes int64
	
	for i := 0; i < s.size; i++ {
		var idx int
		if s.size < s.maxEntries {
			idx = i
		} else {
			idx = (s.head + i) % s.maxEntries
		}
		totalBytes += EstimateMemoryUsage(&s.logs[idx])
	}
	
	return float64(totalBytes) / (1024 * 1024) // 转换为MB
}

// isMemoryLimitExceeded 检查是否超过内存限制
func (s *MemoryStorage) isMemoryLimitExceeded() bool {
	if s.maxMemoryMB <= 0 {
		return false
	}
	return s.calculateMemoryUsage() > s.maxMemoryMB
}

// startCleanup 启动定期清理
func (s *MemoryStorage) startCleanup() {
	// 每5分钟清理一次
	s.cleanupTicker = time.NewTicker(5 * time.Minute)
	
	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.performCleanup()
			case <-s.stopCleanup:
				return
			}
		}
	}()
}

// performCleanup 执行清理操作
func (s *MemoryStorage) performCleanup() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.size == 0 {
		return
	}
	
	cutoff := time.Now().Add(-time.Duration(s.retentionHours) * time.Hour)
	cleaned := 0
	
	// 从最老的日志开始检查
	for i := 0; i < s.size; i++ {
		var idx int
		if s.size < s.maxEntries {
			idx = i
		} else {
			idx = (s.head + i) % s.maxEntries
		}
		
		if s.logs[idx].Timestamp.Before(cutoff) {
			// 清理过期日志
			s.logs[idx] = AccessLog{} // 清空内存
			cleaned++
		} else {
			break // 后面的日志都是更新的
		}
	}
	
	if cleaned > 0 {
		// 重新整理数组
		s.compactArray(cleaned)
		s.cleanupCount++
		s.lastCleanup = time.Now()
	}
}

// forceCleanup 强制清理以释放内存
func (s *MemoryStorage) forceCleanup() {
	if s.size == 0 {
		return
	}
	
	// 清理最老的25%的日志
	cleanCount := s.size / 4
	if cleanCount == 0 {
		cleanCount = 1
	}
	
	s.compactArray(cleanCount)
	s.cleanupCount++
	s.lastCleanup = time.Now()
}

// compactArray 压缩数组，移除已清理的元素
func (s *MemoryStorage) compactArray(removeCount int) {
	if removeCount >= s.size {
		s.head = 0
		s.size = 0
		return
	}
	
	// 移动数据
	if s.size < s.maxEntries {
		// 简单情况：数组未满
		copy(s.logs[0:], s.logs[removeCount:s.size])
		s.size -= removeCount
	} else {
		// 复杂情况：环形缓冲区
		newHead := (s.head + removeCount) % s.maxEntries
		s.head = newHead
		s.size -= removeCount
	}
}
