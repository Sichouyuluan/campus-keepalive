package network

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Status 网络状态
type Status int

const (
	StatusOnline  Status = iota // 在线（已登录）
	StatusOffline               // 离线（未登录或网络不通）
	StatusUnknown               // 未知（检测超时）
)

// String 状态字符串
func (s Status) String() string {
	switch s {
	case StatusOnline:
		return "在线"
	case StatusOffline:
		return "离线"
	default:
		return "未知"
	}
}

// Detector 网络检测器
type Detector struct {
	server string
	client *http.Client
}

// NewDetector 创建检测器
func NewDetector(server string) *Detector {
	return &Detector{
		server: server,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Detect 检测网络状态
// 策略：HTTP 访问认证页面，检查返回内容判断是否已登录
func (d *Detector) Detect() Status {
	url := fmt.Sprintf("http://%s/", d.server)

	resp, err := d.client.Get(url)
	if err != nil {
		// 网络完全不通（没连上校园网，或服务器挂了）
		return StatusOffline
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StatusOffline
	}

	content := string(body)

	// 已登录的特征：
	// 1. 包含 uid='xxx'（用户账号）
	// 2. 包含 NID='xxx'（用户姓名）
	// 3. 页面标题是"用户信息页"
	if strings.Contains(content, "uid='") && !strings.Contains(content, "uid=''") {
		return StatusOnline
	}

	// 未登录（登录页面）
	return StatusOffline
}

// DetectWithTimeout 带超时的检测
func (d *Detector) DetectWithTimeout(timeout time.Duration) Status {
	ch := make(chan Status, 1)

	go func() {
		ch <- d.Detect()
	}()

	select {
	case status := <-ch:
		return status
	case <-time.After(timeout):
		return StatusUnknown
	}
}
