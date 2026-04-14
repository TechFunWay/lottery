# 飞牛 NAS 应用打包 Skill

将彩彩助手项目打包成飞牛 NAS 应用（FunNAS App）。

## 功能特性

- ✅ 多平台打包（ARM/x86）
- ✅ 自动获取版本号
- ✅ 生成 .fpk 安装包
- ✅ 完整的应用清单

## 支持的平台

| 平台 | 架构 | 说明 |
|------|------|------|
| FunNAS | ARM | ARM64 架构 |
| FunNAS | x86 | AMD64 架构 |

## 前置条件

- Go 后端已编译（运行跨平台编译）
- 前端已构建
- 飞牛应用目录已配置
- fnpack 工具已安装

安装 fnpack：
```bash
# 飞牛 NAS 容器内
fnpack --version
```

## 使用方法

### 打包所有平台

```bash
.codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh
```

### 前置步骤

1. 先编译 Go 后端（ARM 和 x86）：
```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
```

2. 再编译前端：
```bash
cd frontend && npm run build
```

3. 最后打包 FunNAS 应用：
```bash
.codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh
```

## 输出目录

打包完成后，产物位于 `release/<version>/` 目录：

```
release/v1.0.0/
├── techfunway-lottery-v1.0.0-arm.fpk   # ARM 版本
└── techfunway-lottery-v1.0.0-x86.fpk   # x86 版本
```

## 飞牛应用目录结构

```
techfunway-lottery/
├── ICON.PNG                    # 应用图标
├── ICON_256.PNG                # 大尺寸图标
├── app/
│   ├── server/                 # 后端服务
│   │   └── lottery            # 可执行文件
│   └── www/                    # 前端资源
│       ├── index.html
│       └── lottery-web/
├── scripts/                    # 启动脚本
├── conf/                       # 配置目录
├── data/                       # 数据目录
├── manifest                    # 应用清单
└── README.md
```

## manifest 文件

应用清单包含以下信息：
- 应用名称
- 版本号
- 平台类型（arm/x86）
- 启动命令
- 端口配置
- 存储路径

## 部署到飞牛 NAS

1. 将生成的 `.fpk` 文件上传到飞牛 NAS
2. 在飞牛应用中心导入应用包
3. 启动应用

## 故障排查

### 未找到编译产物

确保已运行跨平台编译：
```bash
.codebuddy/skills/cross-platform-compile/scripts/compile.sh
```

### fnpack 命令不存在

在飞牛 NAS 环境中使用 fnpack 工具进行打包。
