#!/bin/bash

# Frontend Build Script for Lottery Assistant
# This script compiles the Vue 3 frontend for production deployment

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$SKILL_ROOT/../.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
DIST_DIR="$FRONTEND_DIR/dist"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Frontend Build Script${NC}"
echo -e "${BLUE}========================================${NC}"

# Check if node_modules exists
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
    echo -e "${YELLOW}⚠️  node_modules not found. Running npm install...${NC}"
    (cd "$FRONTEND_DIR" && npm install)
    echo -e "${GREEN}✅ Dependencies installed${NC}"
fi

# Check if dist directory exists and clean it
if [ -d "$DIST_DIR" ]; then
    echo -e "${YELLOW}🧹 Cleaning existing dist directory...${NC}"
    rm -rf "$DIST_DIR"
    echo -e "${GREEN}✅ Dist directory cleaned${NC}"
fi

# Build the project
echo -e "${BLUE}🔨 Building frontend...${NC}"
(cd "$FRONTEND_DIR" && npm run build)

BUILD_RESULT=$?
if [ $BUILD_RESULT -eq 0 ]; then
    echo -e "${GREEN}✅ Build successful!${NC}"
    echo -e "${GREEN}📦 Output directory: $DIST_DIR${NC}"

    # Display build statistics
    if [ -d "$DIST_DIR" ]; then
        TOTAL_SIZE=$(du -sh "$DIST_DIR" | cut -f1)
        FILE_COUNT=$(find "$DIST_DIR" -type f | wc -l)
        echo -e "${BLUE}📊 Build Statistics:${NC}"
        echo -e "   Total size: $TOTAL_SIZE"
        echo -e "   File count: $FILE_COUNT"

        # List main files
        echo -e "${BLUE}📁 Main files:${NC}"
        if [ -f "$DIST_DIR/index.html" ]; then
            echo -e "   - index.html"
        fi
        if [ -d "$DIST_DIR/assets" ]; then
            JS_COUNT=$(find "$DIST_DIR/assets" -name "*.js" | wc -l)
            CSS_COUNT=$(find "$DIST_DIR/assets" -name "*.css" | wc -l)
            echo -e "   - assets/ ($JS_COUNT JS files, $CSS_COUNT CSS files)"
        fi
    fi

    echo -e "${GREEN}🎉 Build completed successfully!${NC}"
else
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi
