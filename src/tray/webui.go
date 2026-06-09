package tray

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"campus-keepalive/src/config"
	"campus-keepalive/src/logger"
)

// openSettings 打开设置页面（Web UI）
func openSettings(cfg *config.Config, log *logger.Logger, tray *Tray) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		account := config.GetCurrentAccount(cfg)
		if account == nil {
			account = &config.Account{}
		}

		carriers := []struct{ Name, Value string }{
			{"校园用户", "campus"},
			{"校园电信 (@dx)", "dx"},
			{"校园联通 (@lt)", "lt"},
			{"校园其他", "other"},
		}

		carrierOptions := ""
		for _, c := range carriers {
			selected := ""
			if c.Value == account.Carrier {
				selected = "selected"
			}
			carrierOptions += fmt.Sprintf("<option value='%s' %s>%s</option>", c.Value, selected, c.Name)
		}

		// 账号列表
		accountOptions := ""
		for i, a := range cfg.Accounts {
			selected := ""
			if i == cfg.CurrentAccount {
				selected = "selected"
			}
			name := a.Name
			if name == "" {
				name = a.Username
			}
			accountOptions += fmt.Sprintf("<option value='%d' %s>%s</option>", i, selected, name)
		}

		checkedAutoStart := ""
		if cfg.AutoStart {
			checkedAutoStart = "checked"
		}
		checkedNotify := ""
		if cfg.Notification {
			checkedNotify = "checked"
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>校园网保活 - 设置</title>
<style>
body { font-family: Microsoft YaHei, sans-serif; background: #1e1e2e; color: #cdd6f4; padding: 20px; max-width: 500px; margin: 0 auto; }
h2 { color: #89b4fa; border-bottom: 2px solid #313244; padding-bottom: 10px; }
label { display: block; margin: 10px 0 5px; color: #a6adc8; }
input[type=text], input[type=password], input[type=number], select { width: 100%%; padding: 8px; background: #313244; color: #cdd6f4; border: 1px solid #45475a; border-radius: 4px; box-sizing: border-box; }
.checkbox-row { display: flex; align-items: center; gap: 8px; margin: 10px 0; }
.checkbox-row input[type=checkbox] { width: auto; }
.btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin: 10px 5px 10px 0; }
.btn-save { background: #89b4fa; color: #1e1e2e; }
.btn-test { background: #a6e3a1; color: #1e1e2e; }
.btn-del { background: #f38ba8; color: #1e1e2e; }
.btn-new { background: #fab387; color: #1e1e2e; }
.btn:hover { opacity: 0.8; }
.msg { padding: 10px; margin: 10px 0; border-radius: 4px; display: none; }
.msg-ok { background: #a6e3a122; color: #a6e3a1; }
.msg-err { background: #f38ba822; color: #f38ba8; }
</style></head><body>
<h2>🔧 校园网保活工具 - 设置</h2>
<form method="POST" action="/save">
<label>当前账号</label>
<select name="account_index">%s</select>
<div style="margin:5px 0">
<button type="submit" name="action" value="new" class="btn btn-new">新增账号</button>
<button type="submit" name="action" value="delete" class="btn btn-del">删除当前</button>
</div>
<label>账号</label>
<input type="text" name="username" value="%s">
<label>密码</label>
<input type="password" name="password" value="%s">
<label>运营商</label>
<select name="carrier">%s</select>
<label>认证服务器</label>
<input type="text" name="server" value="%s">
<label>检测间隔（秒）</label>
<input type="number" name="interval" value="%d" min="5" max="300">
<div class="checkbox-row"><input type="checkbox" name="autostart" %s><label>开机自启</label></div>
<div class="checkbox-row"><input type="checkbox" name="notification" %s><label>弹窗通知</label></div>
<div>
<button type="submit" name="action" value="save" class="btn btn-save">保存设置</button>
<button type="submit" name="action" value="test" class="btn btn-test">测试登录</button>
</div>
</form>
<div id="msg" class="msg"></div>
<p style="color:#585b70;font-size:12px;">保存后自动生效，无需重启程序。</p>
</body></html>`,
			accountOptions,
			account.Username,
			config.DecodePassword(account.Password),
			carrierOptions,
			cfg.Server,
			cfg.CheckInterval,
			checkedAutoStart,
			checkedNotify,
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		r.ParseForm()

		action := r.FormValue("action")

		if action == "new" {
			cfg.Accounts = append(cfg.Accounts, config.Account{
				Name:    fmt.Sprintf("账号%d", len(cfg.Accounts)+1),
				Carrier: "campus",
			})
			cfg.CurrentAccount = len(cfg.Accounts) - 1
			config.Save(cfg)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if action == "delete" {
			if len(cfg.Accounts) > 1 {
				idx := cfg.CurrentAccount
				cfg.Accounts = append(cfg.Accounts[:idx], cfg.Accounts[idx+1:]...)
				if cfg.CurrentAccount >= len(cfg.Accounts) {
					cfg.CurrentAccount = len(cfg.Accounts) - 1
				}
				config.Save(cfg)
			}
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// 保存设置
		idx := cfg.CurrentAccount
		fmt.Sscanf(r.FormValue("account_index"), "%d", &idx)
		cfg.CurrentAccount = idx

		username := r.FormValue("username")
		cfg.Accounts[idx].Username = username
		cfg.Accounts[idx].Password = config.EncodePassword(r.FormValue("password"))
		cfg.Accounts[idx].Carrier = r.FormValue("carrier")

		// 自动使用账号名作为名称
		if username != "" {
			cfg.Accounts[idx].Name = username
		}

		cfg.Server = r.FormValue("server")
		fmt.Sscanf(r.FormValue("interval"), "%d", &cfg.CheckInterval)
		cfg.AutoStart = r.FormValue("autostart") == "on"
		cfg.Notification = r.FormValue("notification") == "on"

		config.Save(cfg)
		tray.UpdateConfig(cfg)
		log.Info("设置已保存")

		if action == "test" {
			http.Redirect(w, r, "/test", http.StatusFound)
			return
		}

		http.Redirect(w, r, "/?saved=1", http.StatusFound)
	})

	// 找一个可用端口
	port := findPort(18080)
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Info("设置页面: %s", url)

	// 打开浏览器
	exec.Command("cmd", "/c", "start", url).Run()

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}

// openStatus 打开状态页面（Web UI）
func openStatus(cfg *config.Config, log *logger.Logger, tray *Tray) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		statusText := "离线"
		statusColor := "#f38ba8"
		switch tray.GetStatus() {
		case StatusOnline:
			statusText = "在线"
			statusColor = "#a6e3a1"
		case StatusRetrying:
			statusText = "重连中"
			statusColor = "#f9e2af"
		}

		account := config.GetCurrentAccount(cfg)
		username := ""
		if account != nil {
			username = account.Username
		}

		// 获取最近日志
		entries := log.GetRecent(30)
		logHTML := ""
		for i := len(entries) - 1; i >= 0; i-- {
			e := entries[i]
			color := "#cdd6f4"
			switch e.Level {
			case logger.LevelWarn:
				color = "#f9e2af"
			case logger.LevelError:
				color = "#f38ba8"
			case logger.LevelDebug:
				color = "#585b70"
			}
			logHTML += fmt.Sprintf("<div style='color:%s;padding:2px 0;'>[%s] %s %s</div>",
				color,
				e.Time.Format("15:04:05"),
				e.Level,
				e.Message,
			)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>校园网保活 - 状态</title>
<meta http-equiv="refresh" content="5">
<style>
body { font-family: Microsoft YaHei, sans-serif; background: #1e1e2e; color: #cdd6f4; padding: 20px; max-width: 600px; margin: 0 auto; }
h2 { color: #89b4fa; border-bottom: 2px solid #313244; padding-bottom: 10px; }
.info { display: grid; grid-template-columns: 120px 1fr; gap: 8px; margin: 15px 0; }
.info-label { color: #a6adc8; }
.info-value { color: #cdd6f4; }
.status { font-size: 24px; font-weight: bold; color: %s; }
.log-box { background: #11111b; padding: 10px; border-radius: 4px; max-height: 400px; overflow-y: auto; font-family: Consolas, monospace; font-size: 12px; margin: 15px 0; }
.btn { padding: 10px 20px; background: #89b4fa; color: #1e1e2e; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
.btn:hover { opacity: 0.8; }
</style></head><body>
<h2>📊 校园网保活工具 - 状态</h2>
<div class="info">
<span class="info-label">当前状态</span><span class="status">%s</span>
<span class="info-label">账号</span><span class="info-value">%s</span>
<span class="info-label">认证服务器</span><span class="info-value">%s</span>
<span class="info-label">检测间隔</span><span class="info-value">%d 秒</span>
</div>
<h3>📋 最近日志</h3>
<div class="log-box">%s</div>
<form method="POST" action="/reconnect">
<button type="submit" class="btn">🔄 手动重连</button>
</form>
<p style="color:#585b70;font-size:12px;">页面每 5 秒自动刷新</p>
</body></html>`,
			statusColor,
			statusText,
			username,
			cfg.Server,
			cfg.CheckInterval,
			logHTML,
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	mux.HandleFunc("/reconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			go tray.ManualReconnect()
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	port := findPort(18081)
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	log.Info("状态页面: %s", url)

	exec.Command("cmd", "/c", "start", url).Run()

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux)
}

// findPort 从指定端口开始找一个可用端口
func findPort(start int) int {
	for port := start; port < start+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	return start
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
