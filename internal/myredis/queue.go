package myredis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisQueue struct {
	client *redis.Client
	ctx context.Context
	queueName string
}

func NewRedisQueue(client *redis.Client, name string) *RedisQueue {
	return &RedisQueue {
		client: client,
		ctx: context.Background(),
		queueName: name,
	}
}

func (q *RedisQueue) Enqueue(item string) error {
	return q.client.LPush(q.ctx, q.queueName, item).Err()
}

func (q *RedisQueue) Dequeue(timeout time.Duration) (string, error) {
	result, err := q.client.BRPop(q.ctx, timeout, q.queueName).Result()
	if err != nil {
		return "", err
	}
	return result[1], err
}

func (q *RedisQueue) Length() (int64, error) {
	return q.client.LLen(q.ctx, q.queueName).Result()
}

func (q *RedisQueue) CheckQueue() (error) {
	length, err := q.Length()
	if err != nil {
		return err
	}

	if length == 0 {
		starting_url := "https://simple.wikipedia.org/wiki/Anime"
		fmt.Println("Queue is empty, give it starting_url")
		return q.Enqueue(starting_url)
	}

	return nil
}
