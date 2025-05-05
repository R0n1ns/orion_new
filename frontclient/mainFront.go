package main

import (
	"fmt"
	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Metrics for API Gateway
var (
	// Request counter
	gatewayRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_request_total",
			Help: "Total number of requests processed by the API gateway, by method and path",
		},
		[]string{"method", "path"},
	)

	// Request duration
	gatewayRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds for the API gateway, by method and path",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Error counter
	gatewayErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_error_total",
			Help: "Total number of errors in the API gateway, by method and path",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(gatewayRequestCounter)
	prometheus.MustRegister(gatewayRequestDuration)
	prometheus.MustRegister(gatewayErrorCounter)
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

// chatHandler отправляет HTML-страницу chat.html
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "front/chat.html")
}

// loginhandler отправляет HTML-страницу chat.html
func loginhandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "front/login.html")
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Запрос: %s %s, IP: %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// registerhandler отправляет HTML-страницу chat.html
func registerhandler(w http.ResponseWriter, r *http.Request) {
	// Путь к файлу chat.html, при необходимости измените его
	http.ServeFile(w, r, "front/register.html")
}

// createReverseProxy creates a reverse proxy to the server
func createReverseProxy(targetHost string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director function to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	// Add metrics collection to the proxy
	proxy.ModifyResponse = func(resp *http.Response) error {
		status := resp.StatusCode
		if status >= 400 {
			gatewayErrorCounter.WithLabelValues(
				resp.Request.Method,
				resp.Request.URL.Path,
				fmt.Sprintf("%d", status),
			).Inc()
		}
		return nil
	}

	return proxy, nil
}

// apiGatewayHandler handles API requests and forwards them to the server
func apiGatewayHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Record start time
		start := time.Now()

		// Increment request counter
		gatewayRequestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

		// Create a response recorder to capture the response
		responseRecorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Forward the request to the server
		proxy.ServeHTTP(responseRecorder, r)

		// Record request duration
		duration := time.Since(start).Seconds()
		gatewayRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		// Log the request
		log.Printf("[API Gateway] %s %s -> %d (%.2fs)", r.Method, r.URL.Path, responseRecorder.statusCode, duration)
	}
}

// responseRecorder is a wrapper for http.ResponseWriter that records the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, the defaultValue is returned.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves the value of the environment variable named by the key as an integer.
// If the variable is not present or cannot be parsed as an integer, the defaultValue is returned.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// main инициализирует маршруты, применяет CORS middleware и запускает HTTP-сервер.
func main() {
	// Get environment variables or use defaults
	consulAddress := getEnv("CONSUL_ADDRESS", "consul:8500")
	serviceName := getEnv("SERVICE_NAME", "api-gateway")
	serviceID := getEnv("SERVICE_ID", "api-gateway1")
	serviceAddress := getEnv("SERVICE_ADDRESS", "")
	servicePort := getEnvAsInt("SERVICE_PORT", 3333)
	serverURL := getEnv("SERVER_URL", "http://server-app:80")

	// Подключаемся к Consul
	client := GetConsul(consulAddress, serviceName, serviceID, serviceAddress, servicePort)

	// Запускаем горутину для обновления TTL-проверки каждые 10 секунд (меньше чем TTL)
	go StartTTLCheck(client, serviceID+"-ttl", 10*time.Second)

	// Create reverse proxy to the server
	serverProxy, err := createReverseProxy(serverURL)
	if err != nil {
		log.Fatalf("Failed to create reverse proxy: %v", err)
	}

	// Создание нового маршрутизатора с использованием Gorilla Mux.
	r := mux.NewRouter()

	// Static content handlers
	r.HandleFunc("/chat", chatHandler)
	r.HandleFunc("/login", loginhandler)
	r.HandleFunc("/register", registerhandler)
	r.HandleFunc("/", chatHandler)

	// API Gateway - forward requests to the server
	apiRouter := r.PathPrefix("/service").Subrouter()
	apiRouter.PathPrefix("/").HandlerFunc(apiGatewayHandler(serverProxy))

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	r.Use(loggingMiddleware)

	// Вывод в консоль информации о запуске сервера.
	fmt.Printf("API Gateway запущен на http://localhost:%d\n", servicePort)

	// Запуск HTTP-сервера на порту из переменной окружения с применением настроенного middleware.
	// Если сервер не сможет запуститься, будет выведена ошибка с помощью log.Fatal.
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", servicePort), r))
}
