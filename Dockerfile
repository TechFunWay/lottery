# 多阶段构建 - 优化镜像大小

# 阶段 1: 编译前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 阶段 2: 编译后端
FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -s -w -buildid=" \
    -o lottery .

# 阶段 3: 运行时镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

COPY --from=backend-builder --chown=appuser:appgroup /app/backend/lottery .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/index.html .
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/lottery-web ./lottery-web
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist/img ./img

RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app/data

USER appuser:appgroup

EXPOSE 8902

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8902/api/version || exit 1

ENTRYPOINT ["./lottery"]
CMD ["-data-dir", "/app/data", "-web-dir", "/app", "-device-type", "docker"]
