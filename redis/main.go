package main

import (
	"context"

	"github.com/go-redis/redis"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	RunStringValueDemos(rdb)

}
