package services

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"orion/frontclient/utils"
	"time"
)

var (
	// Request counter
	GatewayRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_request_total",
			Help: "Total number of requests processed by the API gateway, by method and path",
		},
		[]string{"method", "path"},
	)
	// Request duration
	GatewayRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds for the API gateway, by method and path",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	UserRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_user_request_total",
			Help: "Total requests per user",
		},
		[]string{"user_id"},
	)
	UserPathCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_user_path_total",
			Help: "Total requests paths per user",
		},
		[]string{"user_id", "path"},
	)
	// Метрика для заблокированных запросов
	RateLimitBlockedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_blocked_total",
			Help: "Total requests blocked by rate limiter",
		},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(GatewayRequestCounter)
	prometheus.MustRegister(GatewayRequestDuration)
	prometheus.MustRegister(UserRequestCounter)
	prometheus.MustRegister(UserPathCounter)
	prometheus.MustRegister(RateLimitBlockedCounter)

}

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

// StartTTLCheck запускает обновление TTL-проверки каждые interval.
// Consul ожидает, что вы будете вызывать метод UpdateTTL с состоянием HealthPassing.
func RunConsulFromEnv() {

	// Подключаемся к Consul
	client := GetConsul(utils.ConsulAddress, utils.ServiceName, utils.ServiceID, utils.ServiceAddress, utils.ServicePort)

	// Запускаем горутину для обновления TTL-проверки каждые 10 секунд (меньше чем TTL)
	go StartTTLCheck(client, utils.ServiceID+"-ttl", 10*time.Second)
}
