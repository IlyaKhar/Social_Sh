package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/config"
	"socialsh/backend/internal/db"
	"socialsh/backend/internal/handlers"
	"socialsh/backend/internal/repository"
	"socialsh/backend/internal/routes"
)

// TODO: допиши конфиг (порт, env, логгер, подключение к БД).

func main() {
	cfg := config.Load()

	// Подключаемся к Postgres
	dsn := cfg.PostgresDSN()
	sqlDB, err := db.OpenPostgres(dsn)
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}
	defer sqlDB.Close()

	// Инициализируем репозитории и прокидываем их в хендлеры
	store := repository.NewStore(sqlDB)
	handlers.Repo = store

	app := fiber.New()

	// TODO: middlewares (логгер, cors, recover, auth и т.д.).

	routes.Register(app, cfg.JwtSecret, cfg.RefreshSecret)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("fiber listen: %v", err)
	}
}
