package handlers

import (
	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/repository"
)

// GetGalleryItems отдает изображения для галереи.
// query:
//   - category=intro / tattoo / tokyo / ...
func GetGalleryItems(c *fiber.Ctx) error {
	category := c.Query("category", "")

	items, _ := Repo.Gallery.ListByCategory(category)

	return c.JSON(fiber.Map{"items": items})
}

