package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

// StatusResponse 状态查询响应
type StatusResponse struct {
	Result    int    `json:"result"`
	Msg       string `json:"msg"`
	RetCode   int    `json:"ret_code"`
	UserAccount string `json:"user_account"`
	UserIP      string `json:"user_ip"`
	OnlineTime  int    `json:"online_time"`
	Usage       int    `json:"usage"`
	UserName    string `json:"user_name"`
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
func (d *Detector) Detect() Status {
	// 方式1: 访问认证页面检查是否已登录
	url := fmt.Sprintf("http://%s/", d.server)

	resp, err := d.client.Get(url)
	if err != nil {
		// 网络完全不通
		return StatusOffline
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StatusOffline
	}

	content := string(body)

	// 已登录的特征：包含 uid='xxx'（非空）
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

// GetStatusInfo 获取详细状态信息
func (d *Detector) GetStatusInfo() *StatusResponse {
	url := fmt.Sprintf("http://%s/", d.server)

	resp, err := d.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	content := string(body)

	// 从页面中提取用户信息
	statusResp := &StatusResponse{}

	// 提取 uid
	re := regexp.MustCompile(`uid='([^']+)'`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 && match[1] != "" {
		statusResp.UserAccount = match[1]
		statusResp.Result = 1
	}

	// 提取 IP
	re = regexp.MustCompile(`lip='([^']+)'`)
	match = re.FindStringSubmatch(content)
	if len(match) > 1 {
		statusResp.UserIP = match[1]
	}

	// 提取姓名
	re = regexp.MustCompile(`NID='([^']+)'`)
	match = re.FindStringSubmatch(content)
	if len(match) > 1 {
		statusResp.UserName = match[1]
	}

	return statusResp
}

// parseStatusJSONP 解析状态查询 JSONP 响应
func parseStatusJSONP(jsonp string) (*StatusResponse, error) {
	re := regexp.MustCompile(`\{[^}]+\}`)
	match := re.FindString(jsonp)
	if match == "" {
		return nil, fmt.Errorf("无法解析 JSONP: %s", jsonp)
	}

	var resp StatusResponse
	if err := json.Unmarshal([]byte(match), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
