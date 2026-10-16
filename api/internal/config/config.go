package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	// DemoReset пересоздаёт демо-данные при каждом старте API.
	DemoReset bool
	// WorkerEvery — период, с которым воркер «забирает отчёты» площадок.
	WorkerEvery time.Duration
	CORSOrigin  string
}

func Load() Config {
	return Config{
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://crm:crm@localhost:5432/crm?sslmode=disable"),
		DemoReset:   envBool("DEMO_RESET", true),
		WorkerEvery: envDuration("WORKER_EVERY", 30*time.Second),
		CORSOrigin:  env("CORS_ORIGIN", "*"),
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}

func envDuration(key string, def time.Duration) time.Duration {
	v, err := time.ParseDuration(os.Getenv(key))
	if err != nil || v <= 0 {
		return def
	}
	return v
}
