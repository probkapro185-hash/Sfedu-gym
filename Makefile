.PHONY: run build test migrate-up migrate-down lint docker-up docker-down

# Запуск сервера локально
run:
	go run ./cmd/server/...

# Сборка бинаря
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server

# Тесты
test:
	go test ./... -v -race -coverprofile=coverage.out

# Покрытие
cover:
	go tool cover -html=coverage.out

# Линтер (golangci-lint)
lint:
	golangci-lint run ./...

# Docker
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# Миграции (через migrate CLI)
migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

# Сгенерировать bcrypt-хеш пароля (для seeder)
hash-password:
	@read -p "Enter password: " pw; \
	htpasswd -bnBC 12 "" "$$pw" | tr -d ':\n' | sed 's/$$2y/$$2a/'
