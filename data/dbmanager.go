package data

//
//import (
//	"context"
//	"fmt"
//	"github.com/redis/go-redis/v9"
//	"time"
//)
//
//// storage/redis.go
//type Config struct {
//	Addr        string        `yaml:"addr"`
//	Password    string        `yaml:"password"`
//	User        string        `yaml:"user"`
//	DB          int           `yaml:"db"`
//	MaxRetries  int           `yaml:"max_retries"`
//	DialTimeout time.Duration `yaml:"dial_timeout"`
//	Timeout     time.Duration `yaml:"timeout"`
//}
//
//// storage/redis.go
//func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
//	db := redis.NewClient(&redis.Options{
//		Addr:         cfg.Addr,
//		Password:     cfg.Password,
//		DB:           cfg.DB,
//		Username:     cfg.User,
//		MaxRetries:   cfg.MaxRetries,
//		DialTimeout:  cfg.DialTimeout,
//		ReadTimeout:  cfg.Timeout,
//		WriteTimeout: cfg.Timeout,
//	})
//
//	if err := db.Ping(ctx).Err(); err != nil {
//		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
//		return nil, err
//	}
//
//	return db, nil
//}
//
//var dbRed *redis.Client
//
//func init() {
//	cfg := Config{
//		Addr:        "localhost:6379",
//		Password:    "test1234",
//		User:        "testuser",
//		DB:          0,
//		MaxRetries:  5,
//		DialTimeout: 10 * time.Second,
//		Timeout:     5 * time.Second,
//	}
//
//	db, err := NewClient(context.Background(), cfg)
//	if err != nil {
//		panic(err)
//	}
//	dbRed = db
//}
//
//type DBManager struct {
//}
//
//func (receiver DBManager) GetFromdb() {
//
//}
