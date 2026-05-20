package main

import (
	"context"
	"flag"
	"fmt"
	"lottery-backend/database"
	"lottery-backend/handlers"
	"lottery-backend/logger"
	"lottery-backend/middleware"
	"lottery-backend/services"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 版本信息（编译时注入）
var (
	Version   = "v1.1.2"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// 命令行参数
var (
	dataDirFlag      = flag.String("data-dir", "./data", "Data directory path (default: ./data)")
	webDirFlag       = flag.String("web-dir", "./", "Frontend web root directory (default: ./)")
	portFlag         = flag.String("port", "", "Server port (default: 8902)")
	deviceTypeFlag   = flag.String("device-type", "", "Device type for usage stats (e.g., fnos, docker)")
	showVersion      = flag.Bool("version", false, "Show version information")
	resetAdminPasswd = flag.String("reset-admin-password", "", "Reset admin password (requires db-path or data-dir)")
)

// 配置信息
var config struct {
	dataDir string
	webRoot string
	port    string
	dbPath  string
}

func printVersion() {
	fmt.Printf("Lottery Assistant %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	if GitCommit != "unknown" {
		fmt.Printf("Git Commit: %s\n", GitCommit)
	}
}

func printHelp() {
	fmt.Println("Lottery Assistant - 彩彩助手管理系统")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./lottery [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -data-dir string            Data directory path (default: ./data)")
	fmt.Println("  -web-dir string             Frontend web root directory (default: ./)")
	fmt.Println("  -port string                Server port (default: 8902)")
	fmt.Println("  -device-type string         Device type for usage stats (e.g., fnos, docker)")
	fmt.Println("  -reset-admin-password string  Reset admin password")
	fmt.Println("  -version                    Show version information")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  PORT        Server port (overrides -port)")
	fmt.Println("  DB_PATH     Database file path (overrides default)")
	fmt.Println("  ENV         Environment: development or production")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./lottery")
	fmt.Println("  ./lottery -port 9000")
	fmt.Println("  ./lottery -web-dir /var/lottery/web")
	fmt.Println("  ./lottery -data-dir /var/lottery/data -web-dir /var/lottery/web")
	fmt.Println("  ./lottery -reset-admin-password admin123")
	fmt.Println("  PORT=8080 ./lottery")
}

func main() {
	// 设置自定义帮助函数
	flag.Usage = printHelp

	flag.Parse()

	// 显示版本信息
	if *showVersion {
		printVersion()
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

	// 初始化统一日志系统
	if err := logger.InitLogger(config.dataDir); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugarLogger := logger.GetSugarLogger()

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

	// 重置管理员密码
	if *resetAdminPasswd != "" {
		authService := services.AuthService{}
		username, err := authService.ResetAdminPassword(*resetAdminPasswd)
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n========================================\n")
		fmt.Printf("  管理员密码重置成功\n")
		fmt.Printf("========================================\n")
		fmt.Printf("  用户名: %s\n", username)
		fmt.Printf("  新密码: %s\n", *resetAdminPasswd)
		fmt.Printf("========================================\n")
		fmt.Printf("\n请使用新密码登录。\n")
		return
	}

	// 启动使用统计服务
	usageStatsSvc := services.NewUsageStatsService(Version)
	usageStatsSvc.SetAppName("lottery")
	usageStatsSvc.SetDeviceType(*deviceTypeFlag)
	usageStatsSvc.SetAPIURL("http://techfunway.wycto.cn/api/apps.online/refresh")

	// 初始化设备标识码
	if err := usageStatsSvc.InitDeviceID(config.dataDir); err != nil {
		sugarLogger.Warnf("⚠️  无法获取设备标识码: %v", err)
	} else {
		sugarLogger.Infof("📱 设备标识码: %s", usageStatsSvc.GetDeviceID())
	}

	usageStatsSvc.Start()
	defer usageStatsSvc.Stop() // 确保服务退出时停止统计服务

	// 执行数据库升级
	sugarLogger.Info("🔄 Checking database upgrades...")
	upgradeSvc := services.NewUpgradeService(database.DB, config.dataDir)
	if err := upgradeSvc.RunUpgrades(); err != nil {
		sugarLogger.Errorf("❌ Database upgrade failed: %v", err)
		os.Exit(1)
	}

	// 启动定时抓取任务
	schedulerSvc := services.NewSchedulerService()
	schedulerSvc.Start()
	handlers.SchedulerService = schedulerSvc
	defer schedulerSvc.Stop()

	// 根据环境设置 Gin 模式（默认生产模式，关闭调试日志）
	env := os.Getenv("ENV")
	if env == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化 Gin，使用 zap 日志
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger(logger.GetSugarLogger()))

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
		// 传递版本信息给处理函数
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name":      "彩彩助手",
				"version":   Version,
				"buildTime": BuildTime,
				"gitCommit": GitCommit,
				"status":    "running",
			})
		})
		api.GET("/version/current", handlers.GetCurrentVersion)
		api.GET("/version/history", handlers.GetUpgradeHistory)
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
			authorized.POST("/draws/fetch-auto", handlers.FetchAutoDraws)

			// 中奖记录
			authorized.GET("/winnings", handlers.GetWinnings)
			authorized.POST("/winnings/recheck", handlers.RecheckWinnings)

			// 统计分析
			authorized.GET("/statistics/overview", handlers.GetOverview)
			authorized.GET("/statistics/prizes", handlers.GetPrizeDistribution)
			authorized.GET("/statistics/numbers", handlers.GetNumberFrequency)
			authorized.GET("/statistics/trends", handlers.GetTrends)
			authorized.GET("/statistics/recent-winnings", handlers.GetRecentWinnings)

			// 竞彩足球
			football := authorized.Group("/football")
			{
				football.POST("/matches", handlers.CreateFootballMatch)
				football.GET("/matches", handlers.GetFootballMatches)
				football.PUT("/matches/:id", handlers.UpdateFootballMatch)
				football.DELETE("/matches/:id", handlers.DeleteFootballMatch)
				football.GET("/matches/fetch", handlers.FetchFootballMatches)
				football.POST("/matches/fetch-results", handlers.FetchFootballResults)
				football.POST("/bets", handlers.CreateFootballBet)
				football.GET("/bets", handlers.GetFootballBets)
				football.PUT("/bets/:id", handlers.UpdateFootballBet)
				football.DELETE("/bets/:id", handlers.DeleteFootballBet)
				football.POST("/bets/recheck", handlers.RecheckFootballBets)
				football.GET("/overview", handlers.GetFootballOverview)
			}
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
		sugarLogger.Warnf("⚠️  Frontend index.html not found at %s. Please place frontend build files in %s directory.", indexPath, webRoot)
	} else {
		sugarLogger.Infof("✅ Frontend files loaded from %s", webRoot)
	}

	// 根路径返回 index.html
	r.GET("/", func(c *gin.Context) {
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Frontend not available",
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

	// 图片资源路由（img 目录）
	webImgDir := filepath.Join(webRoot, "img")
	if _, err := os.Stat(webImgDir); err == nil {
		r.Static("/img", webImgDir)
		sugarLogger.Infof("✅ Images served from %s", webImgDir)
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
				"error":   "Frontend not available",
				"message": fmt.Sprintf("Please place index.html in %s directory", webRoot),
			})
		}
	})

	sugarLogger.Infof("🌐 Server starting on port %s", config.port)
	sugarLogger.Infof("📌 API endpoint: http://localhost:%s/api", config.port)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + config.port,
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugarLogger.Errorf("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	sugarLogger.Info("🛑 Shutting down server...")

	// 设置 5 秒超时以完成当前请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugarLogger.Errorf("Server forced to shutdown: %v", err)
	}

	sugarLogger.Info("✅ Server exited")
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
func ginLogger(log *logger.SugaredLogger) gin.HandlerFunc {
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

		log.Infof("[%s] %s | status: %d | IP: %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			c.ClientIP(),
		)

		// 记录错误
		for _, err := range c.Errors {
			log.Errorf("Request error: %v", err.Error())
		}

		_ = start
		_ = cost
	}
}
