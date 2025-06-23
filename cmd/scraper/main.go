package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/go-redis/redis/v8"

	"web-crawler/internal/myredis"
	"web-crawler/internal/scraper"
)

func connectToRedis() (*redis.Client) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB: 0,
	});
	ctxt := context.Background()

	pong, err := rdb.Ping(ctxt).Result()
	if err != nil {
		log.Fatalf("Failed to ping the redis server")
	}
	fmt.Printf("Connected to redis server with pong: %s\n", pong)

	return rdb
}

func main() {
	fmt.Println("Trying to connect to redis server")
	rdb := connectToRedis()

	fmt.Println("Building queue/set")
	redisQueue := myredis.NewRedisQueue(rdb, "websites")
	redisSet := myredis.NewRedisSet(rdb, "visited")
	redisQueue.CheckQueue()

	fmt.Println("Starting scraping")
	scraper.ScrapePage(redisQueue, redisSet)
}
