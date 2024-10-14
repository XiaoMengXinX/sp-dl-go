package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ANSI 颜色常量，用于终端输出
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Green   = "\033[32m"
	Magenta = "\033[35m"
)

type Level slog.Level

const (
	LevelDebug  Level = -4
	LevelInfo   Level = 0
	LevelWarn   Level = 4
	LevelError  Level = 8
	LevelSilent Level = 16
)

// TextHandler 是自定义的 Handler，支持格式化日志输出并添加颜色
type TextHandler struct {
	level slog.Level // 日志级别
}

// NewTextHandler 创建并返回一个 TextHandler
func NewTextHandler(level slog.Level) *TextHandler {
	return &TextHandler{level: level}
}

// levelColor 根据日志级别返回对应的颜色
func levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return Blue
	case slog.LevelInfo:
		return Green
	case slog.LevelWarn:
		return Yellow
	case slog.LevelError:
		return Red
	default:
		return Reset
	}
}

// Handle 处理日志记录，格式化并输出到终端
func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	// 获取文件名和行号
	_, file, line, ok := runtime.Caller(4)
	if ok {
		file = filepath.Base(file)
	} else {
		file = "unknown"
		line = 0
	}

	// 格式化时间
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// 生成彩色的日志输出
	color := levelColor(r.Level)
	logLine := fmt.Sprintf("%s%s [%s] %s (%s:%d)%s\n",
		color, timestamp, r.Level.String(), r.Message, file, line, Reset)

	// 输出日志
	fmt.Print(logLine)
	return nil
}

// Enabled 判断日志级别是否启用
func (h *TextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs 实现 Handler 接口，返回一个带有属性的 Handler
func (h *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TextHandler{level: h.level} // 此处简单返回一个新实例
}

// WithGroup 实现 Handler 接口，用于分组日志
func (h *TextHandler) WithGroup(name string) slog.Handler {
	return h // 简单实现，返回自身即可
}

// 全局 logger 实例
var logger *slog.Logger

// Init 初始化日志系统
func Init(level Level) {
	handler := NewTextHandler(slog.Level(level))
	logger = slog.New(handler)
}

// Info 打印信息日志
func Info(msg string) {
	logger.Info(msg)
}

// Infoln 打印带换行符的信息日志
func Infoln(args ...any) {
	logger.Info(fmt.Sprint(args...))
}

// Infof 格式化并打印信息日志
func Infof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Debug 打印调试日志
func Debug(msg string) {
	logger.Debug(msg)
}

// Debugln 打印带换行符的调试日志
func Debugln(args ...any) {
	logger.Debug(fmt.Sprint(args...))
}

// Debugf 格式化并打印调试日志
func Debugf(format string, args ...any) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Warn 打印警告日志
func Warn(msg string) {
	logger.Warn(msg)
}

// Warnln 打印带换行符的警告日志
func Warnln(args ...any) {
	logger.Warn(fmt.Sprint(args...))
}

// Warnf 格式化并打印警告日志
func Warnf(format string, args ...any) {
	logger.Warn(fmt.Sprintf(format, args...))
}

// Error 打印错误日志
func Error(msg string) {
	logger.Error(msg)
}

// Errorln 打印带换行符的错误日志
func Errorln(args ...any) {
	logger.Error(fmt.Sprint(args...))
}

// Errorf 格式化并打印错误日志
func Errorf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
}

// Fatal 打印致命错误日志并退出程序
func Fatal(msg string) {
	logger.Error(msg)
	os.Exit(1) // 退出程序
}

// Fatalln 打印带换行符的致命错误日志并退出程序
func Fatalln(args ...any) {
	logger.Error(fmt.Sprint(args...))
	os.Exit(1) // 退出程序
}

// Fatalf 格式化并打印致命错误日志并退出程序
func Fatalf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1) // 退出程序
}
