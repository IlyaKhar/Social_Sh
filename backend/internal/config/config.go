package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config — базовый конфиг сервера.
// TODO: добавь сюда все, что нужно (DSN БД, секреты JWT и т.п.).
type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JwtSecret     string
	RefreshSecret string
	BaseUrl       string
}

// Load читает конфиг из переменных окружения.
// Пример: PORT=8080 go run ./cmd/api
func Load() *Config {
	// Пытаемся загрузить .env из разных мест
	// godotenv.Load() ищет .env в текущей рабочей директории
	// Если запускаешь из backend/ - найдёт backend/.env
	// Если запускаешь из корня - нужно указать путь явно
	godotenv.Load()                  // из текущей директории
	godotenv.Load(".env")            // явно из текущей директории
	godotenv.Load("../backend/.env") // если запускаешь из корня проекта

	return &Config{
		Port:          LoadEnv("PORT", "3000"),
		DBHost:        LoadEnv("DB_HOST", "localhost"),
		DBPort:        LoadEnv("DB_PORT", "5432"),
		DBUser:        LoadEnv("DB_USER", "postgres"),
		DBPassword:    LoadEnv("DB_PASSWORD", ""),
		DBName:        LoadEnv("DB_NAME", "postgres"), // дефолт "postgres", но должен быть "socialsh" из .env
		JwtSecret:     LoadEnv("JWT_SECRET", "not found"),
		RefreshSecret: LoadEnv("REFRESH_SECRET", "not found"),
		BaseUrl:       LoadEnv("BASE_URL", "http://localhost:3000"),
	}
}

func LoadEnv(key string, replacment string) string {
	res := os.Getenv(key)

	if res == "" {
		return replacment
	}
	return res
}

// это у нас метод-хелпер чтобы не размазывать сборку dsn по коду
func (c *Config) PostgresDSN() string {
	// Если пароль пустой, не включаем его в DSN (для peer authentication на macOS)
	if c.DBPassword == "" {
		return fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=disable",
			c.DBHost,
			c.DBPort,
			c.DBUser,
			c.DBName,
		)
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}
