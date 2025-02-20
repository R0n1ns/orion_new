package main

import (
	"fmt"
	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"log"
	"net/http"
	"orion/data"
	"orion/handlers"
	"os"
	"strconv"
	"time"
)

var Consul *consulapi.Client

func init() {
	data.Migrate()
	// Регистрируем метрику в реестре Prometheus
	prometheus.MustRegister(handlers.MessageProcessingTime)
	prometheus.MustRegister(handlers.RequestCounter)
	prometheus.MustRegister(handlers.RequestDuration)
	prometheus.MustRegister(handlers.ErrorCounter)
	prometheus.MustRegister(handlers.AppUptime)
	prometheus.MustRegister(handlers.AppInfo)
	prometheus.MustRegister(handlers.ActiveChatsGauge)

}

// loggingMiddleware логирует метод запроса, URI и IP-адрес клиента.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запрос: %s %s, IP: %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
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
	weight, _ := strconv.ParseFloat(os.Getenv("SERVICE_WEIGHT"), 64)

	// Подключаемся к Consul
	client := handlers.GerConsul("consul:8500", serv_name, serv_id, addres, port, weight)
	go handlers.StartTTLCheck(client, serv_id+"-ttl", 10*time.Second)

	// Создание роутера Gorilla Mux.
	r := mux.NewRouter()

	// Подключаем middleware для логирования запросов.
	r.Use(loggingMiddleware)

	// Создаем подроутер, убирающий префикс "/service"
	serviceRouter := r.PathPrefix("/service").Subrouter()

	// Регистрация эндпоинтов без префикса.
	serviceRouter.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")
	serviceRouter.HandleFunc("/ws", handlers.WSmanager.HandleWebSocket)
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
