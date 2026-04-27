# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

彩彩助手 (Lottery Assistant) — 彩票购买记录管理、开奖结果自动抓取、中奖自动识别及统计分析系统。

## Tech Stack

- **Backend**: Go 1.25 + Gin + GORM + SQLite (pure Go driver, no CGo)
- **Frontend**: Vue 3 + TypeScript + Vite + TailwindCSS + ECharts
- **Auth**: JWT (golang-jwt), MD5 double encryption (frontend+backend)
- **Logging**: Zap + Lumberjack (log rotation)

## Build & Dev Commands

```bash
# Dev: build frontend + backend, start server on :8902
make dev

# Release: build all platforms (darwin/linux/windows × amd64/arm64)
make release

# Cross-platform compile only (outputs to release/{version}/)
.codebuddy/skills/cross-platform-compile/scripts/compile.sh

# Build Docker multi-arch image (uses pre-compiled Linux binaries)
bash .trae/skills/docker-builder/scripts/docker_builder.sh

# Single test / run backend directly
cd backend && go run main.go
```

Common env vars: `PORT`, `DATA_DIR`, `DB_PATH`, `ENV=development` (enables Gin debug + Vite CORS origins).

## Architecture

### Backend (`backend/`)

Layered structure, dependency direction: `handlers → services → models/rules/database`.

| Layer | Path | Purpose |
|-------|------|---------|
| Entry | `main.go` | Flags, routing, server lifecycle |
| Handlers | `handlers/` | HTTP request/response, no business logic |
| Services | `services/` | Business logic, external API calls, scheduling |
| Models | `models/` | GORM data models + system config |
| Rules | `rules/` | Lottery winning calculators (per type) |
| Middleware | `middleware/` | JWT auth, admin check |
| Migrations | `migrations/` | DB schema upgrades |
| Logger | `logger/` | Structured logging wrapper |

Key services:
- `scheduler.go` — periodic lottery draw fetching from external API (`api.huiniao.top`)
- `usage_stats_service.go` — anonymous device statistics
- `upgrade_service.go` — DB migration runner
- `stats_service.go` — win/loss analysis

Supported lottery types: 双色球 (SSQ), 大乐透 (DLT), 福彩3D, 排列3, 排列5, 七乐彩 (QLC), 竞彩足球

### Frontend (`frontend/`)

Standard Vue 3 SPA with Vite. Build output: `dist/index.html` + `dist/lottery-web/` (static assets). The backend serves these as static files.

Key views: PurchaseView, DrawView, WinningsView, StatisticsView, HomeView, FootballBetView.

## Docker Build

- `Dockerfile.release` — uses pre-compiled binaries with `ARG TARGETARCH` / `ARG VERSION` (scratch-based)
- Build script creates a local multi-arch image (linux/amd64 + linux/arm64) via `docker buildx build --load`
- Image name: `techfunways/lottery`
- Health check and ca-certificates not included (scratch base)

## Version

Version string in `backend/main.go` `Version` variable. Injected at compile time via `-ldflags -X`. Front-end is decoupled from backend (no version dependency).

## Release Output Structure

```
release/{version}/lottery-assistant-{version}-{os}-{arch}/
├── lottery          # compiled binary
├── index.html       # frontend entry
├── lottery-web/     # frontend static assets
└── VERSION.txt
```
