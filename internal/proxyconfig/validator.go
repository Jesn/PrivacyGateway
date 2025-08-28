package proxyconfig

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	// 子域名格式验证正则
	subdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

// ValidateConfig 验证配置
func ValidateConfig(config *ProxyConfig) error {
	if config.Name == "" {
		return errors.New("name is required")
	}

	if len(config.Name) > 100 {
		return errors.New("name too long (max 100 characters)")
	}

	if err := ValidateSubdomain(config.Subdomain); err != nil {
		return err
	}

	if err := ValidateTargetURL(config.TargetURL); err != nil {
		return err
	}

	if config.Protocol != "http" && config.Protocol != "https" {
		return errors.New("protocol must be http or https")
	}

	return nil
}

// ValidateSubdomain 验证子域名
func ValidateSubdomain(subdomain string) error {
	if subdomain == "" {
		return errors.New("subdomain is required")
	}

	subdomain = strings.ToLower(subdomain)

	if len(subdomain) < 1 || len(subdomain) > 63 {
		return errors.New("subdomain length must be 1-63 characters")
	}

	if !subdomainRegex.MatchString(subdomain) {
		return errors.New("invalid subdomain format")
	}

	// 检查保留子域名
	reserved := []string{"www", "api", "admin", "mail", "ftp", "localhost", "logs", "ws", "proxy"}
	for _, r := range reserved {
		if subdomain == r {
			return errors.New("subdomain is reserved")
		}
	}

	return nil
}

// ValidateTargetURL 验证目标URL
func ValidateTargetURL(targetURL string) error {
	if targetURL == "" {
		return errors.New("target_url is required")
	}

	u, err := url.Parse(targetURL)
	if err != nil {
		return errors.New("invalid target_url format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("target_url must use http or https")
	}

	if u.Host == "" {
		return errors.New("target_url must have a host")
	}

	return nil
}
