# 🌐 校园网自动登录器 (Campus Keepalive)

> 一款轻量级校园网自动保活工具，告别频繁掉线烦恼。

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?style=flat&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/badge/Release-v0.1.15-orange)]()

---

## ✨ 功能特性

- 🔄 **自动保活** — 定时检测网络状态，掉线自动重连
- 🎯 **智能检测** — HTTP 探测认证页面，准确判断登录状态
- 👥 **多账号管理** — 支持多个账号切换，方便多人共用电脑
- 🔧 **配置向导** — 自动检测认证系统，引导 F12 抓包获取登录 API
- 🖥️ **系统托盘** — 常驻托盘，三种颜色图标显示状态（🟢在线 🔴离线 🟡重连中）
- 📊 **状态窗口** — 实时显示在线时长、流量、日志
- ⚙️ **Web 设置** — 浏览器打开设置页面，操作简单
- 🚀 **开机自启** — 一键设置开机自动启动
- 🔔 **Toast 通知** — Windows 原生通知，掉线/重连实时提醒
- 🪶 **轻量小巧** — 仅 ~14MB，内存占用低

---

## 📸 界面预览

### 系统托盘菜单
```
校园网自动登录器
状态: 2023000000 - 在线
上次更新: 5秒前
─────────────────
切换账号
  ✓ 2023000000
  └ + 新增账号
检测间隔 (30秒)
  ✓ 30秒
  └ 5秒 / 10秒 / 60秒
─────────────────
手动重连 | 配置向导 | 设置 | 状态窗口
─────────────────
开机自启 ✓
─────────────────
退出
```

### 配置向导
```
步骤 1: 自动检测 → 识别认证系统类型
步骤 2: 输入账号密码 → 测试登录
步骤 3: F12 抓包引导（备用）
步骤 4: 保存配置
```

---

## 🚀 快速开始

### 下载

从 [Releases](../../releases) 页面下载最新版本的 `campus-keepalive.exe`。

### 首次使用

1. 双击运行 `campus-keepalive.exe`
2. 右键托盘图标 → **配置向导**
3. 点击「开始检测」自动识别认证系统
4. 输入账号密码，点击「测试登录」
5. 测试成功后点击「保存配置」
6. 完成！程序会在后台自动保活

### 如果自动检测失败

配置向导会引导你通过 F12 抓包获取登录 API：

1. 打开浏览器 F12 → Network 标签
2. 勾选「Preserve log（保留日志）」
3. 访问任意网页，被重定向到登录页面
4. 输入账号密码登录
5. 找到包含 `login` 的请求 → 右键 → Copy as cURL
6. 粘贴到配置向导的文本框中
7. 点击「解析 cURL」→「测试登录」→「保存配置」

---

## 🏗️ 技术架构

```
campus-keepalive/
├── main.go                 # 程序入口
├── src/
│   ├── config/             # 配置管理
│   │   └── config.go       # 配置读写、多账号管理
│   ├── logger/             # 日志模块
│   │   └── logger.go       # 文件日志、日志轮转
│   ├── network/            # 网络模块
│   │   ├── detector.go     # 网络状态检测
│   │   └── login.go        # 登录/注销/重连逻辑
│   ├── tray/               # 系统托盘
│   │   ├── walk_tray.go    # 托盘菜单、状态管理
│   │   ├── webui.go        # Web 设置/状态页面
│   │   ├── wizard.go       # 配置向导后端
│   │   └── wizard_html.go  # 配置向导前端
│   └── hotkey/             # 全局快捷键
│       └── hotkey.go       # Ctrl+Alt+L 手动重连
├── winres/                 # Windows 资源文件
│   ├── winres.json         # 版本信息、图标
│   └── icon.ico            # 应用图标
├── plan/                   # 项目计划
│   ├── requirements.md     # 需求文档
│   └── plan.md             # 执行计划
└── ui-comparison/          # UI 框架对比资料
```

### 技术栈

| 组件 | 技术 | 用途 |
|------|------|------|
| 语言 | Go 1.24+ | 主程序 |
| 托盘 | fyne.io/systray | 系统托盘菜单 |
| 通知 | gen2brain/beeep | Windows Toast 通知 |
| 资源 | go-winres | exe 版本信息、图标 |
| 设置 | Go net/http | 本地 Web 服务器 |

---

## 🔌 支持的认证系统

| 系统 | 状态 | 说明 |
|------|------|------|
| Dr.COM EPortal | ✅ 完全支持 | JSONP 接口，自动检测 |
| 锐捷 Ruijie | 🔜 计划中 | 需要适配 API |
| 深澜 Srun | 🔜 计划中 | 需要适配 API |
| 其他 | 📝 手动配置 | 通过 F12 抓包获取 API |

---

## ⚙️ 配置文件

配置文件位置：`%APPDATA%/campus-keepalive/config.json`

```json
{
  "accounts": [
    {
      "name": "2023000000",
      "username": "2023000000",
      "password": "加密后的密码",
      "carrier": "campus"
    }
  ],
  "current_account": 0,
  "server": "210.44.114.32",
  "check_interval": 30,
  "auto_start": false,
  "notification": true,
  "log_file": "%APPDATA%/campus-keepalive/keepalive.log",
  "custom_login_api": ""
}
```

### 运营商后缀

| 选项 | 值 | 说明 |
|------|-----|------|
| 校园用户 | `campus` | 无后缀 |
| 校园电信 | `dx` | 账号后加 `@dx` |
| 校园联通 | `lt` | 账号后加 `@lt` |
| 校园其他 | `other` | 无后缀 |

---

## 🛠️ 开发构建

### 环境要求

- Go 1.24+
- Windows 10/11

### 构建命令

```bash
# 安装依赖
go mod tidy

# 生成 Windows 资源文件
go-winres make

# 编译（无黑窗口）
go build -ldflags "-H windowsgui" -o campus-keepalive.exe .
```

### 项目结构

- **v0.1.0** — 初始版本：系统托盘、网络检测、自动登录框架
- **v0.1.1** — 修复网络检测、添加 Web 设置/状态窗口
- **v0.1.4** — 修复登录 API：使用正确的 EPortal JSONP 接口
- **v0.1.7** — 添加配置向导：自动检测认证系统
- **v0.1.8** — 配置向导添加 F12 抓包引导
- **v0.1.11** — 使用 beeep 显示 Windows Toast 通知
- **v0.1.15** — 程序启动时自动更新账号名称

---

## 📝 TODO

- [ ] 支持锐捷认证系统
- [ ] 支持深澜认证系统
- [ ] 流量统计图表
- [ ] 自定义托盘图标
- [ ] 多语言支持
- [ ] 自动更新检查

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

---

## 📄 许可证

本项目基于 MIT 许可证开源 - 详见 [LICENSE](LICENSE) 文件

---

## 🙏 致谢

- [fyne.io/systray](https://github.com/fyne-io/systray) — 跨平台系统托盘库
- [gen2brain/beeep](https://github.com/gen2brain/beeep) — 跨平台通知库
- [go-winres](https://github.com/tc-hib/go-winres) — Windows 资源文件生成工具

---

## 📮 反馈与联系

- **问题反馈表单**: [点击提交反馈](https://qcnq86kqcocv.feishu.cn/share/base/form/shrcnlxrdR9L1kRKaY7JdcZjhuf)
- Issues: [GitHub Issues](../../issues)
- Discussions: [GitHub Discussions](../../discussions)

---

<p align="center">
  <b>如果这个项目对你有帮助，请给一个 ⭐ Star 支持一下！</b>
</p>
