# AGENTS.md — 彩彩助手 (Lottery Assistant)

## Tech Stack
- **Backend**: Go 1.25 + Gin + GORM + SQLite (pure Go via `github.com/glebarez/sqlite`, no CGo)
- **Frontend**: Vue 3 + TS + Vite + TailwindCSS + ECharts
- **Auth**: JWT (`golang-jwt`), MD5 double-encrypted password (frontend + backend)

## Dev Commands

```bash
make dev              # builds frontend + backend, starts server on :8902 (background, log: develop/lottery.log)
make release          # frontend build → cross-platform compile → fnnas .fpk packaging
make dev-frontend     # frontend build only
make dev-backend      # Go build only (output: develop/lottery)
make run              # run existing develop/lottery binary
make clean            # remove all build artifacts
make frontend-dev     # Vite dev server on :5176 (hot-reload)
make check-deps       # verify Node.js, npm, Go are installed
cd backend && go run main.go   # manual backend start
cd frontend && npm run dev     # Vite dev server on :5176
```

Env overrides: `PORT`, `DATA_DIR`, `DB_PATH`, `ENV=development` (enables Gin debug + CORS for :5173/:5176).

## Architecture

**Backend** (`backend/`): layered — `handlers → services → models/rules/database`
- Entry: `backend/main.go:25` — `Version` var (also injected via `-ldflags -X` at release)
- DB migrations: `backend/migrations/upgrade.go` — sequential Go funcs registered in `UpgradeScripts` map, run at startup
- Draw fetch: `services/scheduler.go` — polls `api.huiniao.top` periodically
- Supported: SSQ, DLT, 3D, PL3, PL5, QLC, football

**Frontend** (`frontend/`): Vue 3 SPA
- Build output: `dist/index.html` + `dist/lottery-web/` (custom `assetsDir: 'lottery-web'` in `vite.config.ts:9`)
- `make dev` copies frontend build from `frontend/dist/` into `develop/` (root-relative)
- Backend serves static files from `-web-dir` (default `./`) at `/{index.html,lottery-web/*,img/*}`
- Dev CORS origins (auto): `localhost:5173`, `localhost:5176`

## Release Pipeline

```bash
make release  # triggers in order:
              # 1) frontend build (.skill/frontend-build/scripts/build.sh)
              # 2) cross-compile for darwin/linux/windows × amd64/arm64 (.skill/cross-platform-compile/scripts/compile.sh)
              # 3) fnOS .fpk packaging (.skill/fnnas-packager/scripts/package-multiplatform.sh)
```

Individual steps can be run via `.skill/` scripts directly. Docker multi-arch build: `bash .skill/docker-builder/scripts/docker_builder.sh` (requires `make release` first).

## Notable Details

- **No CGo**: SQLite uses pure Go driver; `CGO_ENABLED=0` in cross-compile.
- **DB upgrades** are idempotent (check table/column existence before altering). See `backend/migrations/upgrade.go`.
- **First run**: no admin exists → registration flow creates the first admin.
- **Reset admin password**: `./lottery -reset-admin-password <newpass>`.
- **Usage stats**: anonymous, posts to `http://techfunway.wycto.cn/api/apps.online/refresh`, disable via `DISABLE_STATS=true`.
- **Full pack** (one-shot build all): `bash pack.sh`.
- **No test framework** detected in the repo.

## Style / Conventions

- Backend modules import prefix `lottery-backend/` (defined in `go.mod:1`).
- Frontend type defs in `frontend/src/types/index.ts`.
- API client in `frontend/src/api/index.ts`.
- New lottery type = add enum in `backend/models/`, calculator in `backend/rules/`, UI in `frontend/src/components/`.

## Git Commit

- **Language**: 全中文
- **Format**: `类型：描述` — 如 `新增：`、`修复：`、`更新：`、`修正：`、`发版：`、`同步：`
- **Workflow**: 不要直接提交推送。写好备注后先询问用户确认，确认通过后方可提交推送。
- Examples from history:
  ```
  新增：OpenCode 会话指南 AGENTS.md
  修复：双色球输入两位数时重复校验误触导致输入被清除
  发版：v1.1.1
  同步飞牛应用 v1.1.0 编译产物及更新日志
  修正：大乐透/七乐透奖金表、排列3计算逻辑；修复定时器实例双份问题
  ```
