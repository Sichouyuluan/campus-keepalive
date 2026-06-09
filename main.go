package main

import (
	"fmt"
	"os"

	"campus-keepalive/src/config"
	"campus-keepalive/src/logger"
	"campus-keepalive/src/tray"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log := logger.New(cfg.LogFile)
	defer log.Close()
	log.Info("校园网自动保活工具启动")

	// 创建并启动托盘（阻塞）
	t := tray.New(cfg, log)
	t.Run()
}
