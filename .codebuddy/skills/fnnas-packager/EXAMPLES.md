# 飞牛 NAS 应用打包示例

## 示例：打包 FunNAS 应用

### 完整打包流程

```bash
# 1. 编译后端（多平台）
.codebuddy/skills/cross-platform-compile/scripts/compile.sh

# 2. 编译前端
cd frontend && npm run build && cd ..

# 3. 打包 FunNAS 应用
.codebuddy/skills/fnnas-packager/scripts/package-multiplatform.sh
```

### 输出结果

```
输出目录: release/v1.0.0/
├── techfunway-lottery-v1.0.0-arm.fpk   # ARM 版本
└── techfunway-lottery-v1.0.0-x86.fpk   # x86 版本
```

### 部署到飞牛 NAS

1. 将生成的 `.fpk` 文件上传到飞牛 NAS
2. 在飞牛应用中心导入应用包
3. 启动应用
4. 访问 http://nas-ip:8902
