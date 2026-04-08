package config

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
)

var (
	RedisClient *asynq.Client
	RedisOpt    asynq.RedisClientOpt
)

func ConnectRedis() {

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatalf("Failed to connect to Redis: REDIS_URL is not set")
	}

	RedisOpt = asynq.RedisClientOpt{
		Addr: redisURL,
	}

	RedisClient = asynq.NewClient(RedisOpt)

	log.Println("Successfully connected to Redis")
}