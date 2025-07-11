package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hesoyamTM/apphelper-notification/pkg/redpanda"
	"github.com/hesoyamTM/apphelper-schedule/internal/clients"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage/psql"
	"github.com/hesoyamTM/apphelper-schedule/internal/storage/redis"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string        `yaml:"env" env-required:"true" env:"ENV"`
	StateTTL time.Duration `yaml:"state-ttl" env-required:"true" env:"STATE_TTL"`

	Grpc           GRPC                      `yaml:"grpc"`
	Psql           psql.PsqlConfig           `yaml:"psql"`
	Redis          redis.RedisConfig         `yaml:"redis"`
	GoogleCalendar clients.GoogleCalendarCfg `yaml:"google-calendar"`
	Redpanda       redpanda.RedpandaConfig   `yaml:"redpanda"`
}

type GRPC struct {
	Host string `yaml:"host" env-required:"true" env:"GRPC_HOST"`
	Port int    `yaml:"port" env-required:"true" env:"GRPC_PORT"`
}

func fetchConfigPath() string {
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "", "config path")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	return cfgPath
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("failed to load .env")
	}

	cfgPath := fetchConfigPath()
	if cfgPath != "" {
		return MustLoadByPath(cfgPath)
	}

	return MustLoadEnv()
}

func MustLoadEnv() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func MustLoadByPath(cfgPath string) *Config {
	if cfgPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic("config file does not exist: " + err.Error())
		}

		panic(err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
