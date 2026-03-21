# Frontend Build Skill 使用示例

## 示例 1: 日常开发构建

```bash
# 构建前端
.codebuddy/skills/frontend-build/scripts/build.sh

# 输出示例：
# ========================================
#   Frontend Build Script
# ========================================
# 🔨 Building frontend...
# 
# vite v5.4.21 building for production...
# ✓ 2397 modules transformed.
# ✓ built in 2.17s
# 
# ✅ Build successful!
# 📦 Output directory: /Users/weiyi/develop/gitee/TechFunWay/lottery-assistant/frontend/dist
# 📊 Build Statistics:
#    Total size: 920K
#    File count: 21
# 📁 Main files:
#    - index.html
#    - assets/ (18 JS files, 1 CSS files)
# 🎉 Build completed successfully!
```

## 示例 2: 配合 Go 后端使用

```bash
# 1. 先构建前端
.codebuddy/skills/frontend-build/scripts/build.sh

# 2. 构建跨平台 Go 程序（需要先安装 go-cross-platform-build skill）
.codebuddy/skills/go-cross-platform-build/scripts/build.sh --compress

# 3. 生成的可执行文件将包含前端资源
# 位于 release/v1.0.0/ 目录
```

## 示例 3: 直接部署到 Nginx

```bash
# 1. 构建前端
.codebuddy/skills/frontend-build/scripts/build.sh

# 2. 复制到 Nginx 静态文件目录
sudo cp -r frontend/dist/* /var/www/lottery/

# 3. 配置 Nginx 反向代理 Go 后端
# server {
#     listen 80;
#     server_name example.com;
#
#     root /var/www/lottery;
#     index index.html;
#     
#     location / {
#         try_files $uri $uri/ /index.html;
#     }
#     
#     location /api {
#         proxy_pass http://localhost:8902;
#         proxy_set_header Host $host;
#         proxy_set_header X-Real-IP $remote_addr;
#     }
# }
```

## 示例 4: 预览生产构建

```bash
# 构建后本地预览
cd frontend
npm run preview

# 访问 http://localhost:4173 查看生产版本
```

## 常见场景

### 场景 1: 开发完成后准备部署

```bash
# 1. 确保所有代码已提交
git status
git add .
git commit -m "准备发布 v1.0.0"

# 2. 构建前端
.codebuddy/skills/frontend-build/scripts/build.sh

# 3. 测试构建结果
cd frontend
npm run preview

# 4. 确认无误后，部署或打包 Go 程序
```

### 场景 2: CI/CD 自动化

```yaml
# .github/workflows/deploy.yml 示例
name: Deploy
on:
  push:
    tags:
      - 'v*'
      
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '18'
      - run: npm install
        working-directory: ./frontend
      - run: npm run build
        working-directory: ./frontend
      - uses: actions/upload-artifact@v2
        with:
          name: frontend-dist
          path: frontend/dist
```

### 场景 3: 清理并重新构建

```bash
# 如果遇到构建问题，清理后重新构建
cd frontend
rm -rf node_modules dist package-lock.json
npm install
cd ..
.codebuddy/skills/frontend-build/scripts/build.sh
```

## 性能优化建议

### 1. 启用 Gzip 压缩

构建后的文件已经经过 gzip 优化，在服务器上启用 gzip 压缩可以进一步减少传输大小：

```nginx
# Nginx 配置
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
gzip_min_length 1000;
```

### 2. 使用 CDN

将静态资源上传到 CDN 可以提升加载速度：

```bash
# 上传到 CDN
aws s3 sync frontend/dist/ s3://your-cdn-bucket/ --acl public-read
```

### 3. 缓存策略

配置浏览器缓存：

```nginx
# 静态资源缓存一年
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```
