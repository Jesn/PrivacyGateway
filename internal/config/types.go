package config

// ProxyAuth 代理认证信息
type ProxyAuth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	URL     string     `json:"url"`                // 代理服务器地址
	Type    string     `json:"type,omitempty"`     // 代理类型: http, socks5
	Auth    *ProxyAuth `json:"auth,omitempty"`     // 认证信息
	Timeout int        `json:"timeout,omitempty"`  // 超时时间(秒)
}

// Config 存储应用程序的配置
type Config struct {
	Port             string
	SensitiveHeaders []string
	DefaultProxy     *ProxyConfig // 默认代理配置
	ProxyWhitelist   []string     // 代理白名单
	AllowPrivateIP   bool         // 是否允许私有IP代理
}
