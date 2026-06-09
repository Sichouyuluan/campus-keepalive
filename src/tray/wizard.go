package tray

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"campus-keepalive/src/config"
	"campus-keepalive/src/logger"
)

// DetectResult 检测结果
type DetectResult struct {
	Server    string `json:"server"`
	Port      int    `json:"port"`
	System    string `json:"system"`
	Found     bool   `json:"found"`
}

// openWizard 打开配置向导（Web UI）
func openWizard(cfg *config.Config, log *logger.Logger, tray *Tray) {
	mux := http.NewServeMux()

	// 主页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := wizardHTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	// 检测 API
	mux.HandleFunc("/api/detect", func(w http.ResponseWriter, r *http.Request) {
		result := detectCampusNetwork()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"found":%v,"server":"%s","port":%d,"system":"%s"}`,
			result.Found, result.Server, result.Port, result.System)
	})

	// 测试登录 API（默认 API）
	mux.HandleFunc("/api/test-login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(r.Body)
		reqStr := string(body)

		server := extractJSON(reqStr, "server")
		username := extractJSON(reqStr, "username")
		password := extractJSON(reqStr, "password")
		carrier := extractJSON(reqStr, "carrier")

		success, message := testLoginAPI(server, username, password, carrier)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":%v,"message":"%s"}`, success, escapeJSON(message))
	})

	// 解析 curl 命令 API
	mux.HandleFunc("/api/parse-curl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(r.Body)
		reqStr := string(body)
		curlCmd := extractJSON(reqStr, "curl")

		result := parseCurlCommand(curlCmd)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":%v,"url":"%s","method":"%s","message":"%s"}`,
			result.Success, escapeJSON(result.URL), result.Method, escapeJSON(result.Message))
	})

	// 使用自定义 API 测试登录
	mux.HandleFunc("/api/test-custom-login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(r.Body)
		reqStr := string(body)

		apiURL := extractJSON(reqStr, "url")
		username := extractJSON(reqStr, "username")
		password := extractJSON(reqStr, "password")

		success, message := testCustomLoginAPI(apiURL, username, password)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":%v,"message":"%s"}`, success, escapeJSON(message))
	})

	// 保存配置 API
	mux.HandleFunc("/api/save-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, _ := io.ReadAll(r.Body)
		reqStr := string(body)

		server := extractJSON(reqStr, "server")
		username := extractJSON(reqStr, "username")
		password := extractJSON(reqStr, "password")
		carrier := extractJSON(reqStr, "carrier")
		customAPI := extractJSON(reqStr, "custom_api")

		if len(cfg.Accounts) == 0 {
			cfg.Accounts = append(cfg.Accounts, config.Account{})
		}

		cfg.Accounts[cfg.CurrentAccount].Username = username
		cfg.Accounts[cfg.CurrentAccount].Password = config.EncodePassword(password)
		cfg.Accounts[cfg.CurrentAccount].Carrier = carrier

		// 自动使用账号名作为名称
		if username != "" {
			cfg.Accounts[cfg.CurrentAccount].Name = username
		}

		parts := strings.Split(server, ":")
		if len(parts) >= 1 {
			cfg.Server = parts[0]
		}

		// 保存自定义 API（如果有）
		if customAPI != "" {
			cfg.CustomLoginAPI = customAPI
		}

		if err := config.Save(cfg); err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"success":false,"message":"%s"}`, escapeJSON(err.Error()))
			return
		}

		tray.UpdateConfig(cfg)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"message":"配置已保存"}`)
	})

	port := findPort(18082)
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Info("配置向导: %s", url)

	exec.Command("cmd", "/c", "start", url).Run()

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}

// ParseCurlResult 解析 curl 结果
type ParseCurlResult struct {
	Success bool
	URL     string
	Method  string
	Message string
}

// parseCurlCommand 解析 curl 命令
func parseCurlCommand(curlCmd string) ParseCurlResult {
	result := ParseCurlResult{}

	// 清理命令
	curlCmd = strings.TrimSpace(curlCmd)

	// 提取 URL
	// 格式: curl 'http://xxx' 或 curl "http://xxx" 或 curl http://xxx
	urlRegex := regexp.MustCompile(`(?:curl\s+(?:-[^\s]+\s+)*)['"]?(https?://[^'"\s\\]+)['"]?`)
	matches := urlRegex.FindStringSubmatch(curlCmd)
	if len(matches) < 2 {
		result.Message = "无法解析 URL，请确保粘贴的是完整的 curl 命令"
		return result
	}

	result.URL = matches[1]
	result.Method = "GET" // 默认 GET

	// 检查是否有 -d 参数（POST 请求）
	if strings.Contains(curlCmd, " -d ") || strings.Contains(curlCmd, " --data ") {
		result.Method = "POST"
	}

	result.Success = true
	result.Message = "解析成功"
	return result
}

// CustomLoginResult 自定义登录结果
type CustomLoginResult struct {
	Success bool
	Message string
}

// testCustomLoginAPI 测试自定义登录 API
func testCustomLoginAPI(apiURL, username, password string) (bool, string) {
	// 替换 URL 中的账号密码占位符
	apiURL = strings.Replace(apiURL, "{username}", url.QueryEscape(username), -1)
	apiURL = strings.Replace(apiURL, "{password}", url.QueryEscape(password), -1)

	// 如果 URL 中没有占位符，尝试添加参数
	if !strings.Contains(apiURL, username) {
		separator := "?"
		if strings.Contains(apiURL, "?") {
			separator = "&"
		}
		apiURL = apiURL + separator + "user_account=" + url.QueryEscape(username) + "&user_password=" + url.QueryEscape(password)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	// 检查是否成功
	if strings.Contains(result, "success") || strings.Contains(result, "result") || strings.Contains(result, "login_ok") {
		return true, result
	}

	return false, result
}

// detectCampusNetwork 检测校园网认证系统
func detectCampusNetwork() DetectResult {
	result := DetectResult{}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	testURLs := []string{
		"http://connect.rom.miui.com/generate_204",
		"http://captive.apple.com",
		"http://www.msftconnecttest.com/connecttest.txt",
		"http://www.baidu.com",
	}

	for _, testURL := range testURLs {
		resp, err := client.Get(testURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location != "" {
				result = extractServerFromURL(location)
				if result.Found {
					return result
				}
			}
		}

		body, _ := io.ReadAll(resp.Body)
		content := string(body)

		if strings.Contains(content, "eportal") || strings.Contains(content, "Dr.COM") {
			result = extractServerFromURL(testURL)
			result.Found = true
			result.System = "Dr.COM EPortal"
			return result
		}
	}

	commonAddresses := getCommonAddresses()
	for _, addr := range commonAddresses {
		if tryConnect(addr, 801) {
			testURL := fmt.Sprintf("http://%s:801/", addr)
			resp, err := client.Get(testURL)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			content := string(body)

			if strings.Contains(content, "eportal") || strings.Contains(content, "portal") {
				result.Server = addr
				result.Port = 801
				result.Found = true
				result.System = "Dr.COM EPortal"
				return result
			}
		}
	}

	gateway := getGatewayIP()
	if gateway != "" {
		if tryConnect(gateway, 801) {
			result.Server = gateway
			result.Port = 801
			result.Found = true
			result.System = "Unknown"
			return result
		}
	}

	return result
}

// extractServerFromURL 从 URL 提取服务器地址
func extractServerFromURL(urlStr string) DetectResult {
	result := DetectResult{}

	parts := strings.Split(urlStr, "://")
	if len(parts) < 2 {
		return result
	}

	hostPart := strings.Split(parts[1], "/")[0]
	hostParts := strings.Split(hostPart, ":")

	if len(hostParts) >= 1 {
		result.Server = hostParts[0]
	}
	if len(hostParts) >= 2 {
		fmt.Sscanf(hostParts[1], "%d", &result.Port)
	} else {
		result.Port = 80
	}

	if strings.Contains(urlStr, "eportal") {
		result.System = "Dr.COM EPortal"
	} else if strings.Contains(urlStr, "srun_portal") {
		result.System = "深澜 Srun"
	} else if strings.Contains(urlStr, "ruijie") {
		result.System = "锐捷 Ruijie"
	} else {
		result.System = "Unknown"
	}

	result.Found = true
	return result
}

// getCommonAddresses 获取常见认证服务器地址
func getCommonAddresses() []string {
	var addresses []string

	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				base := fmt.Sprintf("%d.%d.%d", ip[0], ip[1], ip[2])
				addresses = append(addresses, base+".1", base+".2", base+".254")
			}
		}
	}

	addresses = append(addresses, "10.0.0.1", "10.0.0.2", "172.16.0.1", "192.168.1.1")

	return addresses
}

// getGatewayIP 获取默认网关 IP
func getGatewayIP() string {
	cmd := exec.Command("cmd", "/c", "ipconfig")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "默认网关") || strings.Contains(line, "Default Gateway") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ip := strings.TrimSpace(parts[1])
				if ip != "" && ip != "0.0.0.0" {
					return ip
				}
			}
		}
	}

	return ""
}

// tryConnect 尝试连接
func tryConnect(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// testLoginAPI 测试登录 API
func testLoginAPI(server, username, password, carrier string) (bool, string) {
	suffix := ""
	switch carrier {
	case "dx":
		suffix = "@dx"
	case "lt":
		suffix = "@lt"
	}
	account := username + suffix

	loginURL := fmt.Sprintf("http://%s/eportal/portal/login?callback=dr1003&login_method=1&user_account=%s&user_password=%s&wlan_user_ip=&wlan_user_ipv6=&wlan_user_mac=000000000000&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.2.1&terminal_type=1&lang=zh-cn&v=%d&lang=zh",
		server, account, password, time.Now().UnixMilli())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(loginURL)
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	if strings.Contains(result, "success") || strings.Contains(result, "result") {
		return true, result
	}

	return false, result
}

// extractJSON 简单提取 JSON 值
func extractJSON(json, key string) string {
	search := fmt.Sprintf(`"%s":"`, key)
	idx := strings.Index(json, search)
	if idx == -1 {
		return ""
	}
	start := idx + len(search)
	end := strings.Index(json[start:], `"`)
	if end == -1 {
		return ""
	}
	return json[start : start+end]
}

// escapeJSON 转义 JSON 字符串
func escapeJSON(s string) string {
	s = strings.Replace(s, `\`, `\\`, -1)
	s = strings.Replace(s, `"`, `\"`, -1)
	s = strings.Replace(s, "\n", `\n`, -1)
	s = strings.Replace(s, "\r", `\r`, -1)
	s = strings.Replace(s, "\t", `\t`, -1)
	return s
}
