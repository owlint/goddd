package services

//go:generate go run github.com/golang/mock/mockgen -destination=../mocks/mock_queue_service.go -package=mocks github.com/owlint/goddd/services QueueService

import (
	"context"

	"github.com/go-redis/redis/v9"
)

type QueueService interface {
	Push(context.Context, []byte) error
	Pop(context.Context) ([]byte, error)
}

type RedisQueueService struct {
	client    *redis.Client
	queueName string
}

func NewRedisQueueService(client *redis.Client, queueName string) QueueService {
	return &RedisQueueService{
		client:    client,
		queueName: queueName,
	}
}

func (r *RedisQueueService) Push(ctx context.Context, value []byte) error {
	return r.client.LPush(ctx, r.queueName, value).Err()
}

func (r *RedisQueueService) Pop(ctx context.Context) ([]byte, error) {
	values, err := r.client.BRPop(ctx, 0, r.queueName).Result()
	if err != nil {
		return nil, err
	}

	return []byte(values[1]), nil
}
