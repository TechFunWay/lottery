# Docker Builder Skill 使用示例

## 基本使用

### 1. 本地构建镜像

```bash
# 构建默认版本 v1.0.0
.codebuddy/skills/docker-builder-v2/scripts/build.sh

# 构建指定版本
.codebuddy/skills/docker-builder-v2/scripts/build.sh v1.2.0
```

### 2. 构建并推送到 Docker Hub

```bash
# 先登录 Docker Hub
docker login

# 构建并推送
.codebuddy/skills/docker-builder-v2/scripts/build.sh v1.2.0 push
```

## 推送到私有仓库

### Harbor / 私有 Registry

```bash
export DOCKER_REGISTRY=harbor.company.com/project
export DOCKER_USERNAME=myuser
export DOCKER_PASSWORD=mypass

.codebuddy/skills/docker-builder-v2/scripts/build.sh v1.2.0 push
```

### 阿里云容器镜像服务

```bash
export DOCKER_REGISTRY=registry.cn-hangzhou.aliyuncs.com
export DOCKER_USERNAME=your-username
export DOCKER_PASSWORD=your-password
export IMAGE_NAME=lottery-assistant

.codebuddy/skills/docker-builder-v2/scripts/build.sh v1.2.0 push
```

### 腾讯云容器镜像服务

```bash
export DOCKER_REGISTRY=ccr.ccs.tencentyun.com
export DOCKER_USERNAME=your-username
export DOCKER_PASSWORD=your-password
export IMAGE_NAME=lottery-assistant

.codebuddy/skills/docker-builder-v2/scripts/build.sh v1.2.0 push
```

## Docker Compose 集成

### 基础配置

```yaml
version: '3.8'

services:
  lottery:
    image: lottery-assistant:v1.2.0
    container_name: lottery-assistant
    ports:
      - "8902:8902"
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    environment:
      - ENV=production
      - PORT=8902
```

### 使用环境变量

```yaml
version: '3.8'

services:
  lottery:
    image: ${IMAGE_NAME}:${VERSION:-v1.2.0}
    container_name: ${CONTAINER_NAME:-lottery-assistant}
    ports:
      - "${HOST_PORT:-8902}:8902"
    volumes:
      - ${DATA_PATH:-./data}:/app/data
    restart: unless-stopped
    environment:
      - ENV=${ENV:-production}
      - PORT=${PORT:-8902}
      - DB_PATH=${DB_PATH:-/app/data/database.db}
```

使用方式：

```bash
# 使用默认配置
docker-compose up -d

# 使用自定义配置
ENV=production PORT=9000 docker-compose up -d
```

### 多服务编排

```yaml
version: '3.8'

services:
  lottery:
    image: lottery-assistant:v1.2.0
    container_name: lottery-assistant
    ports:
      - "8902:8902"
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    restart: unless-stopped
    environment:
      - ENV=production
      - PORT=8902
    networks:
      - lottery-net
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8902/api/version"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

  # 可选：Nginx 反向代理
  nginx:
    image: nginx:alpine
    container_name: lottery-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    depends_on:
      - lottery
    networks:
      - lottery-net
    restart: unless-stopped

networks:
  lottery-net:
    driver: bridge
```

## Kubernetes 部署

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lottery-assistant
  namespace: default
  labels:
    app: lottery-assistant
spec:
  replicas: 2
  selector:
    matchLabels:
      app: lottery-assistant
  template:
    metadata:
      labels:
        app: lottery-assistant
    spec:
      containers:
      - name: lottery-assistant
        image: lottery-assistant:v1.2.0
        ports:
        - containerPort: 8902
          name: http
        env:
        - name: ENV
          value: "production"
        - name: PORT
          value: "8902"
        volumeMounts:
        - name: data
          mountPath: /app/data
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /api/version
            port: http
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /api/version
            port: http
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: lottery-pvc
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: lottery-service
  namespace: default
spec:
  selector:
    app: lottery-assistant
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8902
  type: ClusterIP
```

### PersistentVolumeClaim

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: lottery-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

部署：

```bash
kubectl apply -f k8s/
```

## CI/CD 示例

### GitHub Actions

```yaml
name: Build and Push Docker Image

on:
  push:
    tags:
      - 'v*'
    branches:
      - main

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push
        run: |
          VERSION=${GITHUB_REF#refs/tags/}
          if [ "$VERSION" = "refs/heads/main" ]; then
            VERSION=latest
          fi
          .codebuddy/skills/docker-builder-v2/scripts/build.sh $VERSION push

      - name: Image digest
        run: echo ${{ steps.docker_build.outputs.digest }}
```

### GitLab CI

```yaml
stages:
  - build
  - deploy

variables:
  DOCKER_IMAGE: $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG

.docker-build:
  image: docker:24
  services:
    - docker:24-dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY

build-docker:
  extends: .docker-build
  stage: build
  script:
    - docker buildx create --use
    - export IMAGE_NAME=$(echo $CI_REGISTRY_IMAGE | sed 's|.*/||')
    - export DOCKER_REGISTRY=$CI_REGISTRY
    - .codebuddy/skills/docker-builder-v2/scripts/build.sh $CI_COMMIT_TAG push
  only:
    - tags

deploy-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl set image deployment/lottery-assistant lottery-assistant=$DOCKER_IMAGE
  only:
    - tags
  when: manual
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    environment {
        DOCKER_CREDENTIALS = credentials('docker-hub')
        IMAGE_NAME = 'lottery-assistant'
    }
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Build') {
            steps {
                sh 'docker buildx create --use'
                sh ".codebuddy/skills/docker-builder-v2/scripts/build.sh ${env.TAG_NAME} push"
            }
        }
        stage('Deploy') {
            steps {
                sh 'kubectl set image deployment/lottery-assistant lottery-assistant=lottery-assistant:${env.TAG_NAME}'
            }
        }
    }
}
```

## 监控和日志

### 查看日志

```bash
# 查看容器日志
docker logs -f lottery-assistant

# 查看最近 100 行日志
docker logs --tail 100 lottery-assistant

# 查看带时间戳的日志
docker logs -t lottery-assistant

# 查看容器内日志文件
docker exec lottery-assistant tail -f /app/logs/app-$(date +%Y%m%d).log
```

### 监控资源使用

```bash
# 查看容器状态
docker ps | grep lottery-assistant

# 查看资源使用
docker stats lottery-assistant

# 查看容器详细信息
docker inspect lottery-assistant
```

## 备份和恢复

### 备份数据

```bash
# 备份数据目录
docker run --rm \
  -v lottery-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/lottery-backup-$(date +%Y%m%d).tar.gz /data

# 备份数据库
docker exec lottery-assistant cp /app/data/database.db /tmp/
docker cp lottery-assistant:/tmp/database.db ./backup/database-$(date +%Y%m%d).db
```

### 恢复数据

```bash
# 恢复数据目录
docker run --rm \
  -v lottery-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/lottery-backup-20240322.tar.gz -C /

# 恢复数据库
docker cp ./backup/database-20240322.db lottery-assistant:/app/data/database.db
docker restart lottery-assistant
```

## 常见问题

### 1. 跨平台构建失败

```bash
# 启用 QEMU 模拟器
docker run --privileged --rm tonistiigi/binfmt --install all

# 创建 buildx 实例
docker buildx create --use
docker buildx inspect --bootstrap
```

### 2. 镜像太大

```bash
# 使用 distroless 基础镜像
# 修改 Dockerfile 中的基础镜像为 gcr.io/distroless/static
```

### 3. 健康检查失败

```bash
# 检查应用是否正常启动
docker logs lottery-assistant

# 手动测试健康检查
docker exec lottery-assistant wget -O- http://localhost:8902/api/version
```

### 4. 数据持久化问题

```bash
# 检查数据卷
docker volume ls

# 检查数据目录权限
docker exec lottery-assistant ls -la /app/data

# 修复权限
docker exec lottery-assistant chown -R appuser:appgroup /app/data
```
