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

	AdminToken string // protects /api/v1/admin; admin API is closed when empty
	PublicURL  string // base URL devices download assets from
}

func MustLoad() Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT is required.")
	}

	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV is required.")

	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		panic("DATABASE_URL is required.")

	}
	cfg := Config{
		Port:  port,
		Env:   env,
		DBUrl: dbUrl,

		AdminToken: os.Getenv("ADMIN_TOKEN"),
		PublicURL:  os.Getenv("PUBLIC_URL"),
	}
	if cfg.PublicURL == "" {
		cfg.PublicURL = "http://localhost:" + port
	}

	return cfg
}
