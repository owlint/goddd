package services

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v9"
)

func NewRedisClient(host, password string, database, port int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})
	_, err := client.Ping(context.TODO()).Result()

	if err != nil {
		panic(err)
	}

	return client
}
