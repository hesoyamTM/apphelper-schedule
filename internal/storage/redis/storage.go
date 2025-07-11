package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	client *redis.Client
}

type RedisConfig struct {
	Host string `yaml:"host" env-required:"true" env:"REDIS_HOST"`
	Port int    `yaml:"port" env-required:"true" env:"REDIS_PORT"`
	Pass string `yaml:"pass" env-required:"true" env:"REDIS_PASS"`
}

func New(cfg RedisConfig) *Storage {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Pass,
		DB:       0,
	})

	return &Storage{
		client: client,
	}
}
