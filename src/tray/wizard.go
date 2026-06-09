package tray

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"campus-keepalive/src/config"
	"campus-keepalive/src/logger"
)

// DetectResult 检测结果
type DetectResult struct {
	Server    string `json:"server"`     // 认证服务器地址
	Port      int    `json:"port"`       // 端口
	System    string `json:"system"`     // 认证系统类型
	LoginPage string `json:"login_page"` // 登录页面路径
	LoginAPI  string `json:"login_api"`  // 登录 API 路径
	Found     bool   `json:"found"`      // 是否找到
}

// openWizard 打开配置向导（Web UI）
func openWizard(cfg *config.Config, log *logger.Logger, tray *Tray) {
	mux := http.NewServeMux()

	// 主页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>校园网自动登录器 - 配置向导</title>
<style>
body { font-family: Microsoft YaHei, sans-serif; background: #1e1e2e; color: #cdd6f4; padding: 20px; max-width: 600px; margin: 0 auto; }
h2 { color: #89b4fa; border-bottom: 2px solid #313244; padding-bottom: 10px; }
.step { background: #313244; padding: 15px; border-radius: 8px; margin: 15px 0; }
.step-title { color: #89b4fa; font-weight: bold; margin-bottom: 10px; }
.btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin: 5px; }
.btn-primary { background: #89b4fa; color: #1e1e2e; }
.btn-success { background: #a6e3a1; color: #1e1e2e; }
.btn-warning { background: #f9e2af; color: #1e1e2e; }
.btn:hover { opacity: 0.8; }
.result { padding: 10px; margin: 10px 0; border-radius: 4px; }
.result-ok { background: #a6e3a122; color: #a6e3a1; }
.result-err { background: #f38ba822; color: #f38ba8; }
.result-warn { background: #f9e2af22; color: #f9e2af; }
input[type=text], input[type=password] { width: 100%; padding: 8px; background: #45475a; color: #cdd6f4; border: 1px solid #585b70; border-radius: 4px; box-sizing: border-box; }
label { display: block; margin: 10px 0 5px; color: #a6adc8; }
.spinner { display: inline-block; width: 16px; height: 16px; border: 2px solid #585b70; border-top: 2px solid #89b4fa; border-radius: 50%; animation: spin 1s linear infinite; }
@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
</style></head><body>
<h2>🔧 校园网自动登录器 - 配置向导</h2>

<div class="step">
<div class="step-title">步骤 1: 自动检测</div>
<p>程序将自动检测校园网认证系统。</p>
<button class="btn btn-primary" onclick="startDetect()">开始检测</button>
<div id="detect-result"></div>
</div>

<div class="step" id="step2" style="display:none;">
<div class="step-title">步骤 2: 输入账号密码</div>
<label>认证服务器</label>
<input type="text" id="server" value="">
<label>账号</label>
<input type="text" id="username" value="">
<label>密码</label>
<input type="password" id="password" value="">
<label>运营商</label>
<select id="carrier" style="width:100%;padding:8px;background:#45475a;color:#cdd6f4;border:1px solid #585b70;border-radius:4px;">
<option value="campus">校园用户</option>
<option value="dx">校园电信 (@dx)</option>
<option value="lt">校园联通 (@lt)</option>
<option value="other">校园其他</option>
</select>
<div style="margin-top:15px;">
<button class="btn btn-success" onclick="testLogin()">测试登录</button>
<button class="btn btn-primary" onclick="saveConfig()">保存配置</button>
</div>
<div id="test-result"></div>
</div>

<div class="step" id="step3" style="display:none;">
<div class="step-title">✅ 配置完成</div>
<p>配置已保存！程序将自动保活校园网连接。</p>
<p>您可以关闭此页面，程序会在后台运行。</p>
</div>

<script>
async function startDetect() {
    document.getElementById('detect-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在检测...</div>';

    try {
        const resp = await fetch('/api/detect');
        const data = await resp.json();

        if (data.found) {
            document.getElementById('detect-result').innerHTML =
                '<div class="result result-ok">✓ 检测成功！<br>认证系统: ' + data.system + '<br>服务器: ' + data.server + ':' + data.port + '</div>';
            document.getElementById('server').value = data.server + ':' + data.port;
            document.getElementById('step2').style.display = 'block';
        } else {
            document.getElementById('detect-result').innerHTML =
                '<div class="result result-err">✗ 未检测到认证系统。请确保您已连接校园网，或手动输入服务器地址。</div>';
            document.getElementById('step2').style.display = 'block';
        }
    } catch (e) {
        document.getElementById('detect-result').innerHTML =
            '<div class="result result-err">✗ 检测失败: ' + e.message + '</div>';
        document.getElementById('step2').style.display = 'block';
    }
}

async function testLogin() {
    const server = document.getElementById('server').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const carrier = document.getElementById('carrier').value;

    if (!username || !password) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">请输入账号和密码</div>';
        return;
    }

    document.getElementById('test-result').innerHTML = '<div class="result result-warn"><span class="spinner"></span> 正在测试登录...</div>';

    try {
        const resp = await fetch('/api/test-login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({server, username, password, carrier})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('test-result').innerHTML = '<div class="result result-ok">✓ 登录成功！' + data.message + '</div>';
        } else {
            document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 登录失败: ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}

async function saveConfig() {
    const server = document.getElementById('server').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const carrier = document.getElementById('carrier').value;

    if (!username || !password) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">请输入账号和密码</div>';
        return;
    }

    try {
        const resp = await fetch('/api/save-config', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({server, username, password, carrier})
        });
        const data = await resp.json();

        if (data.success) {
            document.getElementById('step2').style.display = 'none';
            document.getElementById('step3').style.display = 'block';
        } else {
            document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 保存失败: ' + data.message + '</div>';
        }
    } catch (e) {
        document.getElementById('test-result').innerHTML = '<div class="result result-err">✗ 请求失败: ' + e.message + '</div>';
    }
}
</script>
</body></html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	// 检测 API
	mux.HandleFunc("/api/detect", func(w http.ResponseWriter, r *http.Request) {
		result := detectCampusNetwork()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"found":%v,"server":"%s","port":%d,"system":"%s","login_page":"%s","login_api":"%s"}`,
			result.Found, result.Server, result.Port, result.System, result.LoginPage, result.LoginAPI)
	})

	// 测试登录 API
	mux.HandleFunc("/api/test-login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 解析请求
		body, _ := io.ReadAll(r.Body)
		reqStr := string(body)

		// 简单解析 JSON
		server := extractJSON(reqStr, "server")
		username := extractJSON(reqStr, "username")
		password := extractJSON(reqStr, "password")
		carrier := extractJSON(reqStr, "carrier")

		// 测试登录
		success, message := testLoginAPI(server, username, password, carrier)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":%v,"message":"%s"}`, success, message)
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

		// 更新配置
		if len(cfg.Accounts) == 0 {
			cfg.Accounts = append(cfg.Accounts, config.Account{})
		}

		cfg.Accounts[cfg.CurrentAccount].Username = username
		cfg.Accounts[cfg.CurrentAccount].Password = config.EncodePassword(password)
		cfg.Accounts[cfg.CurrentAccount].Carrier = carrier

		// 解析服务器地址和端口
		parts := strings.Split(server, ":")
		if len(parts) >= 1 {
			cfg.Server = parts[0]
		}

		if err := config.Save(cfg); err != nil {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"success":false,"message":"%s"}`, err.Error())
			return
		}

		// 更新托盘配置
		tray.UpdateConfig(cfg)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"message":"配置已保存"}`)
	})

	// 找一个可用端口
	port := findPort(18082)
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Info("配置向导: %s", url)

	// 打开浏览器
	exec.Command("cmd", "/c", "start", url).Run()

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}

// detectCampusNetwork 检测校园网认证系统
func detectCampusNetwork() DetectResult {
	result := DetectResult{}

	// 1. 尝试访问 HTTP 网站，看是否被重定向
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 不跟随重定向，记录重定向 URL
			return http.ErrUseLastResponse
		},
	}

	// 尝试访问一个 HTTP 网站
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

		// 检查是否被重定向
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location != "" {
				// 从重定向 URL 提取认证服务器地址
				result = extractServerFromURL(location)
				if result.Found {
					return result
				}
			}
		}

		// 检查响应内容是否包含认证页面特征
		body, _ := io.ReadAll(resp.Body)
		content := string(body)

		if strings.Contains(content, "eportal") || strings.Contains(content, "Dr.COM") {
			result = extractServerFromURL(testURL)
			result.Found = true
			result.System = "Dr.COM EPortal"
			return result
		}
	}

	// 2. 尝试常见认证服务器地址
	commonAddresses := getCommonAddresses()
	for _, addr := range commonAddresses {
		if tryConnect(addr, 801) {
			// 尝试访问该地址
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
				result.LoginPage = "/eportal/"
				result.LoginAPI = "/eportal/portal/login"
				return result
			}
		}

		if tryConnect(addr, 80) {
			testURL := fmt.Sprintf("http://%s/", addr)
			resp, err := client.Get(testURL)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			content := string(body)

			if strings.Contains(content, "eportal") || strings.Contains(content, "portal") || strings.Contains(content, "login") {
				result.Server = addr
				result.Port = 80
				result.Found = true
				result.System = "Dr.COM EPortal"
				result.LoginPage = "/"
				result.LoginAPI = "/portal/login"
				return result
			}
		}
	}

	// 3. 尝试获取网关地址
	gateway := getGatewayIP()
	if gateway != "" {
		if tryConnect(gateway, 801) {
			result.Server = gateway
			result.Port = 801
			result.Found = true
			result.System = "Unknown"
			return result
		}
		if tryConnect(gateway, 80) {
			result.Server = gateway
			result.Port = 80
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

	// 解析 URL
	// http://210.44.114.32:801/eportal/...
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

	// 检测系统类型
	if strings.Contains(urlStr, "eportal") {
		result.System = "Dr.COM EPortal"
		result.LoginAPI = "/eportal/portal/login"
	} else if strings.Contains(urlStr, "srun_portal") {
		result.System = "深澜 Srun"
		result.LoginAPI = "/cgi-bin/srun_portal"
	} else if strings.Contains(urlStr, "ruijie") {
		result.System = "锐捷 Ruijie"
		result.LoginAPI = "/srun_portal_pc"
	} else {
		result.System = "Unknown"
	}

	result.Found = true
	return result
}

// getCommonAddresses 获取常见认证服务器地址
func getCommonAddresses() []string {
	var addresses []string

	// 获取本机 IP
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				// 添加同网段的常见地址
				base := fmt.Sprintf("%d.%d.%d", ip[0], ip[1], ip[2])
				addresses = append(addresses,
					base+".1",
					base+".2",
					base+".254",
					base+".100",
				)
			}
		}
	}

	// 添加一些常见的认证服务器地址
	addresses = append(addresses,
		"10.0.0.1",
		"10.0.0.2",
		"172.16.0.1",
		"192.168.1.1",
		"192.168.0.1",
	)

	return addresses
}

// getGatewayIP 获取默认网关 IP
func getGatewayIP() string {
	// Windows: 使用 ipconfig 命令
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

// tryConnect 尝试连接指定地址和端口
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
	// 构建登录 URL
	suffix := ""
	switch carrier {
	case "dx":
		suffix = "@dx"
	case "lt":
		suffix = "@lt"
	}
	account := username + suffix

	// 尝试 Dr.COM EPortal 格式
	loginURL := fmt.Sprintf("http://%s/eportal/portal/login?callback=dr1003&login_method=1&user_account=%s&user_password=%s&wlan_user_ip=&wlan_user_ipv6=&wlan_user_mac=000000000000&wlan_ac_ip=&wlan_ac_name=&jsVersion=4.2.1&terminal_type=1&lang=zh-cn&v=%d&lang=zh",
		server,
		account,
		password,
		time.Now().UnixMilli(),
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(loginURL)
	if err != nil {
		return false, fmt.Sprintf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	result := string(body)

	// 检查结果
	if strings.Contains(result, "success") || strings.Contains(result, "result") {
		return true, result
	}

	return false, result
}

// extractJSON 简单提取 JSON 值
func extractJSON(json, key string) string {
	// 查找 "key":"value" 模式
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
