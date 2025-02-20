package main

import (
	"fmt"
	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"net/http"
	"time"
)

// GerConsul создаёт клиента Consul и регистрирует сервис с TTL-проверкой.
func GetConsul(address, name, serviceID, servAddres string, port int) *consulapi.Client {
	// Создаем клиента с конфигурацией по умолчанию.
	config := consulapi.DefaultConfig()
	if address != "" {
		// Ожидаем, что address будет вида "localhost:8500"
		config.Address = address
	} else {
		config.Address = "192.168.1.100:8500"
	}
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatalf("Error creating Consul client: %v", err)
	}
	if port == 0 {
		port = 8000
	}
	if servAddres == "" {
		servAddres = "127.0.0.1"
	}

	// Регистрируем сервис с использованием TTL-проверки.
	// Вместо HTTP-проверки здесь указываем TTL и время, после которого сервис будет удален, если TTL не обновляется.
	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: servAddres,
		Port:    port,
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        serviceID + "-ttl", // Добавляем явный CheckID
			TTL:                            "15s",
			DeregisterCriticalServiceAfter: "5m",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		log.Fatalf("Error registering service: %v", err)
	}

	fmt.Println("Service registered successfully!")
	return client
}

// StartTTLCheck запускает обновление TTL-проверки каждые interval.
// Consul ожидает, что вы будете вызывать метод UpdateTTL с состоянием HealthPassing.
func StartTTLCheck(client *consulapi.Client, serviceID string, interval time.Duration) {
	checkID := serviceID //+ "-ttl" // Обычно Consul использует ID сервиса + "-ttl" как идентификатор проверки
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		// Отправляем обновление проверки: сообщение и статус passing
		err := client.Agent().UpdateTTL(checkID, "Service is healthy", consulapi.HealthPassing)
		if err != nil {
			log.Printf("Error updating TTL check: %v", err)
		} else {
			//log.Printf("Updated TTL check: %s", checkID)
		}
	}
}

// chatHandler отправляет HTML-страницу chat.html
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "frontclient/front/chat.html")
}

// loginhandler отправляет HTML-страницу chat.html
func loginhandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "frontclient/front/login.html")
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запрос: %s %s, IP: %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// main инициализирует маршруты, применяет CORS middleware и запускает HTTP-сервер.
func main() {
	// Подключаемся к Consul (Consul запущен в контейнере на порту 8500)
	client := GetConsul("localhost:8500", "go-websocket-client", "test2", "localhost", 3031)

	// Запускаем горутину для обновления TTL-проверки каждые 10 секунд (меньше чем TTL)
	go StartTTLCheck(client, "test2-ttl", 10*time.Second) //defer Consul.Agent().ServiceDeregister("test1")
	// Создание нового маршрутизатора с использованием Gorilla Mux.
	r := mux.NewRouter()
	r.HandleFunc("/chat", chatHandler)
	r.HandleFunc("/login", loginhandler)
	r.Use(loggingMiddleware)

	// Вывод в консоль информации о запуске сервера.
	fmt.Println("Сервер запущен на http://localhost:3333")

	// Запуск HTTP-сервера на порту 8080 с применением настроенного middleware.
	// Если сервер не сможет запуститься, будет выведена ошибка с помощью log.Fatal.
	log.Fatal(http.ListenAndServe(":3333", r))
}
