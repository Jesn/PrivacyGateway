package accesslog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ResponseCapture 响应捕获器，包装 http.ResponseWriter
type ResponseCapture struct {
	http.ResponseWriter
	statusCode      int
	body            *bytes.Buffer
	bodySize        int64
	captureBody     bool
	maxBodySize     int
	startTime       time.Time
	actualUserAgent string            // 实际发送给目标服务器的User-Agent
	proxyInfo       string            // 代理服务器信息
	requestHeaders  map[string]string // 请求头信息
	requestBody     string            // 请求体内容
	responseHeaders map[string]string // 响应头信息
	record200       bool              // 是否记录200状态码的详细信息
}

// NewResponseCapture 创建新的响应捕获器
func NewResponseCapture(w http.ResponseWriter, captureBody bool, maxBodySize int, record200 bool) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter:  w,
		statusCode:      200, // 默认状态码
		body:            &bytes.Buffer{},
		bodySize:        0,
		captureBody:     captureBody,
		maxBodySize:     maxBodySize,
		startTime:       time.Now(),
		responseHeaders: make(map[string]string),
		record200:       record200,
	}
}

// WriteHeader 捕获状态码和响应头
func (rc *ResponseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode

	// 捕获响应头
	for key, values := range rc.ResponseWriter.Header() {
		if len(values) > 0 {
			rc.responseHeaders[key] = values[0] // 只取第一个值
		}
	}

	rc.ResponseWriter.WriteHeader(statusCode)
}

// Write 捕获响应体内容
func (rc *ResponseCapture) Write(data []byte) (int, error) {
	// 写入原始响应
	n, err := rc.ResponseWriter.Write(data)

	// 更新响应大小
	rc.bodySize += int64(n)

	// 根据状态码决定是否捕获响应体
	if rc.captureBody && rc.shouldCaptureBody() {
		// 检查是否超过最大大小限制
		if rc.body.Len()+len(data) <= rc.maxBodySize {
			rc.body.Write(data[:n])
		} else {
			// 只写入剩余可用空间
			remaining := rc.maxBodySize - rc.body.Len()
			if remaining > 0 {
				rc.body.Write(data[:remaining])
			}
		}
	}

	return n, err
}

// GetStatusCode 获取状态码
func (rc *ResponseCapture) GetStatusCode() int {
	return rc.statusCode
}

// GetBody 获取捕获的响应体
func (rc *ResponseCapture) GetBody() string {
	if !rc.captureBody || !rc.shouldCaptureBody() {
		return ""
	}

	bodyBytes := rc.body.Bytes()
	if len(bodyBytes) == 0 {
		return ""
	}

	// 检测内容类型
	contentType := rc.detectContentType(bodyBytes)

	// 处理不同类型的响应内容
	body := rc.formatResponseBody(bodyBytes, contentType)

	// 如果响应体被截断，添加截断标记
	if rc.bodySize > int64(rc.maxBodySize) {
		if len(body) > 13 { // len("...[truncated]") = 13
			body = body[:len(body)-13] + "...[truncated]"
		} else {
			body += "...[truncated]"
		}
	}

	return body
}

// GetBodySize 获取响应体总大小
func (rc *ResponseCapture) GetBodySize() int64 {
	return rc.bodySize
}

// GetDuration 获取请求处理时长（毫秒）
func (rc *ResponseCapture) GetDuration() int64 {
	return time.Since(rc.startTime).Milliseconds()
}

// SetActualUserAgent 设置实际发送给目标服务器的User-Agent
func (rc *ResponseCapture) SetActualUserAgent(userAgent string) {
	rc.actualUserAgent = userAgent
}

// GetActualUserAgent 获取实际发送给目标服务器的User-Agent
func (rc *ResponseCapture) GetActualUserAgent() string {
	return rc.actualUserAgent
}

// SetProxyInfo 设置代理服务器信息
func (rc *ResponseCapture) SetProxyInfo(proxyInfo string) {
	rc.proxyInfo = proxyInfo
}

// GetProxyInfo 获取代理服务器信息
func (rc *ResponseCapture) GetProxyInfo() string {
	return rc.proxyInfo
}

// SetRequestHeaders 设置请求头信息
func (rc *ResponseCapture) SetRequestHeaders(headers map[string]string) {
	rc.requestHeaders = headers
}

// GetRequestHeaders 获取请求头信息
func (rc *ResponseCapture) GetRequestHeaders() map[string]string {
	return rc.requestHeaders
}

// SetRequestBody 设置请求体内容
func (rc *ResponseCapture) SetRequestBody(body string) {
	rc.requestBody = body
}

// GetRequestBody 获取请求体内容
func (rc *ResponseCapture) GetRequestBody() string {
	return rc.requestBody
}

// GetResponseHeaders 获取响应头信息
func (rc *ResponseCapture) GetResponseHeaders() map[string]string {
	return rc.responseHeaders
}

// shouldCaptureBody 判断是否应该捕获响应体
func (rc *ResponseCapture) shouldCaptureBody() bool {
	// 如果配置了记录200状态码，则记录所有状态码
	if rc.record200 {
		return true
	}
	// 否则只记录非200状态码
	return rc.statusCode != 200
}

// LoggingMiddleware 日志记录中间件
type LoggingMiddleware struct {
	storage     Storage
	maxBodySize int
	record200   bool
}

// NewLoggingMiddleware 创建新的日志记录中间件
func NewLoggingMiddleware(storage Storage, maxBodySize int, record200 bool) *LoggingMiddleware {
	return &LoggingMiddleware{
		storage:     storage,
		maxBodySize: maxBodySize,
		record200:   record200,
	}
}

// Wrap 包装HTTP处理器，添加日志记录功能
func (lm *LoggingMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 创建响应捕获器
		capture := NewResponseCapture(w, true, lm.maxBodySize, lm.record200)

		// 记录请求开始时间
		startTime := time.Now()

		// 执行原始处理器
		next(capture, r)

		// 使用实际发送给目标服务器的User-Agent，如果没有设置则使用客户端的
		actualUserAgent := capture.GetActualUserAgent()
		if actualUserAgent == "" {
			actualUserAgent = r.UserAgent()
		}

		// 创建日志记录
		log := &AccessLog{
			ID:             GenerateLogID(),
			Timestamp:      startTime,
			Method:         r.Method,
			TargetHost:     extractTargetHost(r),
			TargetPath:     extractTargetPath(r),
			StatusCode:     capture.GetStatusCode(),
			ResponseBody:   capture.GetBody(),
			UserAgent:      actualUserAgent,
			ProxyInfo:      capture.GetProxyInfo(),
			ClientIP:       GetClientIP(r),
			Duration:       capture.GetDuration(),
			RequestSize:    r.ContentLength,
			ResponseSize:   capture.GetBodySize(),
			RequestHeaders: capture.GetRequestHeaders(),
			RequestBody:    capture.GetRequestBody(),
		}

		// 异步记录日志，避免影响响应性能
		go func() {
			if err := lm.storage.Add(log); err != nil {
				// 这里可以记录到系统日志，但不影响主流程
				// 可以考虑添加一个错误回调函数
			}
		}()
	}
}

// extractTargetHost 从请求中提取目标主机
func extractTargetHost(r *http.Request) string {
	// 从查询参数中获取目标URL
	targetURL := r.URL.Query().Get("target")
	if targetURL == "" {
		return ""
	}

	// 简单解析主机名
	// 这里可以使用 url.Parse 进行更严格的解析
	if len(targetURL) > 8 { // 至少包含 "http://" 或 "https://"
		if targetURL[:7] == "http://" {
			targetURL = targetURL[7:]
		} else if targetURL[:8] == "https://" {
			targetURL = targetURL[8:]
		}
	}

	// 查找第一个斜杠，获取主机部分
	for i, char := range targetURL {
		if char == '/' || char == '?' || char == '#' {
			return targetURL[:i]
		}
	}

	return targetURL
}

// extractTargetPath 从请求中提取目标路径
func extractTargetPath(r *http.Request) string {
	// 从查询参数中获取目标URL
	targetURL := r.URL.Query().Get("target")
	if targetURL == "" {
		return ""
	}

	// 简单解析路径部分
	if len(targetURL) > 8 { // 至少包含 "http://" 或 "https://"
		if targetURL[:7] == "http://" {
			targetURL = targetURL[7:]
		} else if targetURL[:8] == "https://" {
			targetURL = targetURL[8:]
		}
	}

	// 查找第一个斜杠，获取路径部分
	for i, char := range targetURL {
		if char == '/' {
			return targetURL[i:]
		}
		if char == '?' || char == '#' {
			return "/"
		}
	}

	return "/"
}

// WrapHandler 便捷函数，包装单个处理器
func WrapHandler(handler http.HandlerFunc, storage Storage, maxBodySize int, record200 bool) http.HandlerFunc {
	middleware := NewLoggingMiddleware(storage, maxBodySize, record200)
	return middleware.Wrap(handler)
}

// ConditionalCapture 条件响应捕获器，只在特定条件下捕获响应体
type ConditionalCapture struct {
	*ResponseCapture
	shouldCapture func(statusCode int) bool
}

// NewConditionalCapture 创建条件响应捕获器
func NewConditionalCapture(w http.ResponseWriter, maxBodySize int, record200 bool, shouldCapture func(int) bool) *ConditionalCapture {
	return &ConditionalCapture{
		ResponseCapture: NewResponseCapture(w, true, maxBodySize, record200),
		shouldCapture:   shouldCapture,
	}
}

// Write 重写Write方法，添加条件判断
func (cc *ConditionalCapture) Write(data []byte) (int, error) {
	// 写入原始响应
	n, err := cc.ResponseCapture.ResponseWriter.Write(data)

	// 更新响应大小
	cc.bodySize += int64(n)

	// 根据条件决定是否捕获响应体
	if cc.captureBody && cc.shouldCapture(cc.statusCode) {
		// 检查是否超过最大大小限制
		if cc.body.Len()+len(data) <= cc.maxBodySize {
			cc.body.Write(data[:n])
		} else {
			// 只写入剩余可用空间
			remaining := cc.maxBodySize - cc.body.Len()
			if remaining > 0 {
				cc.body.Write(data[:remaining])
			}
		}
	}

	return n, err
}

// GetBody 重写GetBody方法，添加条件判断
func (cc *ConditionalCapture) GetBody() string {
	if !cc.captureBody || !cc.shouldCapture(cc.statusCode) {
		return ""
	}

	return cc.ResponseCapture.GetBody()
}

// detectContentType 检测响应内容类型
func (rc *ResponseCapture) detectContentType(data []byte) string {
	if len(data) == 0 {
		return "empty"
	}

	// 检查是否是gzip压缩
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		return "gzip"
	}

	// 检查是否是二进制数据（包含不可打印字符）
	for i, b := range data {
		if i > 512 { // 只检查前512字节
			break
		}
		if b < 32 && b != 9 && b != 10 && b != 13 { // 不是制表符、换行符、回车符的控制字符
			return "binary"
		}
	}

	// 检查是否是JSON
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return "json"
	}

	// 检查是否是XML/HTML
	if bytes.Contains(data[:min(len(data), 100)], []byte("<")) {
		return "xml"
	}

	// 默认为文本
	return "text"
}

// formatResponseBody 格式化响应体内容
func (rc *ResponseCapture) formatResponseBody(data []byte, contentType string) string {
	switch contentType {
	case "empty":
		return "[空响应]"
	case "gzip":
		return "[gzip压缩内容，大小: " + formatBytes(int64(len(data))) + "]"
	case "binary":
		return "[二进制内容，大小: " + formatBytes(int64(len(data))) + "]"
	case "json":
		// 尝试格式化JSON
		var jsonObj interface{}
		if err := json.Unmarshal(data, &jsonObj); err == nil {
			if formatted, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				return string(formatted)
			}
		}
		// 如果格式化失败，返回原始内容
		return string(data)
	case "xml":
		return string(data)
	case "text":
		return string(data)
	default:
		return string(data)
	}
}

// formatBytes 格式化字节大小
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
