package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config — базовый конфиг сервера.
// TODO: добавь сюда все, что нужно (DSN БД, секреты JWT и т.п.).
type Config struct {
	Port string
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	JwtSecret string
	RefreshSecret string
	BaseUrl string
}

// Load читает конфиг из переменных окружения.
// Пример: PORT=8080 go run ./cmd/api
func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:          LoadEnv("PORT", "3000"),
		DBHost:        LoadEnv("DB_HOST", "localhost"),
		DBPort:        LoadEnv("DB_PORT", "5432"),
		DBUser:        LoadEnv("DB_USER", "postgres"),
		DBPassword:    LoadEnv("DB_PASSWORD", ""),
		DBName:        LoadEnv("DB_NAME", "postgres"),
		JwtSecret:     LoadEnv("JWT_SECRET", "not found"),
		RefreshSecret: LoadEnv("REFRESH_SECRET", "not found"),
		BaseUrl:       LoadEnv("BASE_URL", "http://localhost:3000"),
	}
}

func LoadEnv(key string, replacment string) string {
	res := os.Getenv(key)

	if res == ""{
		return replacment
	}
	return res
} 


// это у нас метод-хелпер чтобы не размазывать сборку dsn по коду 
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}