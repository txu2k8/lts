package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DisplayMode control the output format
type DisplayMode int

// display modes
const (
	DisplayModeDefault DisplayMode = iota // default is the interactive output
	DisplayModePlain                      // plain text
	DisplayModeJSON                       // JSON
)

var Logger *zap.SugaredLogger

func fmtDisplayMode(fmt string) DisplayMode {
	var dp DisplayMode
	switch strings.ToLower(fmt) {
	case "json":
		dp = DisplayModeJSON
	case "plain", "text":
		dp = DisplayModePlain
	default:
		dp = DisplayModeDefault
	}
	return dp
}

// getEncoder 获取 encoder
func getEncoder(fmt string) zapcore.Encoder {
	dp := fmtDisplayMode(fmt)
	switch dp {
	case DisplayModeJSON:
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	default:
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// getLogWriter 日志写入文件，按大小切割
func getLogWriter(log_path string, print_console bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   log_path,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	// 利用io.MultiWriter支持文件和终端两个输出目标
	if print_console {
		w := io.MultiWriter(lumberJackLogger, os.Stdout)
		return zapcore.AddSync(w)
	}
	return zapcore.AddSync(lumberJackLogger)
}

// getLogWriter 日志写入文件，按日期切割
func getRotateLogWriter(log_path string, print_console bool) zapcore.WriteSyncer {
	logger, _ := rotatelogs.New(
		log_path+".%Y%m%d%H%M",
		rotatelogs.WithMaxAge(30*24*time.Hour),    // 最长保存30天
		rotatelogs.WithRotationTime(time.Hour*24), // 24小时切割一次
	)
	if print_console {
		w := io.MultiWriter(logger, os.Stdout)
		return zapcore.AddSync(w)
	}

	return zapcore.AddSync(logger)
}

// InitLogger 将err日志单独输出到文件
func InitLogger(log_path string, fmt string, level string, print_console bool) {
	// logLevel
	logLevel := zapcore.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "warn", "warning":
		logLevel = zapcore.WarnLevel
	case "err", "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// encoder
	encoder := getEncoder(fmt)

	// test.log记录全量日志
	logF := getLogWriter("./test.log", print_console)
	c1 := zapcore.NewCore(encoder, zapcore.AddSync(logF), logLevel)
	// test.err.log记录ERROR级别的日志
	errF := getLogWriter("./test.err.log", print_console)
	c2 := zapcore.NewCore(encoder, zapcore.AddSync(errF), zap.ErrorLevel)
	// 使用NewTee将c1和c2合并到core
	core := zapcore.NewTee(c1, c2)

	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}

// Debugf output the debug message to console
func Debugf(format string, args ...any) {
	Logger.Debugf(fmt.Sprintf(format, args...))
}

// Infof output the log message to console
func Infof(format string, args ...any) {
	Logger.Infof(fmt.Sprintf(format, args...))
}

// Warnf output the warning message to console
func Warnf(format string, args ...any) {
	Logger.Warnf(fmt.Sprintf(format, args...))
}

// Errorf output the error message to console
func Errorf(format string, args ...any) {
	Logger.Errorf(fmt.Sprintf(format, args...))
}
