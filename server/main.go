package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/data/manager"
	handlers2 "orion/server/handlers"
	_ "orion/server/services/metrics"
	"orion/server/services/ws"
	"os"
	"strconv"
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

	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	// Создание роутера
	r := mux.NewRouter()

	// Подроутер для /service
	serviceRouter := r.PathPrefix("/service").Subrouter()

	// WebSocket
	serviceRouter.HandleFunc("/ws", ws.WSmanager.HandleWebSocket)

	// Auth
	serviceRouter.HandleFunc("/api/login", handlers2.LoginHandler).Methods("POST")
	serviceRouter.HandleFunc("/api/register", handlers2.RegisterHandler).Methods("POST")

	// HTTP API для чата и профиля
	serviceRouter.HandleFunc("/api/chats", handlers2.GetChatsHandler).Methods("GET")
	serviceRouter.HandleFunc("/api/users", handlers2.GetUsersHandler).Methods("GET")
	serviceRouter.HandleFunc("/api/chat", handlers2.CreateChatHandler).Methods("POST")
	serviceRouter.HandleFunc("/api/messages", handlers2.GetChatMessagesHandler).Methods("GET")
	serviceRouter.HandleFunc("/api/messages/read", handlers2.MarkMessagesReadHandler).Methods("POST", "PUT")
	serviceRouter.HandleFunc("/api/profile", handlers2.UpdateProfileHandler).Methods("PUT")
	serviceRouter.HandleFunc("/api/profile/photo", handlers2.UploadProfilePictureHandler).Methods("POST")
	serviceRouter.HandleFunc("/api/block", handlers2.BlockUserHandler).Methods("POST")
	serviceRouter.HandleFunc("/api/unblock", handlers2.UnblockUserHandler).Methods("POST")
	serviceRouter.HandleFunc("/api/block-status", handlers2.CheckUserBlockedHandler).Methods("GET")
	serviceRouter.HandleFunc("/api/mutual-block", handlers2.CheckMutualBlockHandler).Methods("GET")
	// Метрики Prometheus
	serviceRouter.Handle("/metrics", promhttp.Handler())

	// CORS
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3333", "http://frontclient:3333"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Сервер запущен на http://localhost:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), handler))
}
