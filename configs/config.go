package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Db          DbConfig
	Http        HttpConfig
	Auth        AuthConfig
}

type DbConfig struct {
	DbUrl string
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
	}
}
