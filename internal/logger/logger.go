package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Logger 简单的结构化日志记录器
type Logger struct {
	*log.Logger
}

// New 创建新的日志记录器
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields ...interface{}) {
	logData := map[string]interface{}{
		"level":   "info",
		"message": msg,
		"time":    time.Now().Format(time.RFC3339),
	}
	
	// 添加字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logData[fields[i].(string)] = fields[i+1]
		}
	}
	
	jsonData, _ := json.Marshal(logData)
	l.Logger.Println(string(jsonData))
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, fields ...interface{}) {
	logData := map[string]interface{}{
		"level":   "error",
		"message": msg,
		"time":    time.Now().Format(time.RFC3339),
	}
	
	// 添加字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logData[fields[i].(string)] = fields[i+1]
		}
	}
	
	jsonData, _ := json.Marshal(logData)
	l.Logger.Println(string(jsonData))
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, fields ...interface{}) {
	logData := map[string]interface{}{
		"level":   "warn",
		"message": msg,
		"time":    time.Now().Format(time.RFC3339),
	}
	
	// 添加字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logData[fields[i].(string)] = fields[i+1]
		}
	}
	
	jsonData, _ := json.Marshal(logData)
	l.Logger.Println(string(jsonData))
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields ...interface{}) {
	logData := map[string]interface{}{
		"level":   "debug",
		"message": msg,
		"time":    time.Now().Format(time.RFC3339),
	}
	
	// 添加字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			logData[fields[i].(string)] = fields[i+1]
		}
	}
	
	jsonData, _ := json.Marshal(logData)
	l.Logger.Println(string(jsonData))
}
