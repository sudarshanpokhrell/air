package config

import (
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/sudarshanpokhrell/air/internal/storage"
)

type Config struct {
	Port  string
	Env   string
	DBUrl string

	PublicURL string

	Storage storage.S3Config
}

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

		Storage: storage.S3Config{
			Endpoint:        os.Getenv("S3_ENDPOINT"),
			Region:          getenv("S3_REGION", "auto"),
			Bucket:          os.Getenv("S3_BUCKET"),
			AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
			AssetBaseURL:    os.Getenv("ASSET_BASE_URL"),
		},
	}

	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
