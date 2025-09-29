# Build stage
FROM golang:1.22.2-alpine AS builder

# 安装编译依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# 复制 go mod 文件并下载依赖（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译程序
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o sublinkX .

# Final stage
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# 设置时区为 Asia/Shanghai
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1001 -S sublinkx && \
    adduser -u 1001 -S sublinkx -G sublinkx

WORKDIR /app

# 创建必要的目录并设置权限
RUN mkdir -p /app/db /app/logs /app/template && \
    chown -R sublinkx:sublinkx /app

# 复制编译好的程序
COPY --from=builder /app/sublinkX /app/sublinkX

# 切换到非 root 用户
USER sublinkx

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000 || exit 1

# 启动程序
CMD ["/app/sublinkX"]

