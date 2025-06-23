package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/go-redis/redis/v8"

	"web-crawler/internal/myredis"
	"web-crawler/internal/scraper"
)

func connectToRedis(maxAttempts int) (*redis.Client) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect with attempt %d", attempt)
		rdb := redis.NewClient(&redis.Options{
			Addr: addr,
			Password: "",
			DB: 0,
		});

		ctxt, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		result, err := rdb.Ping(ctxt).Result()
		cancel()

		if err == nil && result == "PONG" {
			log.Printf("Connected to redis server with pong: %s", result)
			return rdb
		}
		log.Printf("Failed to ping the redis server")
		rdb.Close()

		log.Printf("Sleeping for 2 seconds")
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Fatalf("Failed to connect to Redis after %d attempts", maxAttempts)
	return nil
}

func main() {
	fmt.Println("Trying to connect to redis server")
	rdb := connectToRedis(10)
	defer rdb.Close()

	fmt.Println("Building queue/set")
	redisQueue := myredis.NewRedisQueue(rdb, "websites")
	redisSet := myredis.NewRedisSet(rdb, "visited")
	redisQueue.CheckQueue()

	fmt.Println("Starting scraping")
	scraper.ScrapePage(redisQueue, redisSet)
}
