package config

import (
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	Env   string
	DBUrl string

	PublicURL string
}

// Defaults match the Postgres started by `make up`, so the server runs
// locally without a .env file.
const (
	defaultPort  = "8080"
	defaultEnv   = "development"
	defaultDBUrl = "postgres://ota:ota@localhost:5433/ota?sslmode=disable"
)

func MustLoad() Config {
	godotenv.Load()

	port := getenv("PORT", defaultPort)

	cfg := Config{
		Port:  port,
		Env:   getenv("ENV", defaultEnv),
		DBUrl: getenv("DATABASE_URL", defaultDBUrl),

		PublicURL: getenv("PUBLIC_URL", "http://localhost:"+port),
	}

	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
