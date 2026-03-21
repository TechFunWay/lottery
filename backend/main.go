package main

import (
	"lottery-backend/database"
	"lottery-backend/handlers"
	"lottery-backend/middleware"
	"lottery-backend/models"
	"lottery-backend/services"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 版本信息（编译时注入）
var (
	Version   = "v1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// 全局日志器
var logger *zap.Logger
var sugarLogger *zap.SugaredLogger

// 命令行参数
var (
	dataDirFlag = flag.String("dataUrl", "./data", "Data directory path (default: ./data)")
	webDirFlag  = flag.String("webDir", "./", "Frontend web root directory containing index.html and lottery-web folder (default: ./)")
	portFlag    = flag.String("port", "", "Server port (default: 8902)")
	showVersion = flag.Bool("version", false, "Show version information")
)

// 配置信息
var config struct {
	dataDir string
	webRoot string
	port    string
	dbPath  string
}

// 初始化日志系统
func initLogger(dataDir string) {
	// 创建必要的目录
	logsDir := filepath.Join(dataDir, "logs")
	dbDir := filepath.Join(dataDir, "db")

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	// 按日期命名的日志文件
	today := time.Now().Format("2006-01-02")
	infoLogPath := filepath.Join(logsDir, "app-"+today+".log")
	errorLogPath := filepath.Join(logsDir, "error-"+today+".log")

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
		zapcore.ErrorLevel,
	)

	// 判断是否为开发环境
	isDevelopment := os.Getenv("ENV") == "development"

	if isDevelopment {
		// 开发环境：同时输出到控制台
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
		// 组合核心
		logger = zap.New(zapcore.NewTee(infoCore, errorCore, consoleCore), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		sugarLogger = logger.Sugar()
	} else {
		// 生产环境：只输出到文件
		logger = zap.New(zapcore.NewTee(infoCore, errorCore), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		sugarLogger = logger.Sugar()
	}

	// 清理旧日志（非报错日志文件）
	cleanupOldLogs(logsDir)
}

// 清理旧日志文件（只保留3天）
func cleanupOldLogs(logsDir string) {
	// 获取当前时间
	now := time.Now()
	// 3天前的日期
	threeDaysAgo := now.AddDate(0, 0, -3)

	// 读取日志目录
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 只处理 app-YYYY-MM-DD.log 格式的日志文件
		if strings.HasPrefix(file.Name(), "app-") && strings.HasSuffix(file.Name(), ".log") {
			// 提取日期
			dateStr := strings.TrimPrefix(file.Name(), "app-")
			dateStr = strings.TrimSuffix(dateStr, ".log")

			// 解析日期
			fileDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}

			// 如果文件日期超过3天，删除它
			if fileDate.Before(threeDaysAgo) {
				filePath := filepath.Join(logsDir, file.Name())
				os.Remove(filePath)
			}
		}
	}
}

func printVersion() {
	fmt.Printf("Lottery Assistant %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	if GitCommit != "unknown" {
		fmt.Printf("Git Commit: %s\n", GitCommit)
	}
}

func printHelp() {
	fmt.Println("Lottery Assistant - 彩票助手管理系统")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./lottery [options]")
	fmt.Println()
	fmt.Println("Frontend Files:")
	fmt.Println("  index.html   - Should be placed in webRoot directory (default: ./)")
	fmt.Println("  lottery-web/ - Static resources directory in webRoot")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -dataUrl string")
	fmt.Println("        Data directory path (default: ./data)")
	fmt.Println("  -webDir string")
	fmt.Println("        Frontend web root directory containing index.html and lottery-web (default: ./)")
	fmt.Println("  -port string")
	fmt.Println("        Server port (default: 8902)")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  PORT        Server port (overrides -port)")
	fmt.Println("  DB_PATH     Database file path (overrides default)")
	fmt.Println("  ENV         Environment: development or production")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./lottery")
	fmt.Println("  ./lottery -port 9000")
	fmt.Println("  ./lottery -webDir /var/lottery/web")
	fmt.Println("  ./lottery -dataUrl /var/lottery/data -webDir /var/lottery/web")
	fmt.Println("  PORT=8080 ./lottery")
}

func main() {
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		printVersion()
		return
	}

	// 检查是否需要显示帮助
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		return
	}

	// 解析配置
	config.dataDir = *dataDirFlag
	config.webRoot = *webDirFlag
	config.port = *portFlag

	// 优先级：环境变量 > 命令行参数 > 默认值
	if envPort := os.Getenv("PORT"); envPort != "" {
		config.port = envPort
	}
	if config.port == "" {
		config.port = "8902"
	}

	config.dbPath = os.Getenv("DB_PATH")
	if config.dbPath == "" {
		config.dbPath = filepath.Join(config.dataDir, "db", "database.db")
	}

	// 初始化日志
	initLogger(config.dataDir)

	// 打印版本信息
	sugarLogger.Infof("🚀 Lottery Assistant %s", Version)
	sugarLogger.Infof("📦 Build Time: %s", BuildTime)
	if GitCommit != "unknown" {
		sugarLogger.Infof("🔖 Git Commit: %s", GitCommit)
	}
	sugarLogger.Infof("📁 Data directory: %s", config.dataDir)
	sugarLogger.Infof("🌐 Web root: %s", config.webRoot)

	// 初始化数据库
	sugarLogger.Infof("📄 Database path: %s", config.dbPath)
	database.InitDB(config.dbPath)
	database.AutoMigrate(
		&models.User{},
		&models.SystemConfig{},
		&models.PurchaseRecord{},
		&models.DrawResult{},
		&models.WinningRecord{},
	)

	// 初始化系统默认配置
	cfgSvc := services.ConfigService{}
	cfgSvc.InitDefaultConfigs()

	// 初始化 Gin，使用 zap 日志
	gin.DefaultWriter = &zapWriter{sugarLogger}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger(sugarLogger))

	// CORS 配置 - 生产环境允许所有来源
	corsOrigins := []string{"*"}
	if os.Getenv("ENV") == "development" {
		corsOrigins = []string{
			"http://localhost:5173",
			"http://localhost:5176",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:5176",
		}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: corsOrigins[0] != "*",
	}))

	// API 路由
	api := r.Group("/api")
	{
		// 版本信息接口（公开）
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name":       "Lottery Assistant",
				"version":    Version,
				"buildTime":  BuildTime,
				"gitCommit":  GitCommit,
				"status":     "running",
			})
		})
		// 认证相关（不需要登录）
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/check-admin", handlers.CheckAdmin)
		}

		// 公开配置（不需要登录）
		api.GET("/configs/public", handlers.GetPublicConfigs)

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 用户信息
			authorized.GET("/auth/me", handlers.GetCurrentUser)
			authorized.PUT("/auth/password", handlers.ChangePassword)

			// 用户管理（仅管理员）
			users := authorized.Group("/users")
			users.Use(middleware.AdminMiddleware())
			{
				users.GET("", handlers.GetAllUsers)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", handlers.DeleteUser)
			}

			// 系统配置（仅管理员）
			configs := authorized.Group("/configs")
			configs.Use(middleware.AdminMiddleware())
			{
				configs.GET("", handlers.GetConfigs)
				configs.PUT("", handlers.UpdateConfigs)
			}

			// 购买记录
			authorized.POST("/purchases", handlers.CreatePurchase)
			authorized.GET("/purchases", handlers.GetPurchases)
			authorized.PUT("/purchases/:id", handlers.UpdatePurchase)
			authorized.DELETE("/purchases/:id", handlers.DeletePurchase)

			// 开奖结果
			authorized.POST("/draws", handlers.CreateDraw)
			authorized.GET("/draws", handlers.GetDraws)
			authorized.PUT("/draws/:id", handlers.UpdateDraw)
			authorized.DELETE("/draws/:id", handlers.DeleteDraw)
			authorized.GET("/draws/fetch", handlers.FetchDraw)
			authorized.POST("/draws/fetch-batch", handlers.FetchBatchDraws)

			// 中奖记录
			authorized.GET("/winnings", handlers.GetWinnings)
			authorized.POST("/winnings/recheck", handlers.RecheckWinnings)

			// 统计分析
			authorized.GET("/statistics/overview", handlers.GetOverview)
			authorized.GET("/statistics/prizes", handlers.GetPrizeDistribution)
			authorized.GET("/statistics/numbers", handlers.GetNumberFrequency)
			authorized.GET("/statistics/trends", handlers.GetTrends)
			authorized.GET("/statistics/recent-winnings", handlers.GetRecentWinnings)
		}
	}

	// 静态文件服务（前端）
	// 前端目录结构（在 webRoot 指定的目录下）：
	//   index.html - 入口文件
	//   lottery-web/ - 静态资源目录（js/css等）
	webRoot := config.webRoot
	webStaticDir := filepath.Join(webRoot, "lottery-web")

	// 检查前端文件是否存在
	indexPath := filepath.Join(webRoot, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		sugarLogger.Warnf("⚠️  Frontend index.html not found at %s. Please place the frontend build files in %s directory.", indexPath, webRoot)
	} else {
		sugarLogger.Infof("✅ Frontend files loaded from %s", webRoot)
	}

	// 根路径返回 index.html
	r.GET("/", func(c *gin.Context) {
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "Frontend not available",
				"message": fmt.Sprintf("Please place index.html in %s directory", webRoot),
			})
		}
	})

	// 静态资源路由（lottery-web 目录）
	if _, err := os.Stat(webStaticDir); err == nil {
		r.Static("/lottery-web", webStaticDir)
		sugarLogger.Infof("✅ Static resources served from %s", webStaticDir)
	} else {
		sugarLogger.Warnf("⚠️  Static resources directory not found: %s", webStaticDir)
	}

	// NoRoute 处理 SPA 路由，所有非 API 请求都返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是 API 路径，返回 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// 返回 index.html（SPA 路由）
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":  "Frontend not available",
				"message": fmt.Sprintf("Please place index.html in %s directory", webRoot),
			})
		}
	})

	sugarLogger.Infof("🌐 Server starting on port %s", config.port)
	sugarLogger.Infof("📌 API endpoint: http://localhost:%s/api", config.port)
	r.Run(":" + config.port)
}

// zapWriter 实现 io.Writer 接口，用于将 gin 日志写入 zap
type zapWriter struct {
	logger *zap.SugaredLogger
}

func (w *zapWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}

// ginLogger Gin 中间件，记录请求日志
func ginLogger(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := c.Request.Context()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		cost := ""
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			cost = c.GetString("cost")
		}

		if query != "" {
			path = path + "?" + query
		}

		logger.Infof("[%s] %s | status: %d | IP: %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			c.ClientIP(),
		)

		// 记录错误
		for _, err := range c.Errors {
			logger.Errorf("Request error: %v", err.Error())
		}

		_ = start
		_ = cost
	}
}
