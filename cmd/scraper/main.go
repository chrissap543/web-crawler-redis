package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"

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

func connectToNeo4j(maxAttempts int) (neo4j.DriverWithContext) {
	neo4jURI := os.Getenv("NEO4J_URI")
    neo4jUser := os.Getenv("NEO4J_USER")
    neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("Attempting to connect to Neo4j with attempt %d", attempt)

		driver, err := neo4j.NewDriverWithContext(
			neo4jURI,
			neo4j.BasicAuth(neo4jUser, neo4jPassword, ""),
			func(config *config.Config) {
				config.MaxConnectionLifetime = 30 * time.Minute
				config.MaxConnectionPoolSize = 50
				config.ConnectionAcquisitionTimeout = 2 * time.Minute
			},
		)

		if err != nil {
			log.Printf("Failed to create Neo4j driver: %v", err)
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		err = driver.VerifyConnectivity(ctx)
		cancel()

		if err != nil {
			log.Printf("Connected to Neo4j server successfully")
			return driver
		}


		log.Printf("Failed to ping Neo4j server")
		time.Sleep(time.Duration(2) * time.Second)
	}
	log.Fatalf("Failed to connect after %d attempts", maxAttempts)
	return nil
}

func main() {
	log.Println("Trying to connect to redis server")
	rdb := connectToRedis(10)
	defer rdb.Close()

	log.Println("Building queue/set")
	redisQueue := myredis.NewRedisQueue(rdb, "websites")
	redisSet := myredis.NewRedisSet(rdb, "visited")
	redisQueue.CheckQueue()

	log.Println("Connecting to Neo4j")
	neo4jdriver := connectToNeo4j(10)

	log.Println("Starting scraping")
	scraper.ScrapePage(redisQueue, redisSet, neo4jdriver)
}
