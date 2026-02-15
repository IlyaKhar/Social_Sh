package handlers

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/models"
	"socialsh/backend/internal/repository"
)

// Repo будет инициализирован из main.
// TODO: в проде лучше не держать это глобально, а прокинуть через зависимости/структуры.
var Repo *repository.Store

// GetProducts отдает список товаров.
// query:
//   - new=true  -> новые поступления
//   - sale=true -> сезонные скидки
//   - page, limit для пагинации
func GetProducts(c *fiber.Ctx) error {
	newOnly := c.QueryBool("new", false)
	saleOnly := c.QueryBool("sale", false)

	// Парсим page и limit, если не число — используем дефолты
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}

	items, err := Repo.Products.List(newOnly, saleOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении списка товаров",
		})
	}

	// Если items == nil, возвращаем пустой массив (не ошибка)
	if items == nil {
		items = []models.Product{}
	}

	return c.JSON(fiber.Map{"items": items})
}

// GetProduct отдает один товар по slug.
func GetProduct(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug обязателен",
		})
	}

	item, err := Repo.Products.GetBySlug(slug)
	if err != nil {
		// Если товар не найден (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "товар не найден",
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении товара",
		})
	}

	return c.JSON(fiber.Map{"item": item})
}
