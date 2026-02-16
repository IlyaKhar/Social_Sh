package handlers

import (
	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/models"
)

// GetGalleryItems отдает изображения для галереи.
// query:
//   - category=intro / tattoo / tokyo / ...
func GetGalleryItems(c *fiber.Ctx) error {
	category := c.Query("category", "")

	items, err := Repo.Gallery.ListByCategory(category)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении элементов галереи",
		})
	}

	// Если items == nil, возвращаем пустой массив (не ошибка)
	if items == nil {
		items = []models.GalleryItem{}
	}

	return c.JSON(fiber.Map{"items": items})
}
