package network

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"campus-keepalive/src/logger"
)

// LoginResult 登录结果
type LoginResult struct {
	Success bool
	Message string
}

// LoginManager 登录管理器
type LoginManager struct {
	server    string
	username  string
	password  string
	carrier   string
	logger    *logger.Logger
	client    *http.Client
	lastLogin time.Time // 上次登录时间
}

// NewLoginManager 创建登录管理器
func NewLoginManager(server, username, password, carrier string, log *logger.Logger) *LoginManager {
	return &LoginManager{
		server:   server,
		username: username,
		password: password,
		carrier:  carrier,
		logger:   log,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Login 登录（单次尝试）
func (m *LoginManager) Login() LoginResult {
	if m.username == "" || m.password == "" {
		return LoginResult{Success: false, Message: "账号或密码为空，请先配置"}
	}

	// 检查距离上次登录是否太频繁（至少间隔 10 秒）
	if time.Since(m.lastLogin) < 10*time.Second {
		m.logger.Warn("距离上次登录太近，跳过")
		return LoginResult{Success: false, Message: "登录太频繁，请稍后再试"}
	}

	suffix := getCarrierSuffix(m.carrier)
	DDDDD := m.username + suffix

	loginURL := fmt.Sprintf("http://%s:801/eportal/?c=ACSetting&a=Login&url=drappall", m.server)

	form := url.Values{}
	form.Set("DDDDD", DDDDD)
	form.Set("upass", m.password)

	m.logger.Info("发送登录请求: 账号=%s", m.username)
	m.lastLogin = time.Now()

	resp, err := m.client.Post(loginURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		m.logger.Error("登录请求失败: %v", err)
		return LoginResult{Success: false, Message: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	// 检查登录结果
	if strings.Contains(result, "success") || strings.Contains(result, "login_ok") || strings.Contains(result, "Dr.COMWebLoginID_3") {
		m.logger.Info("登录成功！")
		return LoginResult{Success: true, Message: "登录成功"}
	}

	m.logger.Warn("登录响应: %s", truncate(result, 200))
	return LoginResult{Success: false, Message: "登录失败，请检查账号密码"}
}

// Logout 注销
func (m *LoginManager) Logout() bool {
	logoutURL := fmt.Sprintf("http://%s:801/eportal/?c=ACSetting&a=Logout&ver=1.0", m.server)

	m.logger.Info("发送注销请求")

	resp, err := m.client.Get(logoutURL)
	if err != nil {
		m.logger.Error("注销请求失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	m.logger.Info("注销完成")
	return true
}

// SmartReconnect 智能重连（带退避策略）
func (m *LoginManager) SmartReconnect() LoginResult {
	// 1. 先等待 5 秒，让网络稳定
	m.logger.Info("等待 5 秒后开始重连...")
	time.Sleep(5 * time.Second)

	// 2. 尝试登录（不注销！）
	m.logger.Info("尝试登录...")
	result := m.Login()

	if result.Success {
		return result
	}

	// 3. 登录失败，等待 30 秒后再试
	m.logger.Warn("首次登录失败，等待 30 秒后重试...")
	time.Sleep(30 * time.Second)

	// 4. 第二次尝试
	m.logger.Info("第二次尝试登录...")
	return m.Login()
}

// RetryWithBackoff 带退避的重试（最多 3 次，间隔递增）
func (m *LoginManager) RetryWithBackoff(maxRetries int) LoginResult {
	// 等待 5 秒让网络稳定
	m.logger.Info("等待 5 秒后开始重连...")
	time.Sleep(5 * time.Second)

	for i := 0; i < maxRetries; i++ {
		m.logger.Info("重连尝试 %d/%d", i+1, maxRetries)

		result := m.Login()
		if result.Success {
			return result
		}

		// 失败后等待：10秒、20秒、30秒...
		waitSec := 10 * (i + 1)
		m.logger.Warn("登录失败，等待 %d 秒后重试...", waitSec)
		time.Sleep(time.Duration(waitSec) * time.Second)
	}

	m.logger.Error("连续 %d 次重连失败", maxRetries)
	return LoginResult{Success: false, Message: "连续失败，请检查网络或账号密码"}
}

// UpdateCredentials 更新登录凭据
func (m *LoginManager) UpdateCredentials(server, username, password, carrier string) {
	m.server = server
	m.username = username
	m.password = password
	m.carrier = carrier
}

// getCarrierSuffix 获取运营商后缀
func getCarrierSuffix(carrier string) string {
	switch carrier {
	case "dx":
		return "@dx"
	case "lt":
		return "@lt"
	default:
		return ""
	}
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
