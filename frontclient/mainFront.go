package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"orion/frontclient/handlers/gateway"
	"orion/frontclient/handlers/middlewares"
	"orion/frontclient/handlers/pages"
	"orion/frontclient/proxys"
	_ "orion/frontclient/services/metrics"
	"orion/frontclient/utils/env"
)

func main() {
	serverProxy, _ := proxys.CreateReverseProxy(env.ServerURL)
	wsProxy, _ := proxys.CreateReverseProxy(env.ServerURL)

	r := mux.NewRouter()

	// Глобальный middleware
	r.Use(middlewares.CombinedMiddleware)

	// подключение веб хендлеров
	pages.RegisterRoutes(r)

	// API Gateway
	apiRouter := r.PathPrefix("/service").Subrouter()
	apiRouter.PathPrefix("/").HandlerFunc(gateway.ApiGatewayHandler(serverProxy))

	// WebSocket
	r.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsProxy.ServeHTTP(w, r)
	}))

	// Метрики
	r.Handle("/metrics", promhttp.Handler())

	log.Println("Клиент запущен на порту:", env.ServicePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", env.ServicePort), r))
}
