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
	StatusOnline  Status = iota // 在线
	StatusOffline               // 离线
	StatusUnknown               // 未知
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
	server  string
	client  *http.Client
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

// Detect 检测网络状态（通过 HTTP 访问认证页面）
func (d *Detector) Detect() Status {
	// 访问认证页面，检查是否已登录
	// 已登录的用户访问首页会看到"注销"按钮和用户信息
	// 未登录的用户会看到登录页面
	url := fmt.Sprintf("http://%s/", d.server)

	resp, err := d.client.Get(url)
	if err != nil {
		// 网络完全不通（没连上校园网）
		return StatusOffline
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StatusOffline
	}

	content := string(body)

	// 检查是否已登录
	// 已登录页面包含"注销"按钮和用户信息
	if strings.Contains(content, "注销") || strings.Contains(content, "logout") || strings.Contains(content, "NID=") {
		return StatusOnline
	}

	// 未登录页面（登录页面）
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
