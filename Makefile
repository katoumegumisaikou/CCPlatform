.PHONY: build run clean tidy test dev dev-all \
        fe-install fe-dev fe-build fe-lint fe-preview \
        db-create db-init

APP_NAME := ccplatform
BUILD_DIR := build
FRONTEND_DIR := frontend

# ============================================================
# 后端
# ============================================================

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	go run ./cmd/server

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

# ============================================================
# 前端
# ============================================================

fe-install:
	cd $(FRONTEND_DIR) && npm install

fe-dev:
	cd $(FRONTEND_DIR) && npm run dev

fe-build:
	cd $(FRONTEND_DIR) && npm run build

fe-lint:
	cd $(FRONTEND_DIR) && npm run lint

fe-preview:
	cd $(FRONTEND_DIR) && npm run preview

# ============================================================
# 全栈开发
# ============================================================

# dev-all 同时启动后端(8080)和前端(5173)，用于完整开发体验
dev-all:
	@echo "Starting backend on :8080..."
	go run ./cmd/server &
	@echo "Starting frontend on :5173..."
	cd $(FRONTEND_DIR) && npm run dev

# ============================================================
# 数据库
# ============================================================

db-create:
	mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS ccplatform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

db-init:
	mysql -u root -p ccplatform < migrations/init.sql
