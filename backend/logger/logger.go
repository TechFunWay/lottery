package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SugaredLogger 重新导出 zap.SugaredLogger
type SugaredLogger = zap.SugaredLogger

// Logger 全局日志实例
var (
	zapLogger    *zap.Logger
	sugarLogger *zap.SugaredLogger
	once        sync.Once
)

// 日志类型
const (
	LogTypeApp     = "app"     // 应用日志
	LogTypeError   = "error"   // 错误日志
	LogTypeUpgrade = "upgrade" // 升级日志
)

// InitLogger 初始化日志系统
func InitLogger(dataDir string) error {
	var err error
	once.Do(func() {
		err = initLoggerInternal(dataDir)
	})
	return err
}

// initLoggerInternal 内部初始化函数
func initLoggerInternal(dataDir string) error {
	// 创建必要的目录
	logsDir := filepath.Join(dataDir, "logs")
	dbDir := filepath.Join(dataDir, "db")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	// 按日期命名的日志文件（格式：YYYYMMDD）
	today := time.Now().Format("20060102")
	infoLogPath := filepath.Join(logsDir, fmt.Sprintf("%s-%s.log", LogTypeApp, today))
	errorLogPath := filepath.Join(logsDir, fmt.Sprintf("%s-%s.log", LogTypeError, today))

	// 配置普通日志轮转（保留3天）
	infoRotateHook := &lumberjack.Logger{
		Filename:   infoLogPath,
		MaxSize:    100, // MB
		MaxBackups: 3,   // 保留3天
		MaxAge:     3,   // 保留3天
		Compress:   true, // 压缩旧日志
	}

	// 配置错误日志轮转（保留30天）
	errorRotateHook := &lumberjack.Logger{
		Filename:   errorLogPath,
		MaxSize:    100, // MB
		MaxBackups: 30,  // 保留30天
		MaxAge:     30,  // 保留30天
		Compress:   true, // 压缩旧日志
	}

	// 配置 zap 编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      zapcore.OmitKey, // 生产环境不输出调用者路径
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建日志核心
	// Info 和 Debug 级别写入普通日志
	infoCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(infoRotateHook),
		zap.InfoLevel,
	)

	// Error 级别写入错误日志
	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorRotateHook),
		zap.ErrorLevel,
	)

	// 判断是否为开发环境
	isDevelopment := os.Getenv("ENV") == "development"

	if isDevelopment {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
		zapLogger = zap.New(zapcore.NewTee(infoCore, errorCore, consoleCore), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		sugarLogger = zapLogger.Sugar()
	} else {
		consoleEncoderConfig := encoderConfig
		consoleEncoderConfig.TimeKey = "" // Docker json-file driver 会加时间,这里去掉避免重复
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
		zapLogger = zap.New(zapcore.NewTee(infoCore, errorCore, consoleCore), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		sugarLogger = zapLogger.Sugar()
	}

	// 清理旧日志（非报错日志文件）
	cleanupOldLogs(logsDir)

	return nil
}

// cleanupOldLogs 清理旧日志文件（只保留3天的常规日志）
func cleanupOldLogs(logsDir string) {
	// 获取当前时间
	now := time.Now()
	// 3天前的日期
	threeDaysAgo := now.AddDate(0, 0, -3)
	threeDaysAgoStr := threeDaysAgo.Format("20060102")

	// 读取日志目录
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()

		// 只处理常规日志文件（app-YYYYMMDD.log），跳过错误日志和升级日志
		if len(fileName) >= 12 && fileName[:4] == LogTypeApp+"-" {
			// 提取日期部分（格式：app-YYYYMMDD.log -> YYYYMMDD）
			dateStr := fileName[4:12]

			// 如果文件日期超过3天，删除它
			if dateStr < threeDaysAgoStr {
				filePath := filepath.Join(logsDir, fileName)
				os.Remove(filePath)
			}
		}
	}
}

// GetLogger 获取 zap Logger 实例
func GetLogger() *zap.Logger {
	if zapLogger == nil {
		// 如果未初始化，使用默认配置
		_ = InitLogger("./data")
	}
	return zapLogger
}

// GetSugarLogger 获取 SugaredLogger 实例
func GetSugarLogger() *zap.SugaredLogger {
	if sugarLogger == nil {
		// 如果未初始化，使用默认配置
		_ = InitLogger("./data")
	}
	return sugarLogger
}

// GetUpgradeLogger 获取升级日志文件路径
func GetUpgradeLogger(dataDir string) string {
	today := time.Now().Format("20060102")
	return filepath.Join(dataDir, "logs", fmt.Sprintf("%s-%s.log", LogTypeUpgrade, today))
}

// WriteUpgradeLog 写入升级日志
func WriteUpgradeLog(dataDir, message string) error {
	upgradeLogPath := GetUpgradeLogger(dataDir)

	// 创建日志目录
	logsDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// 打开日志文件（追加模式）
	file, err := os.OpenFile(upgradeLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入日志（带时间戳）
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
	_, err = file.WriteString(logLine)

	return err
}

// Sync 刷新日志缓冲区
func Sync() {
	if zapLogger != nil {
		_ = zapLogger.Sync()
	}
}
