package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Account 账号配置
type Account struct {
	Name     string `json:"name"`     // 账号别名
	Username string `json:"username"` // 账号
	Password string `json:"password"` // 密码（简单加密存储）
	Carrier  string `json:"carrier"`  // 运营商：campus/dx/lt/other
}

// Config 应用配置
type Config struct {
	Accounts            []Account `json:"accounts"`        // 账号列表
	CurrentAccount      int       `json:"current_account"` // 当前账号索引
	Server              string    `json:"server"`          // 认证服务器地址
	CheckInterval       int       `json:"check_interval"`  // 检测间隔（秒）
	AutoStart           bool      `json:"auto_start"`      // 开机自启
	Notification        bool      `json:"notification"`    // 弹窗通知
	DisableNotification bool      `json:"disable_notification"` // 禁用所有通知
	LogFile             string    `json:"log_file"`        // 日志文件路径
	CustomLoginAPI      string    `json:"custom_login_api,omitempty"` // 自定义登录 API
}

// configDir 配置文件目录
func configDir() string {
	appData := os.Getenv("APPDATA")
	return filepath.Join(appData, "campus-keepalive")
}

// configPath 配置文件路径
func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

// Load 加载配置
func Load() (*Config, error) {
	path := configPath()

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return createDefault()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if cfg.Server == "" {
		cfg.Server = "210.44.114.32"
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30
	}
	if cfg.LogFile == "" {
		cfg.LogFile = filepath.Join(configDir(), "keepalive.log")
	}

	return &cfg, nil
}

// Save 保存配置
func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath(), data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// createDefault 创建默认配置
func createDefault() (*Config, error) {
	cfg := &Config{
		Accounts: []Account{
			{
				Name:     "默认账号",
				Username: "",
				Password: "",
				Carrier:  "campus",
			},
		},
		CurrentAccount: 0,
		Server:         "210.44.114.32",
		CheckInterval:  30,
		AutoStart:      false,
		Notification:   true,
		LogFile:        filepath.Join(configDir(), "keepalive.log"),
	}

	if err := Save(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetCurrentAccount 获取当前账号
func GetCurrentAccount(cfg *Config) *Account {
	if cfg.CurrentAccount < 0 || cfg.CurrentAccount >= len(cfg.Accounts) {
		return nil
	}
	return &cfg.Accounts[cfg.CurrentAccount]
}

// GetCarrierSuffix 获取运营商后缀
func GetCarrierSuffix(carrier string) string {
	switch carrier {
	case "dx":
		return "@dx"
	case "lt":
		return "@lt"
	default:
		return ""
	}
}

// EncodePassword 简单异或加密密码
func EncodePassword(password string) string {
	key := byte(0x5A) // 简单的异或密钥
	result := make([]byte, len(password))
	for i, b := range []byte(password) {
		result[i] = b ^ key
	}
	return string(result)
}

// DecodePassword 解密密码
func DecodePassword(encoded string) string {
	return EncodePassword(encoded) // 异或加密是对称的
}
