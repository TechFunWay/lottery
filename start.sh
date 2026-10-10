#!/bin/bash

# Lottery Assistant 启动脚本
# 用法: ./start.sh [dev|release|clean|help]

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Lottery Assistant 启动脚本${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${GREEN}用法:${NC}"
    echo "  ./start.sh dev       - 开发环境：构建并启动"
    echo "  ./start.sh release   - 发布环境：构建所有平台"
    echo "  ./start.sh clean     - 清理构建产物"
    echo "  ./start.sh help      - 显示此帮助信息"
    echo ""
    echo -e "${GREEN}环境变量:${NC}"
    echo "  PORT=8902     - 设置服务端口（默认：8902）"
    echo "  DATA_DIR=./data - 设置数据目录（默认：./data）"
    echo "  WEB_DIR=./    - 设置前端目录（默认：./）"
    echo ""
    echo -e "${GREEN}快速开始:${NC}"
    echo "  1. 开发: ./start.sh dev"
    echo "  2. 发布: ./start.sh release"
    echo ""
}

# 检查 Makefile 是否存在
check_makefile() {
    if [ ! -f "Makefile" ]; then
        echo -e "${RED}❌ 错误: Makefile 不存在${NC}"
        echo -e "${YELLOW}请确保在项目根目录运行此脚本${NC}"
        exit 1
    fi
}

# 开发环境
dev_mode() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  启动开发环境${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # 检查依赖
    echo -e "${YELLOW}🔍 检查依赖...${NC}"
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${RED}❌ Node.js 未安装${NC}"
        exit 1
    fi
    
    if ! command -v npm >/dev/null 2>&1; then
        echo -e "${RED}❌ npm 未安装${NC}"
        exit 1
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        echo -e "${RED}❌ Go 未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 所有依赖检查通过${NC}"
    echo ""
    
    # 运行 make dev
    make dev
}

# 发布环境
release_mode() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  启动发布构建${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    # 检查依赖
    echo -e "${YELLOW}🔍 检查依赖...${NC}"
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${RED}❌ Node.js 未安装${NC}"
        exit 1
    fi
    
    if ! command -v npm >/dev/null 2>&1; then
        echo -e "${RED}❌ npm 未安装${NC}"
        exit 1
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        echo -e "${RED}❌ Go 未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 所有依赖检查通过${NC}"
    echo ""
    
    # 运行 make release
    make release
}

# 清理模式
clean_mode() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  清理构建产物${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    make clean
}

# 主函数
main() {
    local mode="${1:-help}"
    
    check_makefile
    
    case "$mode" in
        dev|development)
            dev_mode
            ;;
        release|build)
            release_mode
            ;;
        clean)
            clean_mode
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}❌ 未知命令: $mode${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"