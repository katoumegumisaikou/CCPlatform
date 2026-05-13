.PHONY: build run clean tidy test

APP_NAME := ccplatform
BUILD_DIR := build

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

dev:
	go run ./cmd/server

# Database
db-create:
	mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS ccplatform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

db-init:
	mysql -u root -p ccplatform < migrations/init.sql
