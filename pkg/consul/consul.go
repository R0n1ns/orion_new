package consul

import (
	"fmt"
	"log"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// ConsulClient управляет регистрацией сервиса и TTL-проверками в Consul.
type ConsulClient struct {
	client    *consulapi.Client
	serviceID string
}

// NewConsulClient создает нового клиента Consul и регистрирует сервис.
func NewConsulClient(address, name, serviceID, servAddress string, port int) (*ConsulClient, error) {
	config := consulapi.DefaultConfig()
	if address != "" {
		config.Address = address
	} else {
		config.Address = "192.168.1.100:8500"
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Consul client: %w", err)
	}

	if port == 0 {
		port = 8000
	}
	if servAddress == "" {
		servAddress = "127.0.0.1"
	}

	svc := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: servAddress,
		Port:    port,
		Tags: []string{
			"traefik.enable=true",
			fmt.Sprintf("traefik.http.routers.%s.rule=PathPrefix(`/service`)", name),
			"traefik.http.routers." + name + ".entrypoints=web",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", name, port),
			"traefik.http.services." + name + ".loadbalancer.sticky.cookie=true",
		},
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        serviceID + "-ttl",
			TTL:                            "15s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	if err := client.Agent().ServiceRegister(svc); err != nil {
		return nil, fmt.Errorf("error registering service: %w", err)
	}

	log.Println("Service registered successfully!")
	return &ConsulClient{client: client, serviceID: serviceID}, nil
}

// StartTTLCheck обновляет TTL-проверку каждые interval.
func (c *ConsulClient) StartTTLCheck(interval time.Duration) {
	checkID := c.serviceID + "-ttl"
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.client.Agent().UpdateTTL(checkID, "Service is healthy", consulapi.HealthPassing); err != nil {
			log.Printf("Error updating TTL check: %v", err)
		}
	}
}
