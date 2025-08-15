package accesslog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// GenerateLogID 生成唯一的日志ID
func GenerateLogID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetClientIP 从HTTP请求中获取客户端IP地址
func GetClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP地址
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 检查 X-Real-IP 头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// 使用 RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// TruncateBody 截断响应体内容
func TruncateBody(body []byte, maxSize int) string {
	if len(body) <= maxSize {
		return string(body)
	}
	if maxSize <= 13 { // len("...[truncated]") = 13
		return string(body[:maxSize])
	}
	return string(body[:maxSize-13]) + "...[truncated]"
}

// FormatDuration 格式化持续时间为人类可读的字符串
func FormatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	if ms < 60000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}
	return fmt.Sprintf("%.1fm", float64(ms)/60000)
}

// FormatSize 格式化字节大小为人类可读的字符串
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ParseStatusCodes 解析状态码字符串为整数切片
func ParseStatusCodes(statusStr string) []int {
	if statusStr == "" {
		return nil
	}

	var codes []int
	parts := strings.Split(statusStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 处理状态码范围，如 "2xx", "4xx", "5xx"
		switch part {
		case "2xx":
			for i := 200; i < 300; i++ {
				codes = append(codes, i)
			}
		case "3xx":
			for i := 300; i < 400; i++ {
				codes = append(codes, i)
			}
		case "4xx":
			for i := 400; i < 500; i++ {
				codes = append(codes, i)
			}
		case "5xx":
			for i := 500; i < 600; i++ {
				codes = append(codes, i)
			}
		default:
			// 尝试解析为具体的状态码
			var code int
			if _, err := fmt.Sscanf(part, "%d", &code); err == nil {
				if code >= 100 && code <= 599 {
					codes = append(codes, code)
				}
			}
		}
	}
	return codes
}

// ContainsStatusCode 检查状态码是否在指定的列表中
func ContainsStatusCode(codes []int, target int) bool {
	if len(codes) == 0 {
		return true // 空列表表示不筛选
	}
	for _, code := range codes {
		if code == target {
			return true
		}
	}
	return false
}

// IsWithinTimeRange 检查时间是否在指定范围内
func IsWithinTimeRange(t, from, to time.Time) bool {
	if from.IsZero() && to.IsZero() {
		return true // 没有时间限制
	}
	if !from.IsZero() && t.Before(from) {
		return false
	}
	if !to.IsZero() && t.After(to) {
		return false
	}
	return true
}

// MatchesDomain 检查主机名是否匹配域名筛选条件
func MatchesDomain(host, domain string) bool {
	if domain == "" {
		return true // 空域名表示不筛选
	}
	
	// 转换为小写进行比较
	host = strings.ToLower(host)
	domain = strings.ToLower(domain)
	
	// 精确匹配
	if host == domain {
		return true
	}
	
	// 子域名匹配
	if strings.HasSuffix(host, "."+domain) {
		return true
	}
	
	// 部分匹配
	return strings.Contains(host, domain)
}

// EstimateMemoryUsage 估算日志记录的内存使用量（字节）
func EstimateMemoryUsage(log *AccessLog) int64 {
	size := int64(0)
	
	// 基础结构体大小
	size += 8  // ID (string header)
	size += int64(len(log.ID))
	size += 24 // Timestamp (time.Time)
	size += 8  // Method (string header)
	size += int64(len(log.Method))
	size += 8  // TargetHost (string header)
	size += int64(len(log.TargetHost))
	size += 8  // TargetPath (string header)
	size += int64(len(log.TargetPath))
	size += 8  // StatusCode (int)
	size += 8  // ResponseBody (string header)
	size += int64(len(log.ResponseBody))
	size += 8  // UserAgent (string header)
	size += int64(len(log.UserAgent))
	size += 8  // ClientIP (string header)
	size += int64(len(log.ClientIP))
	size += 8  // Duration (int64)
	size += 8  // RequestSize (int64)
	size += 8  // ResponseSize (int64)
	
	return size
}
