package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/data/manager"
	"orion/server/handlers/chat"
	"orion/server/handlers/login"
	"orion/server/handlers/messages"
	"orion/server/handlers/user"
	"orion/server/services/env"
	_ "orion/server/services/metrics"
	"orion/server/services/ws"
	"time"
)

func init() {
	manager.Migrate()
}

/*
Package main является точкой входа в приложение.

Основные функции:
  - Инициализация маршрутизатора (router) с использованием Gorilla Mux.
  - Регистрация HTTP-эндпоинтов:
      * "/api/login" для аутентификации пользователей (метод POST).
      * "/ws" для установления WebSocket-соединения.
  - Применение middleware для поддержки CORS, позволяющего запросы с указанных источников.
  - Запуск HTTP-сервера на порту 8080.
*/
// main инициализирует маршруты, применяет CORS middleware и запускает HTTP-сервер.
func main() {
	ctx := context.Background()
	go manager.StartUnblockWorker(ctx, time.Duration(env.BlockTimeCheck)*time.Minute) // Проверка каждые 5 минут

	// Создание роутера
	r := mux.NewRouter()

	// Подроутер для /service
	serviceRouter := r.PathPrefix("/service").Subrouter()

	// WebSocket
	serviceRouter.HandleFunc("/ws", ws.WSmanager.HandleWebSocket)

	// регистрируем роутеры
	chat.RegisterRoutes(serviceRouter)
	messages.RegisterRoutes(serviceRouter)
	login.RegisterRoutes(serviceRouter)
	user.RegisterRoutes(serviceRouter)

	// Метрики Prometheus
	serviceRouter.Handle("/metrics", promhttp.Handler())

	// CORS
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3333", "http://frontclient:3333"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Сервер запущен на http://localhost:", env.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", env.Port), handler))
}
