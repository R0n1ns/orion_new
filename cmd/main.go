package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/internal/data/managers/migrate"
	"orion/internal/services/auth"
	"orion/internal/wsService"
	"orion/pkg/consul"
	"orion/pkg/metrics"
	"os"
	"strconv"
	"time"
)

var Consul *consulapi.Client

func init() {
	migrate.Migrate()
	// Регистрируем метрику в реестре Prometheus
	prometheus.MustRegister(metrics.MessageProcessingTime)
	prometheus.MustRegister(metrics.RequestCounter)
	prometheus.MustRegister(metrics.RequestDuration)
	prometheus.MustRegister(metrics.ErrorCounter)
	prometheus.MustRegister(metrics.AppUptime)
	prometheus.MustRegister(metrics.AppInfo)
	prometheus.MustRegister(metrics.ActiveChatsGauge)

}

// loggingMiddleware логирует метод запроса, URI и IP-адрес клиента.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запрос: %s %s, IP: %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Поднимаем WebSocket соединение
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Инициализируем WebSocket менеджер
	wsManager := wsService.NewWebSocketManager()

	// Симуляция получения userID (обычно он берется из JWT или куки)
	var userID uint = 123 // Примерный ID пользователя

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Обрабатываем входящее сообщение
		wsManager.HandleMessage(message, userID)

		// Можно отправить ответ клиенту (если нужно)
		response := []byte(`{"status":"success"}`)
		err = conn.WriteMessage(websocket.TextMessage, response)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
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
	addres := os.Getenv("SERVICE_ADDRES")
	serv_name := os.Getenv("SERVICE_NAME")
	serv_id := os.Getenv("SERVICE_ID")
	//weight, _ := strconv.ParseFloat(os.Getenv("SERVICE_WEIGHT"), 64)

	// Подключаемся к Consul
	//client := consul.GerConsul("consul:8500", serv_name, serv_id, addres, port, weight)
	//go consul.StartTTLCheck(client, serv_id+"-ttl", 10*time.Second)

	client, err := consul.NewConsulClient("consul:8500", serv_name, serv_id, addres, 9000)
	if err != nil {
		log.Fatalf("Failed to initialize Consul client: %v", err)
	}

	go client.StartTTLCheck(10 * time.Second)

	// Создание роутера Gorilla Mux.
	r := mux.NewRouter()

	// Подключаем middleware для логирования запросов.
	r.Use(loggingMiddleware)

	// Создаем подроутер, убирающий префикс "/service"
	serviceRouter := r.PathPrefix("/service").Subrouter()

	// Регистрация эндпоинтов без префикса.
	serviceRouter.HandleFunc("/api/login", auth.LoginHandler).Methods("POST")
	serviceRouter.HandleFunc("/ws", wsHandler)
	serviceRouter.Handle("/metrics", promhttp.Handler())

	// Настройка CORS middleware.
	handler := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Сервер запущен на http://localhost:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), handler))
}
