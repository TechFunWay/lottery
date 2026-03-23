# Docker Builder Skill

用于构建 Docker 镜像的 Skill，支持多架构本地镜像构建。

## 功能特性

- ✅ 自动编译前端和后端
- ✅ 多架构支持 (amd64/arm64)
- ✅ 使用 Docker Buildx 构建交叉编译镜像
- ✅ 优化镜像大小 (多阶段构建)
- ✅ 健康检查配置
- ⚠️ 仅本地构建，不自动推送

## 支持的架构

| 架构 | 平台 | 说明 |
|------|------|------|
| amd64 | x86_64 | Intel/AMD 64位处理器 |
| arm64 | ARM64 | Apple Silicon, ARM服务器 |

## 前置要求

- Docker 20.10+ (支持 Buildx)
- Docker Buildx 已启用
- Go 1.25+
- Node.js 18+ (用于编译前端)

## 使用方法

### 本地构建镜像

构建默认版本 v1.0.0：

```bash
.codebuddy/skills/docker-builder/scripts/build.sh
```

### 指定版本号

```bash
.codebuddy/skills/docker-builder/scripts/build.sh v1.2.0
```

参数说明：
- 第1个参数：版本号（默认 v1.0.0）

## 配置环境变量

可通过环境变量自定义配置：

```bash
# 镜像名称
IMAGE_NAME=lottery-assistant

# 构建平台
PLATFORMS=linux/amd64,linux/arm64
```

## 镜像信息

### 镜像标签

镜像包含以下标签：
- `v{version}` - 版本号（如 v1.2.0）
- `latest` - 最新版本（推送时自动添加）

### 镜像大小

- **未压缩**: ~80MB
- **压缩后**: ~30MB

### 暴露端口

- `8902` - HTTP 服务端口

### 健康检查

镜像配置了健康检查，默认每 30 秒检查一次：

```bash
curl -f http://localhost:8902/api/version || exit 1
```

## Dockerfile 说明

使用多阶段构建优化镜像大小：

1. **前端构建阶段**: 编译 Vue 3 前端
2. **后端构建阶段**: 编译 Go 后端
3. **运行阶段**: 使用轻量级基础镜像 (alpine)

### 基础镜像

- **构建阶段**: golang:1.25-alpine, node:18-alpine
- **运行阶段**: alpine:latest

## 运行容器

### 基本运行

```bash
docker run -d \
  --name lottery-assistant \
  -p 8902:8902 \
  -v /path/to/data:/app/data \
  lottery-assistant:v1.2.0
```

### 自定义配置

```bash
docker run -d \
  --name lottery-assistant \
  -p 9000:8902 \
  -v /path/to/data:/app/data \
  -e PORT=9000 \
  -e ENV=production \
  lottery-assistant:v1.2.0
```

### 数据持久化

建议使用 Docker Volume 或宿主机目录持久化数据：

```bash
# 使用 Volume
docker run -d \
  --name lottery-assistant \
  -p 8902:8902 \
  -v lottery-data:/app/data \
  lottery-assistant:v1.2.0

# 使用宿主机目录
docker run -d \
  --name lottery-assistant \
  -p 8902:8902 \
  -v $(pwd)/data:/app/data \
  lottery-assistant:v1.2.0
```

## Docker Compose

项目根目录包含 `docker-compose.yml`：

```bash
docker-compose up -d
```

## 多架构构建

使用 Docker Buildx 构建跨平台镜像：

```bash
# 启用 Buildx
docker buildx create --use

# 构建多架构镜像
.codebuddy/skills/docker-builder/scripts/build.sh v1.2.0
```

## 手动推送镜像

构建完成后，可以手动推送到 Docker Hub 或私有仓库：

### 推送到 Docker Hub

```bash
# 登录
docker login

# 打标签
docker tag lottery-assistant:v1.2.0 yourusername/lottery-assistant:v1.2.0
docker tag lottery-assistant:v1.2.0 yourusername/lottery-assistant:latest

# 推送
docker push yourusername/lottery-assistant:v1.2.0
docker push yourusername/lottery-assistant:latest
```

### 推送到私有仓库

```bash
# 登录私有仓库
docker login myregistry.com

# 打标签
docker tag lottery-assistant:v1.2.0 myregistry.com/lottery-assistant:v1.2.0

# 推送
docker push myregistry.com/lottery-assistant:v1.2.0
```

## 故障排查

### Buildx 未启用

```bash
# 启用 Buildx
docker buildx create --use
docker buildx inspect --bootstrap
```

### 跨平台编译失败

确保使用 `docker buildx` 而不是 `docker build`：

```bash
# 错误
docker build -t lottery:v1.2.0 .

# 正确
docker buildx build --platform linux/amd64,linux/arm64 -t lottery:v1.2.0 .
```

### 权限问题

```bash
# 添加当前用户到 docker 组
sudo usermod -aG docker $USER
```

## 安全建议

1. **使用非 root 用户**: Dockerfile 已配置使用非 root 用户运行
2. **只暴露必要端口**: 仅暴露 8902 端口
3. **扫描镜像漏洞**: 推送前扫描镜像

```bash
docker scan lottery-assistant:v1.2.0
```

4. **使用 secrets 敏感信息**: 不要在 Dockerfile 中硬编码密码

## CI/CD 集成

### GitHub Actions

```yaml
name: Build and Push Docker Image

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push
        run: |
          .codebuddy/skills/docker-builder/scripts/build.sh ${{ github.ref_name }}
          docker tag lottery-assistant:${{ github.ref_name }} ${{ secrets.DOCKER_USERNAME }}/lottery-assistant:${{ github.ref_name }}
          docker tag lottery-assistant:${{ github.ref_name }} ${{ secrets.DOCKER_USERNAME }}/lottery-assistant:latest
          docker push ${{ secrets.DOCKER_USERNAME }}/lottery-assistant:${{ github.ref_name }}
          docker push ${{ secrets.DOCKER_USERNAME }}/lottery-assistant:latest
```

### GitLab CI

```yaml
docker-build:
  image: docker:24
  services:
    - docker:24-dind
  script:
    - docker buildx create --use
    - .codebuddy/skills/docker-builder/scripts/build.sh $CI_COMMIT_TAG
    - docker tag lottery-assistant:$CI_COMMIT_TAG $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG
  only:
    - tags
```

## 注意事项

1. 构建时间约 3-5 分钟（取决于网络速度）
2. 首次构建会下载基础镜像，可能较慢
3. 构建仅生成本地镜像，需要手动推送
4. 生产环境建议使用固定版本号，不要使用 latest 标签

## 项目集成

要为其他项目使用此 skill，需要：

1. 确保项目结构符合要求：
   - `backend/` - Go 后端代码
   - `frontend/` - 前端项目
   - `backend/main.go` - 入口文件

2. 修改 `scripts/build.sh` 中的配置：
   - `IMAGE_NAME` - 镜像名称
   - `PORT` - 暴露端口

3. 确保 `Dockerfile` 在项目根目录或适配项目结构
