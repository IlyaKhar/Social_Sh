package handlers

import (
	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/repository"
)

// GetPage отдает статическую страницу (оплата, доставка, возврат, контакты).
// slug:
//   - payment
//   - delivery
//   - returns
//   - contacts
func GetPage(c *fiber.Ctx) error {
	slug := c.Params("slug")

	page, _ := Repo.Pages.GetBySlug(slug)
	if page == nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(page)
}

