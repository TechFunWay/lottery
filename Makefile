# Makefile for Lottery Assistant
# 
# Usage:
#   make dev      - 开发环境：打包前端，复制到Go，启动Go服务
#   make release  - 发布环境：编译所有平台，打包飞牛应用
#   make clean    - 清理构建产物
#   make help     - 显示帮助信息

.PHONY: dev dev-frontend dev-backend dev-run release clean help stop status

# 项目配置
PROJECT_NAME := lottery-assistant
DEVELOP_DIR := develop
BINARY_NAME := develop/lottery
FRONTEND_DIR := frontend
BACKEND_DIR := backend
RELEASE_DIR := release
FUNNAS_DIR := techfunway-lottery

# 颜色定义
GREEN := \033[0;32m
BLUE := \033[0;34m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

# 获取版本号
VERSION := $(shell grep -E '^\s*Version\s*=\s*"v[0-9]+\.[0-9]+\.[0-9]+"' $(BACKEND_DIR)/main.go | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "v1.0.0")

# 帮助信息
help:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  Lottery Assistant 构建工具$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@echo ""
	@echo -e "$(GREEN)可用命令:$(NC)"
	@echo "  make dev      - 开发环境：打包前端，复制到Go，启动Go服务"
	@echo "  make release  - 发布环境：编译所有平台，打包飞牛应用"
	@echo "  make clean    - 清理构建产物"
	@echo "  make help     - 显示此帮助信息"
	@echo ""
	@echo -e "$(GREEN)环境变量:$(NC)"
	@echo "  PORT=8902     - 设置服务端口（默认：8902）"
	@echo "  DATA_DIR=./data - 设置数据目录（默认：./data）"
	@echo "  WEB_DIR=./    - 设置前端目录（默认：./develop）"
	@echo ""
	@echo -e "$(GREEN)当前版本:$(NC) $(VERSION)"
	@echo ""

# 开发环境：完整流程
dev: dev-frontend dev-backend copy-frontend dev-run

# 开发环境：仅构建前端
dev-frontend:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  开发环境 - 构建前端$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo -e "$(YELLOW)⚠️  node_modules 不存在，正在安装依赖...$(NC)"; \
		cd $(FRONTEND_DIR) && npm install; \
	fi
	@echo -e "$(BLUE)🔨 构建前端...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo -e "$(GREEN)✅ 前端构建完成$(NC)"
	@echo ""

# 开发环境：复制前端文件到正确位置
copy-frontend:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  开发环境 - 复制前端文件$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ -d "$(FRONTEND_DIR)/dist" ]; then \
		echo -e "$(BLUE)📁 复制前端文件到项目根目录...$(NC)"; \
		cp -f "$(FRONTEND_DIR)/dist/index.html" "$(DEVELOP_DIR)/index.html"; \
		if [ -d "$(FRONTEND_DIR)/dist/lottery-web" ]; then \
			rm -rf "./lottery-web"; \
			cp -r "$(FRONTEND_DIR)/dist/lottery-web" "$(DEVELOP_DIR)/"; \
		fi; \
		echo -e "$(GREEN)✅ 前端文件复制完成$(NC)"; \
	else \
		echo -e "$(RED)❌ 错误: 前端构建产物不存在$(NC)"; \
		exit 1; \
	fi
	@echo ""

# 开发环境：构建后端
dev-backend:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  开发环境 - 构建后端$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)🔨 构建后端...$(NC)"
	@cd $(BACKEND_DIR) && go build -o ../$(BINARY_NAME) .
	@echo -e "$(GREEN)✅ 后端构建完成$(NC)"
	@echo ""

# 开发环境：运行服务
dev-run:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  开发环境 - 启动服务$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ ! -f "./$(BINARY_NAME)" ]; then \
		echo -e "$(YELLOW)⚠️  $(BINARY_NAME) 不存在，正在构建...$(NC)"; \
		$(MAKE) dev-backend; \
	fi
	@if [ ! -f "$(DEVELOP_DIR)/index.html" ] || [ ! -d "$(DEVELOP_DIR)/lottery-web" ]; then \
		echo -e "$(YELLOW)⚠️  前端文件不存在，正在复制...$(NC)"; \
		$(MAKE) copy-frontend; \
	fi
	@echo -e "$(GREEN)🚀 启动 Lottery Assistant...$(NC)"
	@echo -e "$(YELLOW)访问地址: http://localhost:8902$(NC)"
	@echo -e "$(YELLOW)API地址: http://localhost:8902/api$(NC)"
	@echo ""
	@EXISTING_PID=$$(lsof -ti:$${PORT:-8902} 2>/dev/null); \
	if [ -n "$$EXISTING_PID" ]; then \
		echo -e "$(YELLOW)⚠️  端口 $${PORT:-8902} 已被占用 (PID: $$EXISTING_PID)，正在停止旧进程...$(NC)"; \
		kill $$EXISTING_PID 2>/dev/null; \
		sleep 1; \
		echo -e "$(GREEN)✅ 旧进程已停止$(NC)"; \
	fi
	@echo -e "$(GREEN)🚀 启动新服务...$(NC)"
	@PORT=$${PORT:-8902} DATA_DIR=$${DATA_DIR:-$(DEVELOP_DIR)/data} ./$(BINARY_NAME) -web-dir $(DEVELOP_DIR) >> $(DEVELOP_DIR)/lottery.log 2>&1 &
	@sleep 2
	@if lsof -ti:$${PORT:-8902} > /dev/null 2>&1; then \
		echo -e "$(GREEN)✅ 服务启动成功$(NC)"; \
		echo -e "$(YELLOW)日志文件: $(DEVELOP_DIR)/lottery.log$(NC)"; \
		echo -e "$(YELLOW)按 Ctrl+C 停止服务$(NC)"; \
	else \
		echo -e "$(RED)❌ 服务启动失败，请查看日志: $(DEVELOP_DIR)/lottery.log$(NC)"; \
		exit 1; \
	fi

# 发布环境：完整流程
release:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  发布环境 - 完整构建$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(YELLOW)📦 版本: $(VERSION)$(NC)"
	@echo ""
	
	# 步骤1: 构建前端
	@echo -e "$(YELLOW)📦 步骤1: 构建前端$(NC)"
	@$(MAKE) dev-frontend 2>&1 | grep -v "node_modules"
	
	# 步骤2: 跨平台编译
	@echo -e "$(YELLOW)📦 步骤2: 跨平台编译$(NC)"
	@if [ -f ".codebuddy/skills/cross-platform-compile/scripts/compile.sh" ]; then \
		.codebuddy/skills/cross-platform-compile/scripts/compile.sh; \
	else \
		echo -e "$(RED)❌ 错误: 跨平台编译脚本不存在$(NC)"; \
		exit 1; \
	fi
	
	# 步骤3: 打包飞牛应用
	@echo -e "$(YELLOW)📦 步骤3: 打包飞牛应用$(NC)"
	@if [ -f ".codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh" ]; then \
		.codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh; \
	else \
		echo -e "$(RED)❌ 错误: 飞牛打包脚本不存在$(NC)"; \
		exit 1; \
	fi
	
	@echo ""
	@echo -e "$(GREEN)========================================$(NC)"
	@echo -e "$(GREEN)  🎉 发布构建完成！$(NC)"
	@echo -e "$(GREEN)========================================$(NC)"
	@echo -e "$(BLUE)输出目录: $(RELEASE_DIR)/$(VERSION)/$(NC)"
	@echo -e "$(BLUE)飞牛应用: $(RELEASE_DIR)/$(VERSION)/*.fpk$(NC)"
	@echo ""

# 快速发布（使用现有脚本）
release-quick:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  快速发布 - 使用现有脚本$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ -f "pack.sh" ]; then \
		./pack.sh; \
	else \
		echo -e "$(RED)❌ 错误: pack.sh 脚本不存在$(NC)"; \
		exit 1; \
	fi

# 清理构建产物
clean:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  清理构建产物$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	
	# 清理前端
	@if [ -d "$(FRONTEND_DIR)/dist" ]; then \
		echo -e "$(YELLOW)🧹 清理前端 dist 目录...$(NC)"; \
		rm -rf "$(FRONTEND_DIR)/dist"; \
		echo -e "$(GREEN)✅ 前端清理完成$(NC)"; \
	fi
	
	# 清理后端二进制
	@if [ -f "./$(BINARY_NAME)" ]; then \
		echo -e "$(YELLOW)🧹 清理后端二进制...$(NC)"; \
		rm -f "./$(BINARY_NAME)"; \
		echo -e "$(GREEN)✅ 后端清理完成$(NC)"; \
	fi
	
	# 清理复制的前端文件
	@if [ -f "./index.html" ]; then \
		echo -e "$(YELLOW)🧹 清理复制的 index.html...$(NC)"; \
		rm -f "./index.html"; \
	fi
	@if [ -d "./lottery-web" ]; then \
		echo -e "$(YELLOW)🧹 清理复制的 lottery-web 目录...$(NC)"; \
		rm -rf "./lottery-web"; \
	fi
	@echo -e "$(GREEN)✅ 前端文件清理完成$(NC)"
	
	# 清理 release 目录
	@if [ -d "$(RELEASE_DIR)" ]; then \
		echo -e "$(YELLOW)🧹 清理 release 目录...$(NC)"; \
		rm -rf "$(RELEASE_DIR)"; \
		echo -e "$(GREEN)✅ release 清理完成$(NC)"; \
	fi
	
	# 清理临时文件
	@echo -e "$(YELLOW)🧹 清理临时文件...$(NC)"
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name "*.temp" -delete 2>/dev/null || true
	@find . -name "temp-*" -type d -exec rm -rf {} + 2>/dev/null || true
	
	@echo ""
	@echo -e "$(GREEN)✅ 所有构建产物已清理$(NC)"

# 检查依赖
check-deps:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  检查依赖$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	
	@echo -e "$(YELLOW)🔍 检查 Node.js...$(NC)"
	@if command -v node >/dev/null 2>&1; then \
		echo -e "$(GREEN)✅ Node.js: $$(node --version)$(NC)"; \
	else \
		echo -e "$(RED)❌ Node.js 未安装$(NC)"; \
		exit 1; \
	fi
	
	@echo -e "$(YELLOW)🔍 检查 npm...$(NC)"
	@if command -v npm >/dev/null 2>&1; then \
		echo -e "$(GREEN)✅ npm: $$(npm --version)$(NC)"; \
	else \
		echo -e "$(RED)❌ npm 未安装$(NC)"; \
		exit 1; \
	fi
	
	@echo -e "$(YELLOW)🔍 检查 Go...$(NC)"
	@if command -v go >/dev/null 2>&1; then \
		echo -e "$(GREEN)✅ Go: $$(go version)$(NC)"; \
	else \
		echo -e "$(RED)❌ Go 未安装$(NC)"; \
		exit 1; \
	fi
	
	@echo -e "$(YELLOW)🔍 检查 fnpack...$(NC)"
	@if command -v fnpack >/dev/null 2>&1; then \
		echo -e "$(GREEN)✅ fnpack: 已安装$(NC)"; \
	else \
		echo -e "$(YELLOW)⚠️  fnpack 未安装（仅用于飞牛应用打包）$(NC)"; \
	fi
	
	@echo ""
	@echo -e "$(GREEN)✅ 所有必要依赖检查完成$(NC)"

# 显示版本信息
version:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  版本信息$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ -f "$(BACKEND_DIR)/main.go" ]; then \
		cd $(BACKEND_DIR) && go run main.go -version 2>/dev/null || echo "无法获取版本信息"; \
	else \
		echo -e "$(RED)❌ main.go 文件不存在$(NC)"; \
	fi

# 前端开发服务器（热重载）
frontend-dev:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  启动前端开发服务器$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(GREEN)🚀 启动 Vue 开发服务器...$(NC)"
	@echo -e "$(YELLOW)访问地址: http://localhost:5173$(NC)"
	@echo -e "$(YELLOW)按 Ctrl+C 停止服务器$(NC)"
	@echo ""
	@cd $(FRONTEND_DIR) && npm run dev

# 仅构建后端（用于测试）
backend-build:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  仅构建后端$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@$(MAKE) dev-backend

# 仅运行（不构建）
run:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  仅运行服务$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if [ ! -f "./$(BINARY_NAME)" ]; then \
		echo -e "$(RED)❌ 错误: $(BINARY_NAME) 不存在$(NC)"; \
		echo -e "$(YELLOW)请先运行: make dev 或 make backend-build$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(GREEN)🚀 启动 Lottery Assistant...$(NC)"
	@echo -e "$(YELLOW)访问地址: http://localhost:8902$(NC)"
	@echo -e "$(YELLOW)按 Ctrl+C 停止服务$(NC)"
	@echo ""
	@PORT=$${PORT:-8902} DATA_DIR=$${DATA_DIR:-./data} ./$(BINARY_NAME) -web-dir $(DEVELOP_DIR)

# 停止后台服务
stop:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  停止 Lottery Assistant 服务$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if pgrep -x "$(BINARY_NAME)" > /dev/null; then \
		pkill -x "$(BINARY_NAME)"; \
		echo -e "$(GREEN)✅ 服务已停止$(NC)"; \
	else \
		echo -e "$(YELLOW)⚠️  服务未运行$(NC)"; \
	fi

# 查看服务状态
status:
	@echo -e "$(BLUE)========================================$(NC)"
	@echo -e "$(BLUE)  服务状态$(NC)"
	@echo -e "$(BLUE)========================================$(NC)"
	@if pgrep -x "$(BINARY_NAME)" > /dev/null; then \
		echo -e "$(GREEN)🟢 服务正在运行$(NC)"; \
		ps aux | grep "$(BINARY_NAME)" | grep -v grep | awk '{print "   PID: " $$2, "| 内存: " $$6/1024 "MB", "| CPU: " $$3 "%"}'; \
		echo ""; \
		echo -e "$(YELLOW)访问地址: http://localhost:8902$(NC)"; \
		echo -e "$(YELLOW)API地址: http://localhost:8902/api$(NC)"; \
	else \
		echo -e "$(RED)🔴 服务未运行$(NC)"; \
		echo -e "$(YELLOW)启动服务: make dev$(NC)"; \
	fi