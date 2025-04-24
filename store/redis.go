package store

import (
	"log"
	"meals/config"

	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func InitRedis() {
	// Use the configuration from the config package
	redisConfig := config.AppConfig.Redis

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	_, err := RedisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")
}
