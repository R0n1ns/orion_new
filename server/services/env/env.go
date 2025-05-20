package env

import (
	"log"
	"os"
	"strconv"
)

var (
	Port, SecretKeyJwt, Endpoint, AccessKey, SecretKeyMinio, Bucket, DatabaseUrl string
	BlockTimeCheck, PortMinio                                                    int
	UseSSL                                                                       bool
)

func init() {
	var err error
	PortMinio, err = strconv.Atoi(os.Getenv("Minio_SERVICE_PORT"))
	if err != nil {
		log.Panic(err)
	}
	Endpoint = os.Getenv("MINIO_ENDPOINT")
	AccessKey = os.Getenv("MINIO_ACCESS_KEY")
	SecretKeyMinio = os.Getenv("MINIO_SECRET_KEY")
	UseSSL, err = strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	if err != nil {
		log.Panic(err)
	}
	Bucket = os.Getenv("MINIO_BUCKET")
	Port = os.Getenv("SERVICE_PORT")
	SecretKeyJwt = os.Getenv("JWT_SECRET")
	BlockTimeCheck, err = strconv.Atoi(os.Getenv("BlockTimeCheck"))
	if err != nil {
		log.Panic(err)
	}
	DatabaseUrl = os.Getenv("DatabaseUrl")
}
