# Dockerfile for Go API
FROM golang:1.23

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем все файлы в контейнер
COPY . .

# Загружаем зависимости
RUN go mod tidy

# Устанавливаем golang-migrate для миграций
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Собираем бинарный файл
RUN go build ./cmd/main.go

# Устанавливаем переменные окружения
ENV PORT=8080

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./main"]
