package services_test

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v9"
	"github.com/owlint/goddd/services"
	"github.com/owlint/goddd/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRedisQueuePushPop(t *testing.T) {
	testutils.WithTestRedis(func(conn *redis.Client) {
		queue := services.NewRedisQueueService(conn, "test")
		value := []byte("Hello")

		err := queue.Push(context.Background(), value)
		assert.NoError(t, err)

		res, err := queue.Pop(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, value, res)
	})
}

func TestRedisQueuePushPopMultiple(t *testing.T) {
	testutils.WithTestRedis(func(conn *redis.Client) {
		queue := services.NewRedisQueueService(conn, "test")
		hello := []byte("Hello")
		world := []byte("world")

		err := queue.Push(context.Background(), hello)
		assert.NoError(t, err)

		err = queue.Push(context.Background(), world)
		assert.NoError(t, err)

		res, err := queue.Pop(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, hello, res)

		res, err = queue.Pop(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, world, res)
	})
}
