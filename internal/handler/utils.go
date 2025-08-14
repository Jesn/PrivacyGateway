package handler

import "strings"

// IsSensitiveHeader 检查一个头信息是否是敏感的（不区分大小写）
func IsSensitiveHeader(headerKey string, sensitiveList []string) bool {
	lowerHeaderKey := strings.ToLower(headerKey)
	for _, sensitive := range sensitiveList {
		if strings.Contains(lowerHeaderKey, sensitive) {
			return true
		}
	}
	return false
}
