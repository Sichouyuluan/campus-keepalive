package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelInfo Level = iota
	LevelWarn
	LevelError
)

// String 日志级别字符串
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time
	Level   Level
	Message string
}

// Logger 日志模块
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	logFile  string
	recent   []LogEntry // 最近日志（内存中保留最近 100 条）
	maxRecent int
}

// New 创建日志模块
func New(logFile string) *Logger {
	l := &Logger{
		logFile:   logFile,
		recent:    make([]LogEntry, 0, 100),
		maxRecent: 100,
	}

	// 确保日志目录存在
	dir := filepath.Dir(logFile)
	os.MkdirAll(dir, 0755)

	// 打开日志文件
	l.openFile()

	return l
}

// openFile 打开日志文件
func (l *Logger) openFile() {
	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
		return
	}
	l.file = f
}

// log 写入日志
func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
	}

	// 写入文件
	if l.file != nil {
		line := fmt.Sprintf("[%s] %s %s\n", entry.Time.Format("2006-01-02 15:04:05"), level, msg)
		l.file.WriteString(line)
		l.file.Sync()

		// 检查文件大小，超过 5MB 轮转
		l.checkRotation()
	}

	// 保存到内存
	l.recent = append(l.recent, entry)
	if len(l.recent) > l.maxRecent {
		l.recent = l.recent[1:]
	}
}

// Info 信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error 错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// GetRecent 获取最近 N 条日志
func (l *Logger) GetRecent(n int) []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n > len(l.recent) {
		n = len(l.recent)
	}

	result := make([]LogEntry, n)
	copy(result, l.recent[len(l.recent)-n:])
	return result
}

// checkRotation 检查日志文件大小，超过 5MB 则轮转
func (l *Logger) checkRotation() {
	if l.file == nil {
		return
	}

	info, err := l.file.Stat()
	if err != nil {
		return
	}

	if info.Size() > 5*1024*1024 { // 5MB
		l.file.Close()

		// 重命名旧文件
		oldFile := l.logFile + ".old"
		os.Remove(oldFile)
		os.Rename(l.logFile, oldFile)

		// 打开新文件
		l.openFile()
	}
}

// Close 关闭日志
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
