package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/handlers"
)

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
	// Создание нового маршрутизатора с использованием Gorilla Mux.
	r := mux.NewRouter()

	// Регистрация HTTP-эндпоинта для логина пользователя. Обрабатывается функцией LoginHandler.
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")

	// Регистрация WebSocket-эндпоинта для установления соединения с клиентами.
	r.HandleFunc("/ws", handlers.WSmanager.HandleWebSocket)

	// Настройка CORS middleware:
	// - AllowedOrigins: список разрешённых источников (например, для разработки с React на http://localhost:3000).
	// - AllowedMethods: разрешённые HTTP-методы.
	// - AllowedHeaders: разрешённые заголовки.
	// - AllowCredentials: разрешает передачу учетных данных (cookies, заголовки авторизации и т.д.).
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Разрешённые источники запросов
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	// Вывод в консоль информации о запуске сервера.
	fmt.Println("Сервер запущен на http://localhost:8080")

	// Запуск HTTP-сервера на порту 8080 с применением настроенного middleware.
	// Если сервер не сможет запуститься, будет выведена ошибка с помощью log.Fatal.
	log.Fatal(http.ListenAndServe(":8080", handler))
}
