package env

import (
	"fmt"
	"os"
	"strconv"
)

var (
	ServerURL, DatabaseUrl string
	ServicePort            int
	RPS_public, RPS_auth   float64
	SecretKey              []byte
)

func init() {
	var err error
	ServicePort, err = strconv.Atoi(os.Getenv("SERVICE_PORT"))
	if err != nil {
		fmt.Println("SERVICE_PORT must be an integer")
		panic(err)
	}
	RPS_public, _ = strconv.ParseFloat(os.Getenv("RPS_public"), 64)
	RPS_auth, _ = strconv.ParseFloat(os.Getenv("RPS_auth"), 64)
	ServerURL = os.Getenv("SERVER_URL")
	SecretKey = []byte(os.Getenv("JWT_SECRET"))
	DatabaseUrl = os.Getenv("DATABASE_URL")
}
