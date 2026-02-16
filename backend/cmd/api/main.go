package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"socialsh/backend/internal/config"
	"socialsh/backend/internal/db"
	"socialsh/backend/internal/handlers"
	"socialsh/backend/internal/repository"
	"socialsh/backend/internal/routes"
)

// @title           SOCIAL SH API
// @version         1.0
// @description     API для интернет-магазина SOCIAL SH
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  info@socialsh.ru

// @host      localhost:3001
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Введите "Bearer {token}" для авторизации

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

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Статические файлы (загруженные изображения)
	app.Static("/uploads", "./uploads")

	// Swagger JSON - читаем файл и отдаём
	app.Get("/docs/swagger.json", func(c *fiber.Ctx) error {
		// Пробуем разные пути
		paths := []string{
			filepath.Join(".", "docs", "swagger.json"),
			filepath.Join("..", "docs", "swagger.json"),
			filepath.Join("backend", "docs", "swagger.json"),
		}

		var data []byte
		var err error
		for _, p := range paths {
			data, err = os.ReadFile(p)
			if err == nil {
				log.Printf("✅ Swagger JSON найден: %s", p)
				break
			}
		}

		if err != nil {
			log.Printf("❌ Swagger JSON не найден. Пробовали: %v, ошибка: %v", paths, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "swagger.json not found",
			})
		}

		c.Set("Content-Type", "application/json")
		return c.Send(data)
	})

	// Swagger UI
	swaggerHTML := `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>SOCIAL SH API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/docs/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout",
        validatorUrl: null,
        docExpansion: "list",
        filter: true
      });
    };
  </script>
</body>
</html>`

	app.Get("/swagger", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})
	app.Get("/swagger/*", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})

	routes.Register(app, cfg.JwtSecret, cfg.RefreshSecret)

	log.Printf("🚀 Server starting on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("fiber listen: %v", err)
	}
}
