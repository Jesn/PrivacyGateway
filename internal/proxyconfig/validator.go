package proxyconfig

import (
	"errors"
	"net/url"
)

// ValidateConfig 验证配置
func ValidateConfig(config *ProxyConfig) error {
	if config.Name == "" {
		return errors.New("name is required")
	}

	if len(config.Name) > 100 {
		return errors.New("name too long (max 100 characters)")
	}

	if err := ValidateTargetURL(config.TargetURL); err != nil {
		return err
	}

	if config.Protocol != "http" && config.Protocol != "https" {
		return errors.New("protocol must be http or https")
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
