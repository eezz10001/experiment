package controllers

import (
	"context"
	"fmt"
	"github.com/eezz10001/ego/core/elog"
	"github.com/go-redis/redis/v8"
)

var Redis = new(redis.Client)

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "27.132.46.38:6379",
		Password: "1234567aB", // no password set
		DB:       10,          // use default DB
	})
	fmt.Println("redis log 27.132.46.38:6379 ")
	pong, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		elog.Error("redis init fail", elog.FieldErr(err))
		return
	}
	elog.Info("redis init success", elog.FieldValue(pong))
}
