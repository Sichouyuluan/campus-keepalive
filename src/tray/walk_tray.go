package tray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/systray"
	"github.com/gen2brain/beeep"

	"campus-keepalive/src/config"
	"campus-keepalive/src/logger"
	"campus-keepalive/src/network"
)

// Status 托盘状态
type Status int

const (
	StatusOnline   Status = iota // 在线（绿色图标）
	StatusOffline                // 离线（红色图标）
	StatusRetrying               // 重连中（黄色图标）
)

// Tray 托盘管理器
type Tray struct {
	cfg      *config.Config
	log      *logger.Logger
	detector *network.Detector
	loginMgr *network.LoginManager
	status   Status

	// 菜单项
	mTitle     *systray.MenuItem // 标题
	mStatus    *systray.MenuItem // 状态
	mUpdate    *systray.MenuItem // 上次更新时间
	mAccounts  *systray.MenuItem // 切换账号
	mInterval  *systray.MenuItem // 检测间隔
	mReconnect *systray.MenuItem // 手动重连
	mAutoStart *systray.MenuItem // 开机自启

	// 子菜单项
	accountItems []*systray.MenuItem // 账号列表
	intervalItems []*systray.MenuItem // 间隔选项

	// 状态跟踪
	lastUpdate time.Time // 上次更新时间
	stopCh     chan struct{}
}

// New 创建托盘管理器
func New(cfg *config.Config, log *logger.Logger) *Tray {
	account := config.GetCurrentAccount(cfg)
	if account == nil {
		account = &config.Account{}
	}

	return &Tray{
		cfg:      cfg,
		log:      log,
		detector: network.NewDetector(cfg.Server),
		loginMgr: network.NewLoginManager(cfg.Server, account.Username, config.DecodePassword(account.Password), account.Carrier, log),
		status:   StatusOffline,
		lastUpdate: time.Now(),
		stopCh:   make(chan struct{}),
	}
}

// Run 启动系统托盘（阻塞）
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

// onReady 托盘初始化
func (t *Tray) onReady() {
	t.log.Info("托盘初始化开始")

	// 设置图标（初始红色离线状态）
	systray.SetIcon(createIcon(StatusOffline))
	systray.SetTitle("校园网自动登录器")
	systray.SetTooltip("校园网自动登录器")

	// === 第一行：标题 ===
	t.mTitle = systray.AddMenuItem("校园网自动登录器", "校园网自动登录器")

	// === 第二行：状态（账号 + 是否在线）===
	account := config.GetCurrentAccount(t.cfg)
	statusText := "离线"
	if account != nil {
		statusText = fmt.Sprintf("%s - %s", account.Username, statusText)
	}
	t.mStatus = systray.AddMenuItem("状态: "+statusText, "当前状态")

	// === 第三行：上次更新时间 ===
	t.mUpdate = systray.AddMenuItem("上次更新: 刚刚", "状态更新时间")

	systray.AddSeparator()

	// === 切换账号（子菜单）===
	t.mAccounts = systray.AddMenuItem("切换账号", "选择账号")
	t.updateAccountMenu()

	// === 检测间隔（子菜单）===
	t.mInterval = systray.AddMenuItem("检测间隔", "设置检测间隔")
	t.updateIntervalMenu()

	systray.AddSeparator()

	// === 手动重连 ===
	t.mReconnect = systray.AddMenuItem("手动重连", "立即尝试重新连接")

	// === 配置向导 ===
	mWizard := systray.AddMenuItem("配置向导", "首次使用配置")

	// === 设置 ===
	mSettings := systray.AddMenuItem("设置", "打开设置窗口")

	// === 状态窗口 ===
	mStatusWin := systray.AddMenuItem("状态窗口", "查看详细状态")

	systray.AddSeparator()

	// === 开机自启 ===
	t.mAutoStart = systray.AddMenuItem("开机自启", "开机自动启动")
	if t.cfg.AutoStart {
		t.mAutoStart.Check()
	}

	systray.AddSeparator()

	// === 退出 ===
	mQuit := systray.AddMenuItem("退出", "退出程序")

	// 监听菜单事件
	go t.handleMenuEvents(mWizard, mSettings, mStatusWin, mQuit)

	// 启动时立即检测一次
	t.log.Info("启动首次网络检测...")
	t.checkAndReconnect()

	// 显示启动通知
	showWindowsNotification("校园网自动登录器", "程序已启动，在托盘中查看状态")

	// 启动定时检测
	t.log.Info("启动定时检测，间隔: %d 秒", t.cfg.CheckInterval)
	go t.checkLoop()

	// 启动更新时间显示
	go t.updateTimeDisplay()
}

// onExit 托盘退出
func (t *Tray) onExit() {
	close(t.stopCh)
	t.log.Info("程序退出")
}

// updateAccountMenu 更新账号子菜单
func (t *Tray) updateAccountMenu() {
	// 清除旧的子菜单（如果有的话）
	// fyne.io/systray 不支持动态删除菜单项，所以我们只在初始化时创建

	t.accountItems = make([]*systray.MenuItem, len(t.cfg.Accounts))
	for i, account := range t.cfg.Accounts {
		name := account.Name
		if name == "" {
			name = account.Username
		}
		if name == "" {
			name = fmt.Sprintf("账号%d", i+1)
		}

		// 标记当前账号
		prefix := "  "
		if i == t.cfg.CurrentAccount {
			prefix = "✓ "
		}

		t.accountItems[i] = t.mAccounts.AddSubMenuItem(prefix+name, "切换到 "+name)
	}

	// 添加"新增账号"选项
	addNew := t.mAccounts.AddSubMenuItem("+ 新增账号", "添加新账号")
	go func() {
		<-addNew.ClickedCh
		t.cfg.Accounts = append(t.cfg.Accounts, config.Account{
			Name:    fmt.Sprintf("账号%d", len(t.cfg.Accounts)+1),
			Carrier: "campus",
		})
		config.Save(t.cfg)
		t.log.Info("新增账号")
	}()
}

// updateIntervalMenu 更新检测间隔子菜单
func (t *Tray) updateIntervalMenu() {
	intervals := []int{5, 10, 30, 60}
	t.intervalItems = make([]*systray.MenuItem, len(intervals))

	for i, interval := range intervals {
		prefix := "  "
		if interval == t.cfg.CheckInterval {
			prefix = "✓ "
		}

		t.intervalItems[i] = t.mInterval.AddSubMenuItem(
			fmt.Sprintf("%s%d秒", prefix, interval),
			fmt.Sprintf("设置检测间隔为 %d 秒", interval),
		)
	}
}

// handleMenuEvents 处理菜单事件
func (t *Tray) handleMenuEvents(mWizard, mSettings, mStatusWin, mQuit *systray.MenuItem) {
	// 账号切换事件
	for i, item := range t.accountItems {
		go func(idx int, ch <-chan struct{}) {
			for range ch {
				t.switchAccount(idx)
			}
		}(i, item.ClickedCh)
	}

	// 检测间隔事件
	intervals := []int{5, 10, 30, 60}
	for i, item := range t.intervalItems {
		go func(idx int, ch <-chan struct{}) {
			for range ch {
				t.setInterval(intervals[idx])
			}
		}(i, item.ClickedCh)
	}

	// 其他菜单事件
	for {
		select {
		case <-t.mReconnect.ClickedCh:
			t.log.Info("点击：手动重连")
			go t.ManualReconnect()
		case <-mWizard.ClickedCh:
			t.log.Info("点击：配置向导")
			go openWizard(t.cfg, t.log, t)
		case <-mSettings.ClickedCh:
			t.log.Info("点击：设置")
			go openSettings(t.cfg, t.log, t)
		case <-mStatusWin.ClickedCh:
			t.log.Info("点击：状态窗口")
			go openStatus(t.cfg, t.log, t)
		case <-t.mAutoStart.ClickedCh:
			t.cfg.AutoStart = !t.cfg.AutoStart
			if t.cfg.AutoStart {
				t.mAutoStart.Check()
			} else {
				t.mAutoStart.Uncheck()
			}
			config.Save(t.cfg)
			t.log.Info("开机自启: %v", t.cfg.AutoStart)
		case <-mQuit.ClickedCh:
			t.log.Info("点击：退出")
			systray.Quit()
			return
		}
	}
}

// switchAccount 切换账号
func (t *Tray) switchAccount(idx int) {
	if idx < 0 || idx >= len(t.cfg.Accounts) {
		return
	}

	t.cfg.CurrentAccount = idx
	config.Save(t.cfg)

	account := &t.cfg.Accounts[idx]
	t.loginMgr.UpdateCredentials(t.cfg.Server, account.Username, config.DecodePassword(account.Password), account.Carrier)

	t.log.Info("切换到账号: %s", account.Username)

	// 更新状态显示
	t.updateStatusDisplay()
}

// setInterval 设置检测间隔
func (t *Tray) setInterval(seconds int) {
	t.cfg.CheckInterval = seconds
	config.Save(t.cfg)

	t.log.Info("检测间隔设置为: %d 秒", seconds)

	// 更新间隔显示
	t.updateIntervalDisplay()
}

// updateStatusDisplay 更新状态显示
func (t *Tray) updateStatusDisplay() {
	account := config.GetCurrentAccount(t.cfg)
	statusText := "离线"
	switch t.status {
	case StatusOnline:
		statusText = "在线"
	case StatusRetrying:
		statusText = "重连中..."
	}

	if account != nil {
		statusText = fmt.Sprintf("%s - %s", account.Username, statusText)
	}

	t.mStatus.SetTitle("状态: " + statusText)
}

// updateIntervalDisplay 更新间隔显示
func (t *Tray) updateIntervalDisplay() {
	t.mInterval.SetTitle(fmt.Sprintf("检测间隔 (%d秒)", t.cfg.CheckInterval))
}

// updateTimeDisplay 定时更新时间显示
func (t *Tray) updateTimeDisplay() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(t.lastUpdate)
			timeText := formatDuration(elapsed)
			t.mUpdate.SetTitle("上次更新: " + timeText)
		case <-t.stopCh:
			return
		}
	}
}

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 5 {
		return "刚刚"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d秒前", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%d分钟前", minutes)
	}
	hours := minutes / 60
	return fmt.Sprintf("%d小时前", hours)
}

// checkLoop 定时检测循环
func (t *Tray) checkLoop() {
	interval := t.cfg.CheckInterval
	if interval < 5 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	t.log.Info("定时检测循环已启动，每 %d 秒检测一次", interval)

	for {
		select {
		case <-ticker.C:
			t.log.Debug("定时检测触发")
			t.checkAndReconnect()
		case <-t.stopCh:
			t.log.Info("定时检测循环停止")
			return
		}
	}
}

// checkAndReconnect 检测并重连
func (t *Tray) checkAndReconnect() {
	t.log.Info("开始检测网络状态...")
	status := t.detector.Detect()

	// 更新时间
	t.lastUpdate = time.Now()

	switch status {
	case network.StatusOnline:
		t.log.Info("检测结果: 在线")
		if t.status != StatusOnline {
			t.setStatus(StatusOnline)
		}
	case network.StatusOffline:
		t.log.Warn("检测结果: 离线，开始重连")
		t.setStatus(StatusRetrying)

		result := t.loginMgr.RetryWithBackoff(3)
		if result.Success {
			t.setStatus(StatusOnline)
			t.log.Info("重连成功！")
			t.showNotification("重连成功", "校园网已重新连接")
		} else {
			t.setStatus(StatusOffline)
			t.log.Error("重连失败: %s", result.Message)
			t.showNotification("重连失败", result.Message)
		}
	case network.StatusUnknown:
		t.log.Warn("检测结果: 超时/未知")
	}

	// 更新状态显示
	t.updateStatusDisplay()
}

// setStatus 更新托盘状态
func (t *Tray) setStatus(status Status) {
	t.status = status

	switch status {
	case StatusOnline:
		systray.SetIcon(createIcon(StatusOnline))
		systray.SetTooltip("校园网自动登录器 - 在线")
	case StatusOffline:
		systray.SetIcon(createIcon(StatusOffline))
		systray.SetTooltip("校园网自动登录器 - 离线")
	case StatusRetrying:
		systray.SetIcon(createIcon(StatusRetrying))
		systray.SetTooltip("校园网自动登录器 - 重连中...")
	}
}

// showNotification 显示通知
func (t *Tray) showNotification(title, message string) {
	if t.cfg.Notification {
		systray.SetTooltip(fmt.Sprintf("校园网自动登录器 - %s: %s", title, message))
		showWindowsNotification(title, message)
	}
}

// showWindowsNotification 显示 Windows Toast 通知
func showWindowsNotification(title, message string) {
	beeep.Notify(title, message, "")
}

// ManualReconnect 手动重连
func (t *Tray) ManualReconnect() {
	t.log.Info("手动触发重连")
	t.checkAndReconnect()
}

// UpdateConfig 更新配置
func (t *Tray) UpdateConfig(cfg *config.Config) {
	t.cfg = cfg
	t.detector = network.NewDetector(cfg.Server)

	account := config.GetCurrentAccount(cfg)
	if account != nil {
		t.loginMgr.UpdateCredentials(cfg.Server, account.Username, config.DecodePassword(account.Password), account.Carrier)
	}
}

// Stop 停止
func (t *Tray) Stop() {
	systray.Quit()
}

// GetStatus 获取当前状态
func (t *Tray) GetStatus() Status {
	return t.status
}

// ========== 图标生成 ==========

// createIcon 根据状态创建 ICO 格式图标
func createIcon(status Status) []byte {
	size := 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	var circleColor color.RGBA
	switch status {
	case StatusOnline:
		circleColor = color.RGBA{R: 76, G: 175, B: 80, A: 255} // 绿色
	case StatusOffline:
		circleColor = color.RGBA{R: 244, G: 67, B: 54, A: 255} // 红色
	case StatusRetrying:
		circleColor = color.RGBA{R: 255, G: 193, B: 7, A: 255} // 黄色
	}

	centerX, centerY := size/2, size/2
	radius := size/2 - 1

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := dx*dx + dy*dy
			if dist <= radius*radius {
				img.Set(x, y, circleColor)
			} else {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			}
		}
	}

	return imageToICO(img)
}

// imageToICO 将 RGBA 图像转换为 ICO 格式字节
func imageToICO(img *image.RGBA) []byte {
	bounds := img.Bounds()
	size := bounds.Dx()

	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, uint16(0))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(1))

	pixelData := make([]byte, size*size*4)
	maskData := make([]byte, ((size+31)/32)*4*size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			targetY := size - 1 - y
			srcOff := y*size + x
			dstOff := targetY*size + x

			r := img.Pix[srcOff*4+0]
			g := img.Pix[srcOff*4+1]
			b := img.Pix[srcOff*4+2]
			a := img.Pix[srcOff*4+3]

			pixelData[dstOff*4+0] = b
			pixelData[dstOff*4+1] = g
			pixelData[dstOff*4+2] = r
			pixelData[dstOff*4+3] = a

			if a < 128 {
				maskByteOff := targetY*((size+31)/32)*4 + x/8
				maskData[maskByteOff] |= 1 << (7 - uint(x%8))
			}
		}
	}

	bitmapInfoSize := 40
	imageDataSize := len(pixelData) + len(maskData)
	offset := 6 + 16

	buf.WriteByte(byte(size))
	buf.WriteByte(byte(size))
	buf.WriteByte(0)
	buf.WriteByte(0)
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(imageDataSize+bitmapInfoSize))
	binary.Write(&buf, binary.LittleEndian, uint32(offset))

	binary.Write(&buf, binary.LittleEndian, uint32(40))
	binary.Write(&buf, binary.LittleEndian, int32(size))
	binary.Write(&buf, binary.LittleEndian, int32(size*2))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(imageDataSize))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, int32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	binary.Write(&buf, binary.LittleEndian, uint32(0))

	buf.Write(pixelData)
	buf.Write(maskData)

	return buf.Bytes()
}
