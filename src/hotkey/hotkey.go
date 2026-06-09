package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	registerHotKey   = user32.NewProc("RegisterHotKey")
	unregisterHotKey = user32.NewProc("UnregisterHotKey")
	getMessage       = user32.NewProc("GetMessageW")
)

// MOD_* 修饰键常量
const (
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008
)

// VK_* 虚拟键码常量
const (
	VK_L = 0x4C
)

// MSG 消息结构
type MSG struct {
	HWND    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PT      struct{ X, Y int32 }
}

// HotkeyManager 全局快捷键管理器
type HotkeyManager struct {
	id     uintptr
	action func()
}

// New 创建快捷键管理器
func New(action func()) *HotkeyManager {
	return &HotkeyManager{
		id:     1,
		action: action,
	}
}

// Register 注册全局快捷键 Ctrl+Alt+L
func (h *HotkeyManager) Register() error {
	// 注册 Ctrl+Alt+L
	ret, _, err := registerHotKey.Call(0, h.id, MOD_CONTROL|MOD_ALT, VK_L)
	if ret == 0 {
		return fmt.Errorf("注册快捷键失败: %v", err)
	}
	return nil
}

// Unregister 注销快捷键
func (h *HotkeyManager) Unregister() {
	unregisterHotKey.Call(0, h.id)
}

// Listen 监听快捷键消息（阻塞）
func (h *HotkeyManager) Listen() {
	var msg MSG
	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		if msg.Message == 0x0312 && msg.WParam == h.id { // WM_HOTKEY
			if h.action != nil {
				h.action()
			}
		}
	}
}
