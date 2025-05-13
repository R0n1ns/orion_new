package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"orion/frontclient/gateway"
	"orion/frontclient/services"
	"orion/frontclient/utils"
)

func main() {
	services.RunConsulFromEnv()
	serverURL := utils.GetEnv("SERVER_URL", "http://server-app:80")

	serverProxy, _ := gateway.CreateReverseProxy(serverURL)
	wsProxy, _ := gateway.CreateReverseProxy(serverURL)

	r := mux.NewRouter()

	// Глобальный middleware
	r.Use(gateway.CombinedMiddleware)

	// Статические маршруты
	r.HandleFunc("/chat", gateway.ChatHandler)
	r.HandleFunc("/login", gateway.Loginhandler)
	r.HandleFunc("/register", gateway.Registerhandler)
	r.HandleFunc("/", gateway.ChatHandler)

	// API Gateway
	apiRouter := r.PathPrefix("/service").Subrouter()
	apiRouter.PathPrefix("/").HandlerFunc(gateway.ApiGatewayHandler(serverProxy))

	// WebSocket
	r.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsProxy.ServeHTTP(w, r)
	}))

	// Метрики
	r.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", utils.ServicePort), r))
	fmt.Println("Сервер запущен на порту:", utils.ServicePort)
}
