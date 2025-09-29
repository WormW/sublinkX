# Makefile for sublinkX

# 变量定义
APP_NAME = sublinkX
VERSION = 2.1
BUILD_DIR = build
BINARY_NAME = $(APP_NAME)
GO = go
GOFLAGS = -v
LDFLAGS = -s -w -X main.version=$(VERSION)

# 目标平台
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# 颜色定义
GREEN = \033[0;32m
YELLOW = \033[0;33m
RED = \033[0;31m
NC = \033[0m # No Color

.PHONY: all build run clean test help install deps dev prod docker start stop restart status

# 默认目标
all: build

# 显示帮助信息
help:
	@echo "$(GREEN)sublinkX Makefile 使用说明$(NC)"
	@echo ""
	@echo "$(YELLOW)可用命令:$(NC)"
	@echo "  make build       - 编译当前平台的二进制文件"
	@echo "  make run         - 编译并运行程序"
	@echo "  make dev         - 开发模式运行（端口8000）"
	@echo "  make prod        - 生产模式运行（端口80）"
	@echo "  make clean       - 清理编译生成的文件"
	@echo "  make test        - 运行测试"
	@echo "  make install     - 安装依赖"
	@echo "  make cross       - 交叉编译多平台版本"
	@echo "  make docker      - 构建Docker镜像"
	@echo "  make start       - 后台启动服务"
	@echo "  make stop        - 停止服务"
	@echo "  make restart     - 重启服务"
	@echo "  make status      - 查看服务状态"
	@echo "  make version     - 显示版本信息"
	@echo "  make reset-user  - 重置管理员账号密码"
	@echo ""
	@echo "$(YELLOW)示例:$(NC)"
	@echo "  make dev         - 开发环境快速启动"
	@echo "  make prod        - 生产环境启动"
	@echo "  make build && ./$(BINARY_NAME) run -port 3000"

# 安装依赖
install: deps
deps:
	@echo "$(GREEN)正在安装依赖...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)依赖安装完成！$(NC)"

# 编译
build:
	@echo "$(GREEN)正在编译 $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)编译完成: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# 快速编译（用于开发）
quick:
	@echo "$(GREEN)快速编译...$(NC)"
	$(GO) build -o $(BINARY_NAME) .
	@echo "$(GREEN)编译完成: $(BINARY_NAME)$(NC)"

# 运行程序
run: build
	@echo "$(GREEN)启动 $(APP_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式（端口8000，启用gin debug模式）
dev:
	@echo "$(GREEN)开发模式启动 (端口: 8000)...$(NC)"
	@export GIN_MODE=debug && $(GO) run . run -port 8000

# 生产模式（端口80，禁用gin debug模式）
prod:
	@echo "$(GREEN)生产模式启动 (端口: 80)...$(NC)"
	@export GIN_MODE=release && sudo $(GO) run . run -port 80

# 自定义端口运行
port:
	@read -p "请输入端口号: " port; \
	echo "$(GREEN)启动服务 (端口: $$port)...$(NC)"; \
	$(GO) run . run -port $$port

# 后台启动服务
start: build
	@echo "$(GREEN)后台启动服务...$(NC)"
	@if [ -f $(APP_NAME).pid ]; then \
		echo "$(RED)服务已在运行中，PID: $$(cat $(APP_NAME).pid)$(NC)"; \
		exit 1; \
	fi
	@nohup ./$(BUILD_DIR)/$(BINARY_NAME) > $(APP_NAME).log 2>&1 & echo $$! > $(APP_NAME).pid
	@echo "$(GREEN)服务已启动，PID: $$(cat $(APP_NAME).pid)$(NC)"
	@echo "$(YELLOW)日志文件: $(APP_NAME).log$(NC)"

# 停止服务
stop:
	@echo "$(YELLOW)停止服务...$(NC)"
	@if [ -f $(APP_NAME).pid ]; then \
		kill -9 $$(cat $(APP_NAME).pid) 2>/dev/null || true; \
		rm -f $(APP_NAME).pid; \
		echo "$(GREEN)服务已停止$(NC)"; \
	else \
		echo "$(RED)服务未运行$(NC)"; \
	fi

# 重启服务
restart: stop start

# 查看服务状态
status:
	@if [ -f $(APP_NAME).pid ]; then \
		if ps -p $$(cat $(APP_NAME).pid) > /dev/null; then \
			echo "$(GREEN)服务运行中，PID: $$(cat $(APP_NAME).pid)$(NC)"; \
		else \
			echo "$(YELLOW)PID文件存在但进程未运行$(NC)"; \
			rm -f $(APP_NAME).pid; \
		fi \
	else \
		echo "$(RED)服务未运行$(NC)"; \
	fi

# 查看日志
logs:
	@if [ -f $(APP_NAME).log ]; then \
		tail -f $(APP_NAME).log; \
	else \
		echo "$(RED)日志文件不存在$(NC)"; \
	fi

# 清理
clean:
	@echo "$(YELLOW)清理编译文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f $(APP_NAME).pid
	@rm -f $(APP_NAME).log
	@echo "$(GREEN)清理完成！$(NC)"

# 运行测试
test:
	@echo "$(GREEN)运行测试...$(NC)"
	$(GO) test -v ./...

# 测试覆盖率
test-coverage:
	@echo "$(GREEN)生成测试覆盖率报告...$(NC)"
	$(GO) test -v -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告已生成: coverage.html$(NC)"

# 代码格式化
fmt:
	@echo "$(GREEN)格式化代码...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)格式化完成！$(NC)"

# 代码检查
lint:
	@echo "$(GREEN)运行代码检查...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)请先安装 golangci-lint$(NC)"; \
		echo "$(YELLOW)安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

# 显示版本
version:
	@echo "$(GREEN)$(APP_NAME) 版本: $(VERSION)$(NC)"
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		./$(BUILD_DIR)/$(BINARY_NAME) -version; \
	else \
		echo "$(YELLOW)请先运行 make build 编译程序$(NC)"; \
	fi

# 重置管理员账号密码
reset-user:
	@read -p "请输入新的管理员用户名: " username; \
	read -sp "请输入新的管理员密码: " password; \
	echo ""; \
	$(GO) run . setting -username $$username -password $$password; \
	echo "$(GREEN)管理员账号已重置$(NC)"

# 交叉编译
cross:
	@echo "$(GREEN)开始交叉编译...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} .; \
		echo "$(GREEN)编译完成: $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(NC)"; \
	done

# 构建最小化发布版本
release: clean
	@echo "$(GREEN)构建发布版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	CGO_ENABLED=0 $(GO) build -trimpath $(GOFLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/release/$(BINARY_NAME) .
	@cp -r template $(BUILD_DIR)/release/ 2>/dev/null || true
	@echo "$(GREEN)发布版本构建完成: $(BUILD_DIR)/release/$(NC)"

# Docker相关
docker-build:
	@echo "$(GREEN)构建Docker镜像...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(GREEN)Docker镜像构建完成$(NC)"

docker-run:
	@echo "$(GREEN)运行Docker容器...$(NC)"
	docker run -d --name $(APP_NAME) \
		-p 8000:8000 \
		-v $$(pwd)/db:/app/db \
		-v $$(pwd)/template:/app/template \
		--restart unless-stopped \
		$(APP_NAME):latest

docker-stop:
	@echo "$(YELLOW)停止Docker容器...$(NC)"
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)

docker-logs:
	docker logs -f $(APP_NAME)

# 数据库备份
backup:
	@echo "$(GREEN)备份数据库...$(NC)"
	@mkdir -p backups
	@cp -r db backups/db_$$(date +%Y%m%d_%H%M%S)
	@echo "$(GREEN)数据库备份完成$(NC)"

# 恢复数据库
restore:
	@echo "$(YELLOW)可用的备份:$(NC)"
	@ls -la backups/
	@read -p "请输入要恢复的备份目录名称: " backup; \
	if [ -d "backups/$$backup" ]; then \
		cp -r backups/$$backup/* db/; \
		echo "$(GREEN)数据库恢复完成$(NC)"; \
	else \
		echo "$(RED)备份不存在$(NC)"; \
	fi

# 监控模式（自动重启）
watch:
	@echo "$(GREEN)监控模式启动（文件变化时自动重启）...$(NC)"
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)请先安装 air$(NC)"; \
		echo "$(YELLOW)安装命令: go install github.com/cosmtrek/air@latest$(NC)"; \
	fi

# 性能分析
bench:
	@echo "$(GREEN)运行性能测试...$(NC)"
	$(GO) test -bench=. -benchmem ./...

# 安装为系统服务（Linux systemd）
install-service:
	@echo "$(GREEN)安装为系统服务...$(NC)"
	@sudo cp scripts/sublinkx.service /etc/systemd/system/ 2>/dev/null || \
		(echo "[Unit]\nDescription=sublinkX Service\nAfter=network.target\n\n[Service]\nType=simple\nUser=$$(whoami)\nWorkingDirectory=$$(pwd)\nExecStart=$$(pwd)/$(BUILD_DIR)/$(BINARY_NAME)\nRestart=always\n\n[Install]\nWantedBy=multi-user.target" | sudo tee /etc/systemd/system/sublinkx.service)
	@sudo systemctl daemon-reload
	@sudo systemctl enable sublinkx
	@echo "$(GREEN)系统服务安装完成$(NC)"
	@echo "$(YELLOW)使用以下命令管理服务:$(NC)"
	@echo "  sudo systemctl start sublinkx   - 启动服务"
	@echo "  sudo systemctl stop sublinkx    - 停止服务"
	@echo "  sudo systemctl restart sublinkx - 重启服务"
	@echo "  sudo systemctl status sublinkx  - 查看状态"

# 卸载系统服务
uninstall-service:
	@echo "$(YELLOW)卸载系统服务...$(NC)"
	@sudo systemctl stop sublinkx 2>/dev/null || true
	@sudo systemctl disable sublinkx 2>/dev/null || true
	@sudo rm -f /etc/systemd/system/sublinkx.service
	@sudo systemctl daemon-reload
	@echo "$(GREEN)系统服务已卸载$(NC)"