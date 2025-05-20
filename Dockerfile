# Этап сборки
FROM golang:1.23 AS builder

WORKDIR /app

# Копируем go.mod и go.sum, скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Переходим в папку server и собираем приложение
WORKDIR /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app .

# Финальный минимальный образ
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из билдера
COPY --from=builder /app/app .

## Копируем .env при необходимости (если он используется внутри приложения)
#COPY .env .

# Делаем бинарный файл исполняемым
RUN chmod +x app

# Указываем команду по умолчанию
ENTRYPOINT ["./app"]
