package tray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/systray"

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
	mStatus   *systray.MenuItem
	mReconnect *systray.MenuItem
	mAutoStart *systray.MenuItem

	// 通道
	stopCh chan struct{}
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
		stopCh:   make(chan struct{}),
	}
}

// Run 启动系统托盘（阻塞）
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

// onReady 托盘初始化
func (t *Tray) onReady() {
	// 设置图标（初始红色离线状态）
	systray.SetIcon(createIcon(StatusOffline))
	systray.SetTitle("校园网保活")
	systray.SetTooltip("校园网自动保活工具 - 启动中...")

	// 菜单项
	t.mStatus = systray.AddMenuItem("状态: 离线", "当前网络状态")
	t.mStatus.Disable()

	systray.AddSeparator()

	t.mReconnect = systray.AddMenuItem("手动重连", "立即尝试重新连接")
	mSettings := systray.AddMenuItem("设置", "打开设置窗口")
	mStatusWin := systray.AddMenuItem("状态窗口", "查看详细状态")

	systray.AddSeparator()

	t.mAutoStart = systray.AddMenuItem("开机自启", "开机自动启动")
	if t.cfg.AutoStart {
		t.mAutoStart.Check()
	}

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("退出", "退出程序")

	// 监听菜单事件
	go t.handleMenuEvents(t.mReconnect, mSettings, mStatusWin, t.mAutoStart, mQuit)

	// 启动网络检测
	go t.checkLoop()

	// 启动时立即检测一次
	go t.checkAndReconnect()
}

// onExit 托盘退出
func (t *Tray) onExit() {
	close(t.stopCh)
	t.log.Info("程序退出")
}

// handleMenuEvents 处理菜单事件
func (t *Tray) handleMenuEvents(mReconnect, mSettings, mStatusWin, mAutoStart, mQuit *systray.MenuItem) {
	for {
		select {
		case <-mReconnect.ClickedCh:
			go t.ManualReconnect()
		case <-mSettings.ClickedCh:
			// TODO: 打开设置窗口
			t.log.Info("点击设置（功能开发中）")
		case <-mStatusWin.ClickedCh:
			// TODO: 打开状态窗口
			t.log.Info("点击状态窗口（功能开发中）")
		case <-mAutoStart.ClickedCh:
			t.cfg.AutoStart = !t.cfg.AutoStart
			if t.cfg.AutoStart {
				mAutoStart.Check()
			} else {
				mAutoStart.Uncheck()
			}
			config.Save(t.cfg)
			t.log.Info("开机自启: %v", t.cfg.AutoStart)
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

// checkLoop 定时检测循环
func (t *Tray) checkLoop() {
	ticker := time.NewTicker(time.Duration(t.cfg.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.checkAndReconnect()
		case <-t.stopCh:
			return
		}
	}
}

// checkAndReconnect 检测并重连
func (t *Tray) checkAndReconnect() {
	status := t.detector.Detect()

	switch status {
	case network.StatusOnline:
		if t.status != StatusOnline {
			t.log.Info("网络在线")
			t.setStatus(StatusOnline)
		}
	case network.StatusOffline:
		t.log.Warn("网络掉线，开始重连")
		t.setStatus(StatusRetrying)

		result := t.loginMgr.RetryWithBackoff(5)
		if result.Success {
			t.setStatus(StatusOnline)
			t.log.Info("重连成功")
			t.showNotification("重连成功", "校园网已重新连接")
		} else {
			t.setStatus(StatusOffline)
			t.log.Error("重连失败: %s", result.Message)
			t.showNotification("重连失败", result.Message)
		}
	case network.StatusUnknown:
		t.log.Warn("网络检测超时")
	}
}

// setStatus 更新托盘状态
func (t *Tray) setStatus(status Status) {
	t.status = status

	switch status {
	case StatusOnline:
		systray.SetIcon(createIcon(StatusOnline))
		systray.SetTooltip("校园网保活工具 - 在线")
		t.mStatus.SetTitle("状态: 在线")
	case StatusOffline:
		systray.SetIcon(createIcon(StatusOffline))
		systray.SetTooltip("校园网保活工具 - 离线")
		t.mStatus.SetTitle("状态: 离线")
	case StatusRetrying:
		systray.SetIcon(createIcon(StatusRetrying))
		systray.SetTooltip("校园网保活工具 - 重连中...")
		t.mStatus.SetTitle("状态: 重连中...")
	}
}

// showNotification 显示通知
func (t *Tray) showNotification(title, message string) {
	if t.cfg.Notification {
		systray.SetTooltip(fmt.Sprintf("%s - %s", title, message))
	}
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

	// 填充圆形
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

	// ICO Header
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Type: ICO
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Count: 1

	// 像素数据（BGRA 格式，从底部开始）
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

	// Directory Entry
	buf.WriteByte(byte(size))
	buf.WriteByte(byte(size))
	buf.WriteByte(0)
	buf.WriteByte(0)
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(32))
	binary.Write(&buf, binary.LittleEndian, uint32(imageDataSize+bitmapInfoSize))
	binary.Write(&buf, binary.LittleEndian, uint32(offset))

	// BITMAPINFOHEADER
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
