package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eezz10001/ego/core/elog"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	"github.com/go-redis/redis/v8"
)

var Redis = new(redis.Client)

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "27.132.46.38:6379",
		Password: "1234567aB", // no password set
		DB:       10,          // use default DB
	})

	pong, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		elog.Error("redis init fail", elog.FieldErr(err))
		return
	}
	elog.Info("redis init success", elog.FieldValue(pong))
}

func ObjPublish(obj *experimentv1.Experiment) {

	if obj.Status.Phase == "Running" {
		b, err := json.Marshal(obj)
		fmt.Println("publish obj", string(b))
		if err != nil {
			fmt.Println(err)
		}
		err = Redis.Publish(context.Background(), "experiment", string(b)).Err()
		if err != nil {
			fmt.Println(err)
		}

	}
}
