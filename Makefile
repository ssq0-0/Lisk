BINARY_NAME=lisk

SRC_DIR=./core

ifeq ($(OS),Windows_NT)
    RM=del /F /Q
    BINARY_EXT=.exe
else
    RM=rm -f
    BINARY_EXT=
endif

all: build

build:
	@echo "==> Компиляция проекта..."
	go build -o $(BINARY_NAME)$(BINARY_EXT) $(SRC_DIR)/main.go

run: build
	@echo "==> Запуск приложения..."
	./$(BINARY_NAME)$(BINARY_EXT)

clean:
	@echo "==> Очистка..."
	$(RM) $(BINARY_NAME)$(BINARY_EXT)

deps:
	@echo "==> Установка зависимостей..."
	go mod download

update-deps:
	@echo "==> Обновление зависимостей..."
	go get -u ./...
	go mod tidy

docker-build:
	@echo "==> Сборка Docker-образа..."
	docker build -t $(BINARY_NAME) .

docker-run:
	@echo "==> Запуск Docker-контейнера..."
	docker run --rm -v "$(shell pwd)/account:/app/account" $(BINARY_NAME)

test:
	@echo "==> Запуск тестов..."
	go test ./... -v

fmt:
	@echo "==> Форматирование кода..."
	go fmt ./...

lint:
	@echo "==> Проверка кода..."
	golangci-lint run

help:
	@echo "Использование: make [команда]"
	@echo ""
	@echo "Доступные команды:"
	@echo "  build          Компилирует проект"
	@echo "  run            Компилирует и запускает проект"
	@echo "  clean          Удаляет скомпилированные файлы"
	@echo "  deps           Устанавливает зависимости"
	@echo "  update-deps    Обновляет зависимости"
	@echo "  docker-build   Собирает Docker-образ"
	@echo "  docker-run     Запускает Docker-контейнер"
	@echo "  test           Запускает тесты"
	@echo "  fmt            Форматирует код"
	@echo "  lint           Проверяет код линтером"
	@echo "  help           Показывает эту справку"

.PHONY: all build run clean deps update-deps docker-build docker-run test fmt lint help