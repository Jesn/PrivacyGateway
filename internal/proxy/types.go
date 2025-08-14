package proxy

import "privacygateway/internal/config"

// 重新导出配置类型，便于使用
type Config = config.ProxyConfig
type Auth = config.ProxyAuth
