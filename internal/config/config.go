package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	Redis       `yaml:"redis"`
	JWTSecret   []byte
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"1m"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

type Redis struct {
	Address  string        `yaml:"address" env-default:"localhost:6379"`
	Password string        `yaml:"password" env-required:"true"`
	DB       int           `yaml:"db" env-default:"0"`
	TTL      time.Duration `yaml:"ttl" env-default:"24h"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to open config file: %v", err)
	}

	var config Config
	err := cleanenv.ReadConfig(configPath, &config)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	if len(jwtSecret) != 32 {
		log.Fatal("JWT_SECRET must be 32 bytes long")
	}

	config.JWTSecret = []byte(jwtSecret)
	return &config
}
