package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"campus-keepalive/src/logger"
)

// LoginResponse 登录响应
type LoginResponse struct {
	Result   int    `json:"result"`
	Msg      string `json:"msg"`
	RetCode  int    `json:"ret_code"`
}

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
	ip        string
	mac       string
	customAPI string        // 自定义登录 API
	logger    *logger.Logger
	client    *http.Client
	lastLogin time.Time
}

// NewLoginManager 创建登录管理器
func NewLoginManager(server, username, password, carrier string, log *logger.Logger) *LoginManager {
	ip := getLocalIP()
	mac := getLocalMAC()

	return &LoginManager{
		server:   server,
		username: username,
		password: password,
		carrier:  carrier,
		ip:       ip,
		mac:      mac,
		logger:   log,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// getLocalIP 获取本机 IP 地址
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "10.0.0.1" // 默认值
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getLocalMAC 获取本机 MAC 地址
func getLocalMAC() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "000000000000"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) > 0 {
			return strings.ReplaceAll(iface.HardwareAddr.String(), ":", "")
		}
	}
	return "000000000000"
}

// Login 登录
func (m *LoginManager) Login() LoginResult {
	if m.username == "" || m.password == "" {
		return LoginResult{Success: false, Message: "账号或密码为空，请先配置"}
	}

	// 检查距离上次登录是否太频繁（至少间隔 5 秒）
	if time.Since(m.lastLogin) < 5*time.Second {
		m.logger.Warn("距离上次登录太近，跳过")
		return LoginResult{Success: false, Message: "登录太频繁，请稍后再试"}
	}

	suffix := getCarrierSuffix(m.carrier)
	account := m.username + suffix

	// 构建登录 URL
	var loginURL string
	if m.customAPI != "" {
		// 使用自定义 API，替换占位符
		loginURL = m.customAPI
		loginURL = strings.Replace(loginURL, "{username}", account, -1)
		loginURL = strings.Replace(loginURL, "{password}", m.password, -1)
		loginURL = strings.Replace(loginURL, "{ip}", m.ip, -1)
		loginURL = strings.Replace(loginURL, "{mac}", m.mac, -1)
		// 替换 user_account 和 user_password（如果存在）
		loginURL = strings.Replace(loginURL, "user_account=2023405021", "user_account="+account, -1)
		loginURL = strings.Replace(loginURL, "user_password=28287X", "user_password="+m.password, -1)
		// 更新时间戳
		loginURL = regexp.MustCompile(`v=\d+`).ReplaceAllString(loginURL, fmt.Sprintf("v=%d", time.Now().UnixMilli()))
	} else {
		// 使用默认 API
		loginURL = fmt.Sprintf("http://%s:801/eportal/portal/login?callback=dr1003&login_method=1&user_account=%s&user_password=%s&wlan_user_ip=%s&wlan_user_ipv6=&wlan_user_mac=%s&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.2.1&terminal_type=1&lang=zh-cn&v=%d&lang=zh",
			m.server,
			account,
			m.password,
			m.ip,
			m.mac,
			time.Now().UnixMilli(),
		)
	}

	m.logger.Info("发送登录请求: 账号=%s", account)
	m.lastLogin = time.Now()

	resp, err := m.client.Get(loginURL)
	if err != nil {
		m.logger.Error("登录请求失败: %v", err)
		return LoginResult{Success: false, Message: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	m.logger.Info("登录响应: %s", result)

	// 解析 JSONP 响应
	loginResp, err := parseJSONP(result)
	if err != nil {
		m.logger.Error("解析响应失败: %v", err)
		return LoginResult{Success: false, Message: "解析响应失败"}
	}

	// 判断登录结果
	// result=0 且 ret_code=2 表示已在线
	// result=1 表示成功
	if loginResp.Result == 1 || loginResp.RetCode == 2 {
		m.logger.Info("登录成功: %s", loginResp.Msg)
		return LoginResult{Success: true, Message: loginResp.Msg}
	}

	m.logger.Warn("登录失败: %s", loginResp.Msg)
	return LoginResult{Success: false, Message: loginResp.Msg}
}

// Logout 注销
func (m *LoginManager) Logout() bool {
	logoutURL := fmt.Sprintf("http://%s:801/eportal/portal/logout?callback=dr1004&login_method=1&user_account=%s&wlan_user_ip=%s&wlan_user_ipv6=&wlan_user_mac=%s&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.2.1&terminal_type=1&lang=zh-cn&v=%d&lang=zh",
		m.server,
		m.username,
		m.ip,
		m.mac,
		time.Now().UnixMilli(),
	)

	m.logger.Info("发送注销请求")

	resp, err := m.client.Get(logoutURL)
	if err != nil {
		m.logger.Error("注销请求失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	m.logger.Info("注销响应: %s", string(body))

	return true
}

// SmartReconnect 智能重连
func (m *LoginManager) SmartReconnect() LoginResult {
	m.logger.Info("开始重连，等待 3 秒...")
	time.Sleep(3 * time.Second)

	m.logger.Info("尝试登录...")
	result := m.Login()

	if result.Success {
		return result
	}

	m.logger.Warn("首次登录失败，等待 10 秒后重试...")
	time.Sleep(10 * time.Second)

	m.logger.Info("第二次尝试登录...")
	return m.Login()
}

// RetryWithBackoff 带退避的重试
func (m *LoginManager) RetryWithBackoff(maxRetries int) LoginResult {
	m.logger.Info("等待 3 秒后开始重连...")
	time.Sleep(3 * time.Second)

	for i := 0; i < maxRetries; i++ {
		m.logger.Info("重连尝试 %d/%d", i+1, maxRetries)

		result := m.Login()
		if result.Success {
			return result
		}

		waitSec := 5 * (i + 1)
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

// UpdateCustomAPI 更新自定义登录 API
func (m *LoginManager) UpdateCustomAPI(customAPI string) {
	m.customAPI = customAPI
}

// UpdateIP 更新 IP 地址
func (m *LoginManager) UpdateIP(ip string) {
	m.ip = ip
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

// parseJSONP 解析 JSONP 响应
// 输入: dr1003({"result":0,"msg":"xxx","ret_code":2})
// 输出: LoginResponse
func parseJSONP(jsonp string) (*LoginResponse, error) {
	// 提取 JSON 部分
	re := regexp.MustCompile(`\{[^}]+\}`)
	match := re.FindString(jsonp)
	if match == "" {
		return nil, fmt.Errorf("无法解析 JSONP: %s", jsonp)
	}

	var resp LoginResponse
	if err := json.Unmarshal([]byte(match), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
