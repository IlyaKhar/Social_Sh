package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

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

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// TODO: обработай ошибку и верни 500/400 в зависимости от ситуации.
	items, _ := Repo.Products.List(newOnly, saleOnly, page, limit)

	return c.JSON(fiber.Map{"items": items})
}

// GetProduct отдает один товар по slug.
func GetProduct(c *fiber.Ctx) error {
	slug := c.Params("slug")

	item, _ := Repo.Products.GetBySlug(slug)
	if item == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(fiber.Map{"item": item})
}

