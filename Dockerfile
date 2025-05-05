# Этап сборки
FROM golang:1.23 AS builder

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и .env
COPY . .

# Собираем бинарный файл (статически слинкованный)
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Копируем бинарный файл из билд-стадии
COPY --from=builder /app/app .

# При необходимости копируем .env (для godotenv.Load)
COPY .env .

# Копируем скрипт entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Делаем бинарный файл исполняемым
RUN chmod +x app

# Используем наш entrypoint, который устанавливает переменную SERVICE_ADDRES
ENTRYPOINT ["/entrypoint.sh"]
