package handlers

import (
	"fmt"
	"log"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// GerConsul создаёт клиента Consul и регистрирует сервис с TTL-проверкой.
func GerConsul(address, name, serviceID, servAddres string, port int, weight float64) *consulapi.Client {
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
		Tags: []string{
			"traefik.enable=true",
			"traefik.http.routers." + name + ".rule=PathPrefix(`/service`)",
			"traefik.http.routers." + name + ".entrypoints=web",
			fmt.Sprint("traefik.http.services.", name, ".loadbalancer.server.port=", port),
			"traefik.http.services." + name + ".loadbalancer.sticky.cookie=true",
			//fmt.Sprint("traefik.http.services.", name, ".loadbalancer.server.weight=", weight),
		},
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        serviceID + "-ttl", // Добавляем явный CheckID
			TTL:                            "15s",
			DeregisterCriticalServiceAfter: "1m",
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
