package myredis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisSet struct {
	client *redis.Client
	ctx context.Context
	setName string
}

func NewRedisSet(client *redis.Client, name string) *RedisSet {
	return &RedisSet {
		client: client,
		ctx: context.Background(),
		setName: name,
	}
}

func (s *RedisSet) IsMember(name string) (bool, error) {
	return s.client.SIsMember(s.ctx, s.setName, name).Result();
}

func (s *RedisSet) Add(name string) (error) {
	return s.client.SAdd(s.ctx, s.setName, name).Err();
}
func (s *RedisSet) Remove(name string) (error) {
	return s.client.SRem(s.ctx, s.setName, name).Err();
}

