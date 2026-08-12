package configs

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Db          DbConfig
	Http        HttpConfig
	Auth        AuthConfig
	CORS        CORSConfig
}

type DbConfig struct {
	DbUrl string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type AuthConfig struct {
	Secret string
}

type HttpConfig struct {
	Port string
	Host string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error file .env", err.Error())
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "local"
	}
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	allowedOrigins := parseCommaSeparatedList(corsOrigins)
	if len(allowedOrigins) == 0 && environment == "local" {
		allowedOrigins = []string{
			"http://localhost:3000",
		}
	}

	return &Config{
		Environment: environment,
		Db: DbConfig{
			DbUrl: os.Getenv("DATABASE_URL"),
		},
		Http: HttpConfig{
			Port: port,
			Host: os.Getenv("HOST"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("TOKEN"),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
	}
}

func parseCommaSeparatedList(value string) []string {
	if value == "" {
		return []string{}
	}
	items := strings.Split(value, ",")

	result := make([]string, 0, len(items))

	for _, item := range items {
		origin := strings.TrimSpace(item)
		if origin == "" {
			continue
		}
		result = append(result, origin)
	}

	return result
}
