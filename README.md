# 彩彩助手 - 中奖记录管理系统

一个帮助您记录彩票购买、自动识别中奖情况、并提供全面统计分析的系统。

**注意**: 本应用包含匿名使用统计功能，会收集设备标识码用于统计独立设备数量。详情请查看 [PRIVACY_POLICY.md](PRIVACY_POLICY.md)。

## 功能特性

- **多用户系统**：支持用户注册、登录、权限管理
- **数据隔离**：每个用户只能访问自己的数据，保证隐私安全
- **管理员管理**：管理员可以管理系统用户，启用/禁用账号
- **MD5密码加密**：前后端双重加密，确保密码安全
- **JWT认证**：基于令牌的身份验证机制
- **支持多种彩票类型**：双色球、大乐透、福彩3D、排列3、排列5、七乐彩
- **购买记录管理**：记录购买日期、彩票类型、期号、选号、购买金额、投注方式
- **开奖结果管理**：支持手动录入和自动抓取开奖结果
- **中奖自动识别**：根据彩票规则自动匹配号码，识别中奖等级和奖金
- **统计分析**：
  - 盈亏总览（总投入 vs 总中奖金额）
  - 各奖级中奖次数分布
  - 号码热冷分析
  - 中奖率趋势图

## 技术栈

- **前端**：Vue 3 + TypeScript + Vite + TailwindCSS + ECharts
- **后端**：Go + Gin 框架 + GORM
- **数据库**：SQLite
- **日志**：Zap + Lumberjack（日志轮转）

## 项目结构

```
caipiao/
├── backend/              # Go 后端
│   ├── main.go          # 程序入口
│   ├── models/          # 数据模型
│   ├── handlers/        # API 处理器
│   ├── services/        # 业务逻辑
│   ├── rules/           # 彩票中奖规则引擎
│   └── database/        # 数据库连接
├── frontend/            # Vue 前端
│   ├── src/
│   │   ├── views/       # 页面视图
│   │   ├── components/  # 组件
│   │   ├── api/         # API 封装
│   │   └── types/       # TypeScript 类型
│   └── dist/            # 构建输出
└── README.md
```

## 快速开始

### 自动化构建系统（推荐）

项目提供了完整的自动化构建系统，支持开发环境和发布环境的构建。

#### 开发环境（一键启动）

```bash
# 方法1: 使用 Makefile（完整流程）
make dev

# 方法2: 使用启动脚本
./start.sh dev

# 方法3: 使用环境变量自定义
PORT=8080 DATA_DIR=/path/to/data make dev
```

这会自动：
1. 构建前端（Vue 3 + TypeScript）
2. 构建后端（Go）
3. 启动服务（默认端口 8902）

访问地址：`http://localhost:8902`

#### 发布环境（多平台打包）

```bash
# 方法1: 使用 Makefile
make release

# 方法2: 使用启动脚本
./start.sh release
```

这会自动：
1. 构建前端
2. 跨平台编译（macOS, Linux, Windows）
3. 打包飞牛 NAS 应用（.fpk 文件）

#### 更多命令

```bash
# 显示所有可用命令
make help
./start.sh help

# 仅构建前端
make dev-frontend

# 仅构建后端
make dev-backend

# 仅运行服务（不构建）
make run

# 清理构建产物
make clean
./start.sh clean

# 检查依赖
make check-deps

# 显示版本信息
make version
```

### 传统方式（手动启动）

#### 首次使用（系统初始化）

1. 启动后端服务
```bash
cd backend
go run main.go
```

2. 启动前端服务
```bash
cd frontend
npm install
npm run build
```

3. 访问应用
打开浏览器访问 `http://localhost:8902`

4. 注册管理员账号
首次访问时，系统会检测到没有管理员用户，显示"注册管理员账号"选项。注册第一个用户，该用户将自动成为管理员。

### 日常使用

1. **登录**：使用用户名和密码登录系统
2. **普通用户**：可以访问购买记录、开奖管理、中奖记录、统计分析等功能，所有数据仅限于当前用户
3. **管理员**：除了普通用户的所有功能，还可以访问"用户管理"菜单，管理所有用户账号

### 启动后端服务

```bash
cd backend
go run -tags=dev main.go
```

后端服务将在 `http://localhost:8902` 启动

**命令行参数：**

```bash
# 指定端口
go run -tags=dev main.go -port 9000

# 指定数据目录
go run -tags=dev main.go -data-dir /var/lottery/data

# 组合使用
go run -tags=dev main.go -port 9000 -data-dir /var/lottery/data

# 查看版本
./lottery-assistant -version

# 查看帮助
./lottery-assistant -help
```

**环境变量：**

```bash
# 使用环境变量配置（优先级高于命令行参数）
PORT=8080 DB_PATH=./custom.db go run -tags=dev main.go

# 开发环境
ENV=development go run -tags=dev main.go

# 生产环境
ENV=production go run -tags=dev main.go

# 禁用使用统计（可选）
DISABLE_STATS=true go run -tags=dev main.go
```

### 启动前端服务

```bash
cd frontend
npm install
npm run dev
```

前端服务将在 `http://localhost:5173` 启动

### 访问应用

打开浏览器访问 `http://localhost:5173`

### 用户管理

#### 首次使用
1. 访问系统后，如果检测到没有管理员，会显示"注册管理员账号"
2. 填写用户名、密码（至少6位）和可选的邮箱
3. 点击注册，该用户将成为系统的唯一管理员

#### 日常登录
1. 点击导航栏的"登录"按钮
2. 输入用户名和密码
3. 登录成功后可以访问所有功能

#### 管理员功能
1. 登录后，导航栏会显示"用户管理"菜单（仅管理员可见）
2. 可以查看所有用户列表
3. 可以启用/禁用用户账号
4. 可以删除普通用户（不能删除管理员）
5. 注意：系统始终保证至少有一个管理员

## 使用指南

### 录入购买记录

1. 点击顶部导航「购买记录」
2. 点击「新增记录」按钮
3. 选择彩票类型、填写期号、购买日期
4. 输入号码（根据彩票类型自动适配输入格式）
5. 选择投注方式、填写金额
6. 点击「保存」

### 录入开奖结果

1. 点击顶部导航「开奖管理」
2. 点击「录入开奖」按钮
3. 选择彩票类型、填写期号
4. 对于双色球和大乐透，可点击「自动抓取开奖结果」按钮从官方接口获取
5. 或手动输入开奖号码
6. 点击「保存」

### 查看统计分析

1. 点击顶部导航「统计分析」
2. 查看盈亏总览卡片
3. 浏览盈亏趋势图、奖级分布饼图、中奖率趋势图

## API 接口

### 认证接口（无需登录）
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| GET | /api/auth/check-admin | 检查管理员是否存在 |

### 用户接口（需要登录）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/auth/me | 获取当前用户信息 |

### 用户管理（需要管理员）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/users | 获取所有用户 |
| PUT | /api/users/:id | 更新用户（状态/角色） |
| DELETE | /api/users/:id | 删除用户 |

### 业务接口（需要登录）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/statistics/overview | 盈亏总览 |
| GET | /api/purchases | 购买记录列表 |
| POST | /api/purchases | 创建购买记录 |
| GET | /api/draws | 开奖结果列表 |
| POST | /api/draws | 创建开奖结果 |
| GET | /api/draws/fetch | 自动抓取开奖结果 |
| GET | /api/winnings | 中奖记录列表 |
| GET | /api/statistics/trends | 趋势数据 |

### 安全特性

- **JWT认证**：所有业务接口需要携带有效的JWT令牌
- **权限控制**：管理员接口需要管理员权限
- **数据隔离**：每个用户只能访问自己的数据
- **密码加密**：前后端MD5双重加密

## 彩票规则说明

### 双色球
- 红球：6个，范围 01-33
- 蓝球：1个，范围 01-16
- 奖级：一等奖（6红+1蓝）、二等奖（6红）、三等奖（5红+1蓝）...六等奖（1红+1蓝 或 0红+1蓝）

### 大乐透
- 前区：5个，范围 01-35
- 后区：2个，范围 01-12
- 奖级：一等奖（5前+2后）...七等奖（0前+1后 或 1前+0后）

### 福彩3D
- 3个数字，每个 0-9
- 奖级：直选（完全一致）、组选6（3个不同数字任意顺序）、组选3（2个相同+1个不同任意顺序）

### 排列3/排列5
- 3个/5个数字，每个 0-9
- 直选（完全一致）、组选

### 七乐彩
- 7个球，范围 01-30
- 另有1个特别号
- 奖级：一等奖（7个主球）...七等奖

## 数据存储

### 目录结构

```
data/
├── db/                     # 数据库文件目录
│   └── lottery-assistant.db
└── logs/                   # 日志文件目录
    ├── app.log             # 当前日志文件
    ├── app-2024-03-20.log  # 历史日志（已归档）
    └── app-2024-03-19.log  # 历史日志（已归档）
```

### 数据库

数据存储在 SQLite 数据库中，默认路径为 `data/db/lottery-assistant.db`，包含以下表：

- `purchase_records` - 购买记录
- `draw_results` - 开奖结果
- `winning_records` - 中奖记录
- `users` - 用户信息
- `system_configs` - 系统配置

### 日志系统

使用 Zap + Lumberjack 实现结构化日志和日志轮转：

**日志轮转配置：**
- 单个日志文件最大 100MB
- 保留最近 10 个备份文件
- 保留 30 天历史日志
- 自动压缩旧日志文件

**日志输出：**
- 同时输出到文件 (`data/logs/app.log`) 和控制台
- 结构化日志格式，包含时间、级别、调用者信息
- 错误级别自动记录堆栈信息

**配置方式：**

优先级：环境变量 > 命令行参数 > 默认值

| 配置项 | 命令行参数 | 环境变量 | 默认值 |
|--------|-----------|---------|--------|
| 端口 | `-port` | `PORT` | 8902 |
| 数据目录 | `-data-dir` | - | ./data |
| 数据库路径 | - | `DB_PATH` | ./data/db/database.db |
| 环境 | - | `ENV` | production |
| 使用统计 | - | `DISABLE_STATS` | false (启用) |

**示例：**
```bash
# 命令行参数
./lottery-assistant -port 9000 -data-dir /var/lottery/data

# 环境变量
PORT=8080 DB_PATH=./custom.db ./lottery-assistant

# 混合使用
PORT=8080 ./lottery-assistant -data-dir /var/lottery/data
```

## 开发说明

### 添加新的彩票类型

1. 在 `backend/models/models.go` 添加新的彩票类型枚举
2. 在 `backend/rules/calculator.go` 实现中奖计算规则
3. 在 `frontend/src/types/index.ts` 添加类型配置
4. 在 `frontend/src/components/NumberInput.vue` 添加输入适配

### 查看日志

```bash
# 实时查看日志
tail -f data/logs/app.log

# 查看最近的日志
tail -n 100 data/logs/app.log

# 搜索错误日志
grep ERROR data/logs/app.log

# 查看特定日期的日志
cat data/logs/app-2024-03-20.log
```

## 使用统计与隐私

### 统计功能说明

彩彩助手应用包含匿名使用统计功能，用于帮助开发者了解：

1. **设备统计** - 统计独立设备数量
2. **版本分布** - 了解不同版本的使用情况
3. **平台统计** - 了解不同操作系统和架构的用户分布

### 收集的数据

统计会收集以下匿名数据：

1. **设备标识码** - 基于系统信息生成的唯一标识符（MD5哈希）
2. **系统信息** - 操作系统类型、架构、主机名
3. **应用信息** - 应用版本号

**注意**：不收集任何个人身份信息、购买记录、中奖信息等敏感数据。

### 设备标识码生成原理

设备标识码基于以下系统信息生成：
- 操作系统类型和架构
- 主机名（如果可获取）
- 系统硬件信息（根据不同平台）

生成的标识码使用 MD5 哈希算法，无法反向推导原始信息。

### 如何禁用统计

```bash
# 方法1: 使用环境变量
DISABLE_STATS=true ./lottery

# 方法2: 使用 Makefile
DISABLE_STATS=true make dev

# 方法3: 使用启动脚本
DISABLE_STATS=true ./start.sh dev
```

### 查看设备标识码

应用启动时会显示设备标识码，格式为 32 位十六进制字符串：
```
📱 设备标识码: 5d41402abc4b2a76b9719d911017c592
```

设备标识码存储在本地 `./data/device_id.txt` 文件中。

详情请查看 [PRIVACY_POLICY.md](PRIVACY_POLICY.md)。

## License

MIT
