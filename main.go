package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	redisURL := os.Getenv("REDIS_URL")
	if len(redisURL) == 0 {
		redisURL = "redis://:@localhost:6379/1"
	}

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	redisCache := cache.New(&cache.Options{
		Redis: redisClient,
	})
	server := &Server{
		RedisCache: redisCache,
	}

	addr := ":" + port

	fmt.Printf("-= Starting web server, listening on %s =-\n", addr)

	err = http.ListenAndServe(addr, server)
	if err != nil {
		panic(err)
	}
}
