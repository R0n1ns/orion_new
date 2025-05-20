package env

import (
	"os"
	"strconv"
)

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present, the defaultValue is returned.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves the value of the environment variable named by the key as an integer.
// If the variable is not present or cannot be parsed as an integer, the defaultValue is returned.
func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

var (
	ConsulAddress, ServiceName, ServiceID, ServiceAddress string
	ServicePort                                           int
	RPS_public, RPS_auth                                  float64
)

func init() {
	// Get environment variables or use defaults
	ConsulAddress = GetEnv("CONSUL_ADDRESS", "consul:8500")
	ServiceName = GetEnv("SERVICE_NAME", "api-handlers")
	ServiceAddress = GetEnv("SERVICE_ADDRESS", "")
	ServicePort = GetEnvAsInt("SERVICE_PORT", 3333)
	ServiceID = GetEnv("SERVICE_ID", "api-gateway1")
	RPS_public, _ = strconv.ParseFloat(GetEnv("RPS_public", "5"), 64)
	RPS_auth, _ = strconv.ParseFloat(GetEnv("RPS_auth", "5"), 64)

}
