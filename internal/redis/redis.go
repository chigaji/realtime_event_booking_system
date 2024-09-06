package redis

import (
	"context"
	"fmt"

	"github.com/chigaji/realtime_event_booking_system/internal/config"
	"github.com/go-redis/redis/v8"
)

func Init(cfg config.RedisConfig) (*redis.Client, error) {

	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Username: cfg.Username,
		Password: "",
		DB:       cfg.DB,
	})

	fmt.Println("redis infor ===>", cfg.Address, cfg.Password)

	pong, err := client.Ping(ctx).Result()

	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to Redis:", pong)

	return client, nil
}
