package logviewer

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"privacygateway/internal/accesslog"
)

// FilterParams 筛选参数
type FilterParams struct {
	Domain     string    `json:"domain,omitempty"`      // 域名筛选
	StatusCode []int     `json:"status_code,omitempty"` // 状态码筛选
	FromTime   time.Time `json:"from_time,omitempty"`   // 开始时间
	ToTime     time.Time `json:"to_time,omitempty"`     // 结束时间
	Page       int       `json:"page"`                  // 页码
	Limit      int       `json:"limit"`                 // 每页条数
	SortBy     string    `json:"sort_by,omitempty"`     // 排序字段
	SortOrder  string    `json:"sort_order,omitempty"`  // 排序方向
	Search     string    `json:"search,omitempty"`      // 搜索关键词
}

// FilterBuilder 筛选器构建器
type FilterBuilder struct {
	params *FilterParams
}

// NewFilterBuilder 创建新的筛选器构建器
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		params: &FilterParams{
			Page:      1,
			Limit:     50,
			SortBy:    "timestamp",
			SortOrder: "desc",
		},
	}
}

// FromRequest 从HTTP请求构建筛选器
func (fb *FilterBuilder) FromRequest(r *http.Request) *FilterBuilder {
	query := r.URL.Query()

	// 域名筛选
	if domain := query.Get("domain"); domain != "" {
		fb.params.Domain = strings.TrimSpace(domain)
	}

	// 状态码筛选
	if statusStr := query.Get("status"); statusStr != "" {
		fb.params.StatusCode = parseStatusCodes(statusStr)
	}

	// 时间范围筛选
	if fromStr := query.Get("from"); fromStr != "" {
		if fromTime, err := parseTime(fromStr); err == nil {
			fb.params.FromTime = fromTime
		}
	}

	if toStr := query.Get("to"); toStr != "" {
		if toTime, err := parseTime(toStr); err == nil {
			fb.params.ToTime = toTime
		}
	}

	// 分页参数
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			fb.params.Page = page
		}
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			fb.params.Limit = limit
		}
	}

	// 排序参数
	if sortBy := query.Get("sort_by"); sortBy != "" {
		if isValidSortField(sortBy) {
			fb.params.SortBy = sortBy
		}
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		if sortOrder == "asc" || sortOrder == "desc" {
			fb.params.SortOrder = sortOrder
		}
	}

	// 搜索关键词
	if search := query.Get("search"); search != "" {
		fb.params.Search = strings.TrimSpace(search)
	}

	return fb
}

// Domain 设置域名筛选
func (fb *FilterBuilder) Domain(domain string) *FilterBuilder {
	fb.params.Domain = domain
	return fb
}

// StatusCode 设置状态码筛选
func (fb *FilterBuilder) StatusCode(codes ...int) *FilterBuilder {
	fb.params.StatusCode = codes
	return fb
}

// TimeRange 设置时间范围
func (fb *FilterBuilder) TimeRange(from, to time.Time) *FilterBuilder {
	fb.params.FromTime = from
	fb.params.ToTime = to
	return fb
}

// Page 设置页码
func (fb *FilterBuilder) Page(page int) *FilterBuilder {
	if page > 0 {
		fb.params.Page = page
	}
	return fb
}

// Limit 设置每页条数
func (fb *FilterBuilder) Limit(limit int) *FilterBuilder {
	if limit > 0 && limit <= 1000 {
		fb.params.Limit = limit
	}
	return fb
}

// Sort 设置排序
func (fb *FilterBuilder) Sort(field, order string) *FilterBuilder {
	if isValidSortField(field) {
		fb.params.SortBy = field
	}
	if order == "asc" || order == "desc" {
		fb.params.SortOrder = order
	}
	return fb
}

// Search 设置搜索关键词
func (fb *FilterBuilder) Search(keyword string) *FilterBuilder {
	fb.params.Search = keyword
	return fb
}

// Build 构建筛选器
func (fb *FilterBuilder) Build() *accesslog.LogFilter {
	return &accesslog.LogFilter{
		Domain:     fb.params.Domain,
		StatusCode: fb.params.StatusCode,
		FromTime:   fb.params.FromTime,
		ToTime:     fb.params.ToTime,
		Page:       fb.params.Page,
		Limit:      fb.params.Limit,
		Search:     fb.params.Search,
	}
}

// GetParams 获取筛选参数
func (fb *FilterBuilder) GetParams() *FilterParams {
	return fb.params
}

// ToQueryString 转换为查询字符串
func (fb *FilterBuilder) ToQueryString() string {
	values := url.Values{}

	if fb.params.Domain != "" {
		values.Set("domain", fb.params.Domain)
	}

	if len(fb.params.StatusCode) > 0 {
		statusStrs := make([]string, len(fb.params.StatusCode))
		for i, code := range fb.params.StatusCode {
			statusStrs[i] = strconv.Itoa(code)
		}
		values.Set("status", strings.Join(statusStrs, ","))
	}

	if !fb.params.FromTime.IsZero() {
		values.Set("from", fb.params.FromTime.Format(time.RFC3339))
	}

	if !fb.params.ToTime.IsZero() {
		values.Set("to", fb.params.ToTime.Format(time.RFC3339))
	}

	if fb.params.Page != 1 {
		values.Set("page", strconv.Itoa(fb.params.Page))
	}

	if fb.params.Limit != 50 {
		values.Set("limit", strconv.Itoa(fb.params.Limit))
	}

	if fb.params.SortBy != "timestamp" {
		values.Set("sort_by", fb.params.SortBy)
	}

	if fb.params.SortOrder != "desc" {
		values.Set("sort_order", fb.params.SortOrder)
	}

	if fb.params.Search != "" {
		values.Set("search", fb.params.Search)
	}

	return values.Encode()
}

// parseStatusCodes 解析状态码字符串
func parseStatusCodes(statusStr string) []int {
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

		// 处理状态码范围
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
			// 尝试解析为具体状态码
			if code, err := strconv.Atoi(part); err == nil {
				if code >= 100 && code <= 599 {
					codes = append(codes, code)
				}
			}
		}
	}

	return codes
}

// parseTime 解析时间字符串
func parseTime(timeStr string) (time.Time, error) {
	// 支持多种时间格式
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

// isValidSortField 检查排序字段是否有效
func isValidSortField(field string) bool {
	validFields := []string{
		"timestamp",
		"method",
		"target_host",
		"status_code",
		"duration",
		"client_ip",
	}

	for _, validField := range validFields {
		if field == validField {
			return true
		}
	}

	return false
}

// GetStatusCodeGroups 获取状态码分组
func GetStatusCodeGroups() map[string][]int {
	return map[string][]int{
		"2xx": {200, 201, 202, 204, 206},
		"3xx": {301, 302, 304, 307, 308},
		"4xx": {400, 401, 403, 404, 405, 409, 410, 422, 429},
		"5xx": {500, 501, 502, 503, 504, 505},
	}
}

// GetCommonDomains 获取常见域名列表（可以从日志中动态生成）
func GetCommonDomains(storage accesslog.Storage) []string {
	// 这里可以实现从存储中获取常见域名的逻辑
	// 暂时返回一些示例域名
	return []string{
		"httpbin.org",
		"api.github.com",
		"www.google.com",
		"api.example.com",
	}
}

// ValidateFilter 验证筛选参数
func ValidateFilter(params *FilterParams) error {
	if params.Page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}

	if params.Limit < 1 || params.Limit > 1000 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}

	if !params.FromTime.IsZero() && !params.ToTime.IsZero() {
		if params.FromTime.After(params.ToTime) {
			return fmt.Errorf("from_time must be before to_time")
		}
	}

	if params.SortBy != "" && !isValidSortField(params.SortBy) {
		return fmt.Errorf("invalid sort field: %s", params.SortBy)
	}

	if params.SortOrder != "" && params.SortOrder != "asc" && params.SortOrder != "desc" {
		return fmt.Errorf("sort order must be 'asc' or 'desc'")
	}

	return nil
}
