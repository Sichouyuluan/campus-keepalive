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
	server     string
	username   string
	password   string
	carrier    string
	logger     *logger.Logger
	client     *http.Client
	failCount  int
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

// Login 登录
func (m *LoginManager) Login() LoginResult {
	suffix := getCarrierSuffix(m.carrier)
	DDDDD := m.username + suffix

	loginURL := fmt.Sprintf("http://%s:801/eportal/?c=ACSetting&a=Login&url=drappall", m.server)

	// 构建 POST 表单数据
	form := url.Values{}
	form.Set("DDDDD", DDDDD)
	form.Set("upass", m.password)

	m.logger.Info("尝试登录: 账号=%s, 运营商=%s", m.username, m.carrier)

	resp, err := m.client.Post(loginURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		m.logger.Error("登录请求失败: %v", err)
		m.failCount++
		return LoginResult{Success: false, Message: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	// 检查登录结果
	if strings.Contains(result, "Dr.COMWebLoginID_3.htm") || strings.Contains(result, "success") || strings.Contains(result, "login_ok") {
		m.logger.Info("登录成功")
		m.failCount = 0
		return LoginResult{Success: true, Message: "登录成功"}
	}

	m.logger.Warn("登录响应: %s", truncate(result, 200))
	m.failCount++
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

// SmartReconnect 智能重连
// 注意：此方法只在确认离线后才调用，不要在已登录状态下调用
func (m *LoginManager) SmartReconnect() LoginResult {
	// 直接尝试登录（如果已登录，认证系统会返回成功或忽略）
	result := m.Login()
	if result.Success {
		return result
	}

	// 登录失败，可能需要先注销
	// 但注销会断开现有连接，所以只在登录确实失败时才注销
	m.logger.Info("登录失败，尝试注销后重新登录")
	m.Logout()
	time.Sleep(2 * time.Second)

	return m.Login()
}

// RetryWithBackoff 带退避的重试
func (m *LoginManager) RetryWithBackoff(maxRetries int) LoginResult {
	delays := []time.Duration{3 * time.Second, 5 * time.Second, 10 * time.Second, 10 * time.Second, 10 * time.Second}

	for i := 0; i < maxRetries; i++ {
		result := m.SmartReconnect()
		if result.Success {
			return result
		}

		if i < len(delays) {
			m.logger.Warn("重试 %d/%d，等待 %v", i+1, maxRetries, delays[i])
			time.Sleep(delays[i])
		}
	}

	m.logger.Error("连续 %d 次失败", maxRetries)
	return LoginResult{Success: false, Message: "连续失败，请检查网络或账号密码"}
}

// UpdateCredentials 更新登录凭据
func (m *LoginManager) UpdateCredentials(server, username, password, carrier string) {
	m.server = server
	m.username = username
	m.password = password
	m.carrier = carrier
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
