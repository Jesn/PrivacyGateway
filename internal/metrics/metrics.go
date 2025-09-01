// Package metrics 提供性能监控和指标收集功能
package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 性能指标收集器
type Metrics struct {
	// 请求统计
	totalRequests    int64
	successRequests  int64
	errorRequests    int64
	
	// 响应时间统计
	totalResponseTime int64
	minResponseTime   int64
	maxResponseTime   int64
	
	// 令牌统计
	totalTokens      int64
	activeTokens     int64
	tokenValidations int64
	
	// 配置统计
	totalConfigs     int64
	activeConfigs    int64
	
	// 系统资源
	mutex            sync.RWMutex
	lastUpdate       time.Time
	memStats         runtime.MemStats
	
	// 历史数据（最近1小时，每分钟一个数据点）
	requestHistory   [60]int64
	responseHistory  [60]int64
	errorHistory     [60]int64
	historyIndex     int
	lastHistoryUpdate time.Time
}

// NewMetrics 创建新的指标收集器
func NewMetrics() *Metrics {
	m := &Metrics{
		minResponseTime:   int64(^uint64(0) >> 1), // 设置为最大值
		lastUpdate:        time.Now(),
		lastHistoryUpdate: time.Now(),
	}
	
	// 启动定期更新
	go m.updateLoop()
	
	return m
}

// RecordRequest 记录请求
func (m *Metrics) RecordRequest(duration time.Duration, success bool) {
	atomic.AddInt64(&m.totalRequests, 1)
	
	durationMs := duration.Milliseconds()
	atomic.AddInt64(&m.totalResponseTime, durationMs)
	
	if success {
		atomic.AddInt64(&m.successRequests, 1)
	} else {
		atomic.AddInt64(&m.errorRequests, 1)
	}
	
	// 更新最小/最大响应时间
	for {
		current := atomic.LoadInt64(&m.minResponseTime)
		if durationMs >= current || atomic.CompareAndSwapInt64(&m.minResponseTime, current, durationMs) {
			break
		}
	}
	
	for {
		current := atomic.LoadInt64(&m.maxResponseTime)
		if durationMs <= current || atomic.CompareAndSwapInt64(&m.maxResponseTime, current, durationMs) {
			break
		}
	}
}

// RecordTokenValidation 记录令牌验证
func (m *Metrics) RecordTokenValidation() {
	atomic.AddInt64(&m.tokenValidations, 1)
}

// UpdateTokenCount 更新令牌数量
func (m *Metrics) UpdateTokenCount(total, active int64) {
	atomic.StoreInt64(&m.totalTokens, total)
	atomic.StoreInt64(&m.activeTokens, active)
}

// UpdateConfigCount 更新配置数量
func (m *Metrics) UpdateConfigCount(total, active int64) {
	atomic.StoreInt64(&m.totalConfigs, total)
	atomic.StoreInt64(&m.activeConfigs, active)
}

// GetSnapshot 获取当前指标快照
func (m *Metrics) GetSnapshot() *Snapshot {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	totalReq := atomic.LoadInt64(&m.totalRequests)
	successReq := atomic.LoadInt64(&m.successRequests)
	errorReq := atomic.LoadInt64(&m.errorRequests)
	totalRespTime := atomic.LoadInt64(&m.totalResponseTime)
	minRespTime := atomic.LoadInt64(&m.minResponseTime)
	maxRespTime := atomic.LoadInt64(&m.maxResponseTime)
	
	var avgResponseTime int64
	if totalReq > 0 {
		avgResponseTime = totalRespTime / totalReq
	}
	
	var successRate float64
	if totalReq > 0 {
		successRate = float64(successReq) / float64(totalReq) * 100
	}
	
	return &Snapshot{
		Timestamp: time.Now(),
		
		// 请求统计
		TotalRequests:   totalReq,
		SuccessRequests: successReq,
		ErrorRequests:   errorReq,
		SuccessRate:     successRate,
		
		// 响应时间统计
		AvgResponseTime: avgResponseTime,
		MinResponseTime: minRespTime,
		MaxResponseTime: maxRespTime,
		
		// 令牌统计
		TotalTokens:      atomic.LoadInt64(&m.totalTokens),
		ActiveTokens:     atomic.LoadInt64(&m.activeTokens),
		TokenValidations: atomic.LoadInt64(&m.tokenValidations),
		
		// 配置统计
		TotalConfigs:  atomic.LoadInt64(&m.totalConfigs),
		ActiveConfigs: atomic.LoadInt64(&m.activeConfigs),
		
		// 系统资源
		MemoryUsage:    m.memStats.Alloc,
		MemoryTotal:    m.memStats.TotalAlloc,
		GCCount:        m.memStats.NumGC,
		Goroutines:     int64(runtime.NumGoroutine()),
		
		// 历史数据
		RequestHistory:  m.requestHistory,
		ResponseHistory: m.responseHistory,
		ErrorHistory:    m.errorHistory,
	}
}

// Snapshot 指标快照
type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	
	// 请求统计
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	SuccessRate     float64 `json:"success_rate"`
	
	// 响应时间统计 (毫秒)
	AvgResponseTime int64 `json:"avg_response_time"`
	MinResponseTime int64 `json:"min_response_time"`
	MaxResponseTime int64 `json:"max_response_time"`
	
	// 令牌统计
	TotalTokens      int64 `json:"total_tokens"`
	ActiveTokens     int64 `json:"active_tokens"`
	TokenValidations int64 `json:"token_validations"`
	
	// 配置统计
	TotalConfigs  int64 `json:"total_configs"`
	ActiveConfigs int64 `json:"active_configs"`
	
	// 系统资源
	MemoryUsage uint64 `json:"memory_usage"`
	MemoryTotal uint64 `json:"memory_total"`
	GCCount     uint32 `json:"gc_count"`
	Goroutines  int64  `json:"goroutines"`
	
	// 历史数据（最近60分钟）
	RequestHistory  [60]int64 `json:"request_history"`
	ResponseHistory [60]int64 `json:"response_history"`
	ErrorHistory    [60]int64 `json:"error_history"`
}

// updateLoop 定期更新系统指标
func (m *Metrics) updateLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.updateSystemMetrics()
		m.updateHistory()
	}
}

// updateSystemMetrics 更新系统指标
func (m *Metrics) updateSystemMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	runtime.ReadMemStats(&m.memStats)
	m.lastUpdate = time.Now()
}

// updateHistory 更新历史数据
func (m *Metrics) updateHistory() {
	now := time.Now()
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 每分钟更新一次历史数据
	if now.Sub(m.lastHistoryUpdate) >= time.Minute {
		// 计算当前分钟的请求数
		currentRequests := atomic.LoadInt64(&m.totalRequests)
		currentErrors := atomic.LoadInt64(&m.errorRequests)
		currentResponseTime := atomic.LoadInt64(&m.totalResponseTime)
		
		// 更新历史数组
		m.requestHistory[m.historyIndex] = currentRequests
		m.errorHistory[m.historyIndex] = currentErrors
		
		// 计算平均响应时间
		if currentRequests > 0 {
			m.responseHistory[m.historyIndex] = currentResponseTime / currentRequests
		}
		
		// 移动到下一个位置
		m.historyIndex = (m.historyIndex + 1) % 60
		m.lastHistoryUpdate = now
	}
}

// Reset 重置所有计数器
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.successRequests, 0)
	atomic.StoreInt64(&m.errorRequests, 0)
	atomic.StoreInt64(&m.totalResponseTime, 0)
	atomic.StoreInt64(&m.minResponseTime, int64(^uint64(0)>>1))
	atomic.StoreInt64(&m.maxResponseTime, 0)
	atomic.StoreInt64(&m.tokenValidations, 0)
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 清空历史数据
	for i := range m.requestHistory {
		m.requestHistory[i] = 0
		m.responseHistory[i] = 0
		m.errorHistory[i] = 0
	}
	m.historyIndex = 0
	m.lastHistoryUpdate = time.Now()
}

// GetHealthStatus 获取健康状态
func (m *Metrics) GetHealthStatus() *HealthStatus {
	snapshot := m.GetSnapshot()
	
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Checks:    make(map[string]CheckResult),
	}
	
	// 检查错误率
	if snapshot.SuccessRate < 95.0 && snapshot.TotalRequests > 100 {
		status.Status = "degraded"
		status.Checks["error_rate"] = CheckResult{
			Status:  "warning",
			Message: "High error rate detected",
			Value:   100 - snapshot.SuccessRate,
		}
	} else {
		status.Checks["error_rate"] = CheckResult{
			Status:  "ok",
			Message: "Error rate is normal",
			Value:   100 - snapshot.SuccessRate,
		}
	}
	
	// 检查响应时间
	if snapshot.AvgResponseTime > 1000 {
		status.Status = "degraded"
		status.Checks["response_time"] = CheckResult{
			Status:  "warning",
			Message: "High response time detected",
			Value:   float64(snapshot.AvgResponseTime),
		}
	} else {
		status.Checks["response_time"] = CheckResult{
			Status:  "ok",
			Message: "Response time is normal",
			Value:   float64(snapshot.AvgResponseTime),
		}
	}
	
	// 检查内存使用
	memUsageMB := float64(snapshot.MemoryUsage) / 1024 / 1024
	if memUsageMB > 500 {
		status.Status = "degraded"
		status.Checks["memory"] = CheckResult{
			Status:  "warning",
			Message: "High memory usage detected",
			Value:   memUsageMB,
		}
	} else {
		status.Checks["memory"] = CheckResult{
			Status:  "ok",
			Message: "Memory usage is normal",
			Value:   memUsageMB,
		}
	}
	
	return status
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult 检查结果
type CheckResult struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}
