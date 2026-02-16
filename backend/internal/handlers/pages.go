package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

// GetPage отдает статическую страницу (оплата, доставка, возврат, контакты).
// slug:
//   - payment
//   - delivery
//   - returns
//   - contacts
func GetPage(c *fiber.Ctx) error {
	slug := c.Params("slug")

	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug обязателен",
		})
	}

	// Валидация slug (опционально, но полезно)
	validSlugs := map[string]bool{
		"payment":  true,
		"delivery": true,
		"returns":  true,
		"contacts": true,
	}
	if !validSlugs[slug] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "недопустимый slug",
		})
	}

	page, err := Repo.Pages.GetBySlug(slug)
	if err != nil {
		// Если страница не найдена (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "страница не найдена",
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении страницы",
		})
	}

	if page == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "страница не найдена",
		})
	}

	return c.JSON(page)
}
