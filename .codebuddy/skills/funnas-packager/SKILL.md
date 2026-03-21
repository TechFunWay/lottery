# 飞牛 NAS 应用打包 Skill

## 功能

将彩票助手项目打包成飞牛 NAS 应用（FunNAS App）。

## 快速开始

```bash
# 基本打包
.codebuddy/skills/funnas-packager/scripts/package.sh

# 打包并压缩
.codebuddy/skills/funnas-packager/scripts/package.sh --compress

# 指定输出目录
.codebuddy/skills/funnas-packager/scripts/package.sh --output /path/to/output
```

## 输出结构

```
techfunway-lottery/
├── ICON.PNG                    # 应用图标
├── ICON_256.PNG                # 应用图标大尺寸
├── app/
│   ├── bin/lottery            # 后端二进制
│   ├── conf/                  # 配置文件
│   ├── data/db/               # 数据库
│   ├── ui/                    # 前端资源
│   ├── scripts/               # 启动脚本
│   └── manifest              # 应用清单
└── README.md
```

## 使用场景

✅ 打包应用到指定目录
✅ 生成应用清单文件
✅ 处理应用图标和资源
✅ 准备应用发布包

## 前置条件

- 后端已编译（backend/lottery）
- 前端已构建（frontend/dist/）
- 应用图标已生成
- manifest 已配置
