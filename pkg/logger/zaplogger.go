package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func logFmtDisplayMode(logFmt string) DisplayMode {
	var dp DisplayMode
	switch strings.ToLower(logFmt) {
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
func getEncoder(logFmt string) zapcore.Encoder {
	dp := logFmtDisplayMode(logFmt)
	switch dp {
	case DisplayModeJSON:
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	default:
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.ConsoleSeparator = "  "
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// getlumberJackLogWriter 日志写入文件，按大小切割
func getlumberJackLogWriter(logPath string, printConsole bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	// 利用io.MultiWriter支持文件和终端两个输出目标
	if printConsole {
		w := io.MultiWriter(lumberJackLogger, os.Stdout)
		return zapcore.AddSync(w)
	}
	return zapcore.AddSync(lumberJackLogger)
}

// getRotateLogWriter 日志写入文件，按日期切割
func getRotateLogWriter(logPath string, printConsole bool) zapcore.WriteSyncer {
	logger, _ := rotatelogs.New(
		logPath, // logPath+".%Y%m%d%H%M",
		rotatelogs.WithMaxAge(30*24*time.Hour),    // 最长保存30天
		rotatelogs.WithRotationTime(time.Hour*24), // 24小时切割一次
	)
	if printConsole {
		w := io.MultiWriter(logger, os.Stdout)
		return zapcore.AddSync(w)
	}

	return zapcore.AddSync(logger)
}

// getLogWriter 日志写入文件
func getLogWriter(logPath string, printConsole bool, loggerType string) zapcore.WriteSyncer {
	switch loggerType {
	case "lumberjack":
		return getlumberJackLogWriter(logPath, printConsole)
	case "rotatelogs":
		return getRotateLogWriter(logPath, printConsole)
	default:
		return getlumberJackLogWriter(logPath, printConsole)
	}
}

func getLogDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	logDir := filepath.Join(dir, "log")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.MkdirAll(logDir, os.ModePerm)
	}
	return logDir
}

// InitLogger 将err日志单独输出到文件
func InitLogger(prefix string, logFmt string, level string, printConsole bool) {
	// log dir
	logDir := getLogDir()
	timeStr := time.Now().Format("20060102_150405")

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
	encoder := getEncoder(logFmt)

	// test.log记录全量日志
	logF := getLogWriter(filepath.Join(logDir, fmt.Sprintf("/%s_%s_all.log", timeStr, prefix)), printConsole, "rotatelogs") // lumberjack
	c1 := zapcore.NewCore(encoder, zapcore.AddSync(logF), logLevel)
	// test.err.log记录ERROR级别的日志
	errF := getLogWriter(filepath.Join(logDir, fmt.Sprintf("/%s_%s_err.log", timeStr, prefix)), printConsole, "rotatelogs") // lumberjack
	c2 := zapcore.NewCore(encoder, zapcore.AddSync(errF), zap.ErrorLevel)
	// 使用NewTee将c1和c2合并到core
	core := zapcore.NewTee(c1, c2)

	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
}
