package accesslog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"privacygateway/internal/config"
	"privacygateway/internal/logger"
)

// Recorder 日志记录器
type Recorder struct {
	storage Storage
	config  *config.Config
	logger  *logger.Logger

	// 异步处理
	logChan chan *AccessLog
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// 统计信息
	mutex         sync.RWMutex
	totalLogs     int64
	errorCount    int64
	lastError     error
	lastErrorTime time.Time
}

// NewRecorder 创建新的日志记录器
func NewRecorder(cfg *config.Config, log *logger.Logger) (*Recorder, error) {
	// 创建存储
	storage := NewMemoryStorage(
		cfg.LogMaxEntries,
		cfg.LogMaxMemoryMB,
		cfg.LogRetentionHours,
		cfg.LogMaxBodySize,
	)

	ctx, cancel := context.WithCancel(context.Background())

	recorder := &Recorder{
		storage:    storage,
		config:     cfg,
		logger:     log,
		logChan:    make(chan *AccessLog, 1000), // 缓冲1000条日志
		ctx:        ctx,
		cancel:     cancel,
		totalLogs:  0,
		errorCount: 0,
	}

	// 启动异步处理协程
	recorder.startWorkers()

	return recorder, nil
}

// RecordRequest 记录HTTP请求
func (r *Recorder) RecordRequest(req *http.Request, statusCode int, responseBody string, duration time.Duration, responseSize int64, endpoint string) {
	// 创建日志记录
	log := &AccessLog{
		ID:           GenerateLogID(),
		Timestamp:    time.Now(),
		Method:       req.Method,
		RequestType:  DetermineRequestType(req, endpoint),
		TargetHost:   r.extractTargetHost(req),
		TargetPath:   r.extractTargetPath(req),
		StatusCode:   statusCode,
		ResponseBody: r.processResponseBody(responseBody, statusCode),
		UserAgent:    req.UserAgent(),
		ClientIP:     GetClientIP(req),
		Duration:     duration.Milliseconds(),
		RequestSize:  req.ContentLength,
		ResponseSize: responseSize,
	}

	// 异步发送到处理队列
	select {
	case r.logChan <- log:
		// 成功发送
	default:
		// 队列满了，丢弃日志并记录错误
		r.recordError(fmt.Errorf("log queue is full, dropping log"))
	}
}

// RecordFromCapture 从响应捕获器记录日志
func (r *Recorder) RecordFromCapture(req *http.Request, capture *ResponseCapture, endpoint string) {
	// 使用实际发送给目标服务器的User-Agent，如果没有设置则使用客户端的
	actualUserAgent := capture.GetActualUserAgent()
	if actualUserAgent == "" {
		actualUserAgent = req.UserAgent()
	}

	log := &AccessLog{
		ID:             GenerateLogID(),
		Timestamp:      capture.startTime,
		Method:         req.Method,
		RequestType:    DetermineRequestTypeWithResponse(req, endpoint, capture.GetResponseHeaders()),
		TargetHost:     r.extractTargetHost(req),
		TargetPath:     r.extractTargetPath(req),
		StatusCode:     capture.GetStatusCode(),
		ResponseBody:   capture.GetBody(),
		UserAgent:      actualUserAgent,
		ProxyInfo:      capture.GetProxyInfo(),
		ClientIP:       GetClientIP(req),
		Duration:       capture.GetDuration(),
		RequestSize:    req.ContentLength,
		ResponseSize:   capture.GetBodySize(),
		RequestHeaders: capture.GetRequestHeaders(),
		RequestBody:    capture.GetRequestBody(),
	}

	// 异步发送到处理队列
	select {
	case r.logChan <- log:
		// 成功发送
	default:
		// 队列满了，丢弃日志并记录错误
		r.recordError(fmt.Errorf("log queue is full, dropping log"))
	}
}

// Query 查询日志
func (r *Recorder) Query(filter *LogFilter) (*LogResponse, error) {
	return r.storage.Query(filter)
}

// GetStats 获取统计信息
func (r *Recorder) GetStats() *RecorderStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	storageStats := r.storage.GetStats()

	return &RecorderStats{
		StorageStats:  *storageStats,
		TotalLogs:     r.totalLogs,
		ErrorCount:    r.errorCount,
		LastError:     r.formatError(r.lastError),
		LastErrorTime: r.formatTime(r.lastErrorTime),
		QueueSize:     len(r.logChan),
		QueueCapacity: cap(r.logChan),
	}
}

// Close 关闭记录器
func (r *Recorder) Close() error {
	// 停止接收新日志
	r.cancel()

	// 等待所有工作协程完成
	r.wg.Wait()

	// 关闭存储
	return r.storage.Close()
}

// startWorkers 启动工作协程
func (r *Recorder) startWorkers() {
	// 启动多个工作协程处理日志
	workerCount := 3
	for i := 0; i < workerCount; i++ {
		r.wg.Add(1)
		go r.worker()
	}
}

// worker 工作协程
func (r *Recorder) worker() {
	defer r.wg.Done()

	for {
		select {
		case log := <-r.logChan:
			if err := r.processLog(log); err != nil {
				r.recordError(err)
			} else {
				r.incrementTotalLogs()
			}
		case <-r.ctx.Done():
			// 处理剩余的日志
			for {
				select {
				case log := <-r.logChan:
					r.processLog(log)
				default:
					return
				}
			}
		}
	}
}

// processLog 处理单条日志
func (r *Recorder) processLog(log *AccessLog) error {
	// 验证日志
	if err := log.Validate(); err != nil {
		return fmt.Errorf("invalid log: %w", err)
	}

	// 存储日志
	if err := r.storage.Add(log); err != nil {
		return fmt.Errorf("failed to store log: %w", err)
	}

	// 记录到系统日志（仅错误状态码）
	if log.IsErrorStatus() {
		r.logger.Warn("HTTP error recorded",
			"method", log.Method,
			"target", log.TargetHost+log.TargetPath,
			"status", log.StatusCode,
			"duration", log.Duration,
			"client_ip", log.ClientIP,
		)
	}

	return nil
}

// extractTargetHost 提取目标主机
func (r *Recorder) extractTargetHost(req *http.Request) string {
	targetURL := req.URL.Query().Get("target")
	if targetURL == "" {
		return ""
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	return parsed.Host
}

// extractTargetPath 提取目标路径
func (r *Recorder) extractTargetPath(req *http.Request) string {
	targetURL := req.URL.Query().Get("target")
	if targetURL == "" {
		return ""
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// 包含查询参数
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}

	return path
}

// processResponseBody 处理响应体
func (r *Recorder) processResponseBody(body string, statusCode int) string {
	// 根据配置决定是否记录200状态码的响应体
	if statusCode == 200 && !r.config.LogRecord200 {
		return ""
	}

	// 截断过长的响应体
	if len(body) > r.config.LogMaxBodySize {
		return TruncateBody([]byte(body), r.config.LogMaxBodySize)
	}

	return body
}

// recordError 记录错误
func (r *Recorder) recordError(err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.errorCount++
	r.lastError = err
	r.lastErrorTime = time.Now()

	// 记录到系统日志
	r.logger.Error("access log recorder error", "error", err)
}

// incrementTotalLogs 增加总日志计数
func (r *Recorder) incrementTotalLogs() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.totalLogs++
}

// formatError 格式化错误信息
func (r *Recorder) formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// formatTime 格式化时间
func (r *Recorder) formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// RecorderStats 记录器统计信息
type RecorderStats struct {
	StorageStats  StorageStats `json:"storage_stats"`
	TotalLogs     int64        `json:"total_logs"`
	ErrorCount    int64        `json:"error_count"`
	LastError     string       `json:"last_error,omitempty"`
	LastErrorTime string       `json:"last_error_time,omitempty"`
	QueueSize     int          `json:"queue_size"`
	QueueCapacity int          `json:"queue_capacity"`
}

// CreateMiddleware 创建中间件
func (r *Recorder) CreateMiddleware() *LoggingMiddleware {
	return NewLoggingMiddleware(r.storage, r.config.LogMaxBodySize, r.config.LogRecord200)
}

// WrapHandler 包装HTTP处理器
func (r *Recorder) WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	middleware := r.CreateMiddleware()
	return middleware.Wrap(handler)
}

// IsLogRecord200Enabled 返回是否启用了200状态码记录
func (r *Recorder) IsLogRecord200Enabled() bool {
	return r.config.LogRecord200
}
