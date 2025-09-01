package metrics

import (
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	
	if m == nil {
		t.Fatal("Expected metrics to be created")
	}
	
	// 验证初始值
	snapshot := m.GetSnapshot()
	if snapshot.TotalRequests != 0 {
		t.Errorf("Expected initial total requests to be 0, got %d", snapshot.TotalRequests)
	}
	
	if snapshot.MinResponseTime != int64(^uint64(0)>>1) {
		t.Errorf("Expected initial min response time to be max int64")
	}
}

func TestRecordRequest(t *testing.T) {
	m := NewMetrics()
	
	// 记录成功请求
	m.RecordRequest(100*time.Millisecond, true)
	m.RecordRequest(200*time.Millisecond, true)
	m.RecordRequest(150*time.Millisecond, false)
	
	snapshot := m.GetSnapshot()
	
	// 验证总请求数
	if snapshot.TotalRequests != 3 {
		t.Errorf("Expected total requests to be 3, got %d", snapshot.TotalRequests)
	}
	
	// 验证成功请求数
	if snapshot.SuccessRequests != 2 {
		t.Errorf("Expected success requests to be 2, got %d", snapshot.SuccessRequests)
	}
	
	// 验证错误请求数
	if snapshot.ErrorRequests != 1 {
		t.Errorf("Expected error requests to be 1, got %d", snapshot.ErrorRequests)
	}
	
	// 验证成功率
	expectedSuccessRate := float64(2) / float64(3) * 100
	if snapshot.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate to be %.2f, got %.2f", expectedSuccessRate, snapshot.SuccessRate)
	}
	
	// 验证平均响应时间
	expectedAvgTime := (100 + 200 + 150) / 3
	if snapshot.AvgResponseTime != int64(expectedAvgTime) {
		t.Errorf("Expected avg response time to be %d, got %d", expectedAvgTime, snapshot.AvgResponseTime)
	}
	
	// 验证最小响应时间
	if snapshot.MinResponseTime != 100 {
		t.Errorf("Expected min response time to be 100, got %d", snapshot.MinResponseTime)
	}
	
	// 验证最大响应时间
	if snapshot.MaxResponseTime != 200 {
		t.Errorf("Expected max response time to be 200, got %d", snapshot.MaxResponseTime)
	}
}

func TestRecordTokenValidation(t *testing.T) {
	m := NewMetrics()
	
	m.RecordTokenValidation()
	m.RecordTokenValidation()
	m.RecordTokenValidation()
	
	snapshot := m.GetSnapshot()
	
	if snapshot.TokenValidations != 3 {
		t.Errorf("Expected token validations to be 3, got %d", snapshot.TokenValidations)
	}
}

func TestUpdateTokenCount(t *testing.T) {
	m := NewMetrics()
	
	m.UpdateTokenCount(10, 8)
	
	snapshot := m.GetSnapshot()
	
	if snapshot.TotalTokens != 10 {
		t.Errorf("Expected total tokens to be 10, got %d", snapshot.TotalTokens)
	}
	
	if snapshot.ActiveTokens != 8 {
		t.Errorf("Expected active tokens to be 8, got %d", snapshot.ActiveTokens)
	}
}

func TestUpdateConfigCount(t *testing.T) {
	m := NewMetrics()
	
	m.UpdateConfigCount(5, 4)
	
	snapshot := m.GetSnapshot()
	
	if snapshot.TotalConfigs != 5 {
		t.Errorf("Expected total configs to be 5, got %d", snapshot.TotalConfigs)
	}
	
	if snapshot.ActiveConfigs != 4 {
		t.Errorf("Expected active configs to be 4, got %d", snapshot.ActiveConfigs)
	}
}

func TestReset(t *testing.T) {
	m := NewMetrics()
	
	// 记录一些数据
	m.RecordRequest(100*time.Millisecond, true)
	m.RecordTokenValidation()
	m.UpdateTokenCount(10, 8)
	
	// 重置
	m.Reset()
	
	snapshot := m.GetSnapshot()
	
	// 验证所有计数器都被重置
	if snapshot.TotalRequests != 0 {
		t.Errorf("Expected total requests to be 0 after reset, got %d", snapshot.TotalRequests)
	}
	
	if snapshot.TokenValidations != 0 {
		t.Errorf("Expected token validations to be 0 after reset, got %d", snapshot.TokenValidations)
	}
	
	// 注意：令牌和配置计数不会被重置，因为它们是当前状态而不是累计值
}

func TestGetHealthStatus(t *testing.T) {
	m := NewMetrics()
	
	// 记录一些正常的请求
	for i := 0; i < 100; i++ {
		m.RecordRequest(50*time.Millisecond, true)
	}
	
	health := m.GetHealthStatus()
	
	if health.Status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got '%s'", health.Status)
	}
	
	// 检查错误率检查
	if check, exists := health.Checks["error_rate"]; !exists {
		t.Error("Expected error_rate check to exist")
	} else if check.Status != "ok" {
		t.Errorf("Expected error_rate check status to be 'ok', got '%s'", check.Status)
	}
	
	// 检查响应时间检查
	if check, exists := health.Checks["response_time"]; !exists {
		t.Error("Expected response_time check to exist")
	} else if check.Status != "ok" {
		t.Errorf("Expected response_time check status to be 'ok', got '%s'", check.Status)
	}
}

func TestGetHealthStatusDegraded(t *testing.T) {
	m := NewMetrics()
	
	// 记录高错误率的请求
	for i := 0; i < 90; i++ {
		m.RecordRequest(50*time.Millisecond, true)
	}
	for i := 0; i < 20; i++ {
		m.RecordRequest(50*time.Millisecond, false)
	}
	
	health := m.GetHealthStatus()
	
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got '%s'", health.Status)
	}
	
	// 检查错误率检查
	if check, exists := health.Checks["error_rate"]; !exists {
		t.Error("Expected error_rate check to exist")
	} else if check.Status != "warning" {
		t.Errorf("Expected error_rate check status to be 'warning', got '%s'", check.Status)
	}
}

func TestGetHealthStatusHighResponseTime(t *testing.T) {
	m := NewMetrics()
	
	// 记录高响应时间的请求
	for i := 0; i < 10; i++ {
		m.RecordRequest(1500*time.Millisecond, true)
	}
	
	health := m.GetHealthStatus()
	
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got '%s'", health.Status)
	}
	
	// 检查响应时间检查
	if check, exists := health.Checks["response_time"]; !exists {
		t.Error("Expected response_time check to exist")
	} else if check.Status != "warning" {
		t.Errorf("Expected response_time check status to be 'warning', got '%s'", check.Status)
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	
	// 并发记录请求
	done := make(chan bool, 100)
	
	for i := 0; i < 100; i++ {
		go func() {
			m.RecordRequest(100*time.Millisecond, true)
			m.RecordTokenValidation()
			done <- true
		}()
	}
	
	// 等待所有goroutine完成
	for i := 0; i < 100; i++ {
		<-done
	}
	
	snapshot := m.GetSnapshot()
	
	if snapshot.TotalRequests != 100 {
		t.Errorf("Expected total requests to be 100, got %d", snapshot.TotalRequests)
	}
	
	if snapshot.TokenValidations != 100 {
		t.Errorf("Expected token validations to be 100, got %d", snapshot.TokenValidations)
	}
}

func BenchmarkRecordRequest(b *testing.B) {
	m := NewMetrics()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.RecordRequest(100*time.Millisecond, true)
		}
	})
}

func BenchmarkGetSnapshot(b *testing.B) {
	m := NewMetrics()
	
	// 预先记录一些数据
	for i := 0; i < 1000; i++ {
		m.RecordRequest(100*time.Millisecond, true)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetSnapshot()
	}
}
