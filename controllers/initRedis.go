package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eezz10001/ego/core/elog"
	"github.com/go-redis/redis/v8"
)

var Redis = new(redis.Client)

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "60.205.125.113:6379",
		Password: "1234567aB", // no password set
		DB:       1,           // use default DB
	})

	pong, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		elog.Error("redis init fail", elog.FieldErr(err))
		return
	}
	elog.Info("redis init success", elog.FieldValue(pong))
}

func ObjPublish(obj interface{}) {
	b, _ := json.Marshal(obj)
	fmt.Println(string(b))
	err := Redis.Publish(context.Background(), "experiment", string(b)).Err()
	if err != nil {
		fmt.Println(err)
	}
}
