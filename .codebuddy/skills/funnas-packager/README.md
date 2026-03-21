# 飞牛 NAS 应用打包 Skill

## 功能说明

这个 Skill 用于将彩票助手项目打包成飞牛 NAS 应用（FunNAS App）。

## 使用场景

- 📦 打包应用到指定目录
- 🔧 生成应用清单文件
- 📝 创建应用配置
- 🎨 处理应用图标和资源
- 🚀 准备应用发布包

## 前置条件

1. 项目结构正确
2. 后端已编译完成
3. 前端已构建完成
4. 应用图标已生成
5. manifest 文件已配置

## 使用方法

### 基本打包

打包应用到输出目录：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh
```

### 打包并清理

打包应用并清理临时文件：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --clean
```

### 指定输出目录

打包应用到指定目录：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --output /path/to/output
```

### 打包并压缩

打包应用并创建压缩包：

```bash
.codebuddy/skills/funnas-packager/scripts/package.sh --compress
```

## 打包流程

1. **验证环境**
   - 检查必要文件是否存在
   - 验证目录结构
   - 检查编译产物

2. **准备应用目录**
   - 创建应用目录结构
   - 复制应用文件
   - 处理配置文件

3. **打包后端**
   - 复制编译后的二进制文件
   - 复制数据库文件（可选）
   - 配置启动脚本

4. **打包前端**
   - 复制前端构建产物
   - 处理静态资源
   - 配置路由

5. **处理图标和资源**
   - 复制应用图标
   - 处理 UI 资源
   - 生成缩略图

6. **生成清单文件**
   - 生成 app.conf 配置
   - 创建服务配置
   - 生成版本信息

7. **创建发布包**
   - 打包应用目录
   - 生成校验文件
   - 创建安装说明

## 输出结构

打包后的应用目录结构：

```
techfunway-lottery/
├── ICON.PNG                          # 应用图标
├── ICON_256.PNG                      # 应用图标大尺寸
├── app/
│   ├── bin/
│   │   └── lottery                   # 后端二进制文件
│   ├── conf/
│   │   ├── app.conf                  # 应用配置
│   │   └── service.conf              # 服务配置
│   ├── data/
│   │   └── db/
│   │       └── database.db          # 数据库文件（可选）
│   ├── ui/
│   │   ├── images/
│   │   │   ├── icon_64.png           # UI 图标
│   │   │   └── icon_256.png         # UI 图标大尺寸
│   │   └── lottery-web/             # 前端资源
│   ├── scripts/
│   │   ├── start.sh                 # 启动脚本
│   │   └── stop.sh                  # 停止脚本
│   └── manifest                    # 应用清单
├── README.md                        # 应用说明
└── VERSION.txt                      # 版本信息
```

## 配置文件说明

### manifest

应用清单文件，包含应用元信息：

```
Name: 彩票助手
Version: 1.0.0
Description: 彩票选号、购买、管理助手
Author: TechFunWay
License: MIT
...
```

### app.conf

应用配置文件：

```ini
[app]
name=lottery-assistant
version=1.0.0
port=8902

[service]
autostart=true
restart=true
```

## 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--output` | 输出目录 | `./release/funnas/` |
| `--clean` | 清理临时文件 | false |
| `--compress` | 创建压缩包 | false |
| `--version` | 指定版本号 | 从 git 获取 |
| `--with-data` | 包含数据库 | false |

## 注意事项

1. 确保所有必要文件已准备就绪
2. 检查 manifest 文件配置正确
3. 验证图标尺寸符合要求
4. 测试打包后的应用能否正常启动
5. 确认端口号未被占用

## 故障排除

### 打包失败

检查必要文件是否存在：

```bash
ls -la techfunway-lottery/
ls -la backend/lottery
ls -la frontend/dist/
```

### 图标缺失

生成应用图标：

```bash
python3 generate_icons.py
```

### 清单文件错误

检查 manifest 文件格式：

```bash
cat techfunway-lottery/manifest
```

## 相关文档

- [飞牛 NAS 应用开发文档](https://docs.funnas.com/)
- [应用打包规范](https://docs.funnas.com/app/packaging)
- [manifest 配置参考](https://docs.funnas.com/app/manifest)

## 示例

### 示例 1：基本打包

```bash
# 打包应用
.codebuddy/skills/funnas-packager/scripts/package.sh

# 输出：./release/funnas/techfunway-lottery/
```

### 示例 2：打包并压缩

```bash
# 打包并创建压缩包
.codebuddy/skills/funnas-packager/scripts/package.sh --compress

# 输出：./release/funnas/techfunway-lottery.tar.gz
```

### 示例 3：指定版本打包

```bash
# 打包指定版本
.codebuddy/skills/funnas-packager/scripts/package.sh --version v1.0.0

# 输出包含版本信息
```

### 示例 4：包含数据库打包

```bash
# 打包并包含数据库
.codebuddy/skills/funnas-packager/scripts/package.sh --with-data

# 输出包含数据库文件
```
