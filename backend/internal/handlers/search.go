package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/models"
)

// SearchProducts — поиск товаров по названию.
// GET /api/products/search?q=hoodie&page=1&limit=20
// Ответ: { "items": [ ... ], "total": 10 }
func SearchProducts(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q", ""))
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "параметр q (поисковый запрос) обязателен",
		})
	}

	// Парсим page и limit
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}

	// Ищем товары через репозиторий
	items, err := Repo.Products.Search(query, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при поиске товаров",
		})
	}

	if items == nil {
		items = []models.Product{}
	}

	return c.JSON(fiber.Map{"items": items})
}
