package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/models"
	"socialsh/backend/internal/utils"
)

// ═══════════════════════════════════════════════════════════════
// Хендлеры личного кабинета.
// Все роуты в группе /api/account защищены middleware.Protected() —
// к моменту вызова этих хендлеров userID уже лежит в c.Locals("userID").
// ═══════════════════════════════════════════════════════════════

// GetAccountMe — возвращает профиль текущего пользователя.
// GET /api/account/me
// Ответ: { "user": { "id", "email", "name", "role" } }
func GetAccountMe(c *fiber.Ctx) error {
	// userID положен middleware.Protected() в c.Locals
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "не авторизован",
		})
	}

	user, err := Repo.Account.GetUserByID(userID)
	if err != nil {
		// Если пользователь не найден (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "пользователь не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении профиля",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	return c.JSON(fiber.Map{"user": user})
}

// GetOrders — возвращает список заказов текущего пользователя.
// GET /api/account/orders
// Ответ: { "items": [ { "id", "userId", ... }, ... ] }
func GetOrders(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "не авторизован",
		})
	}

	orders, err := Repo.Account.ListOrdersByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении заказов",
		})
	}

	// Если orders == nil, возвращаем пустой массив (не ошибка)
	if orders == nil {
		orders = []models.Order{}
	}

	return c.JSON(fiber.Map{"items": orders})
}

// UpdateProfile — обновление профиля текущего пользователя.
// PATCH /api/account/me
// Body: { "name": "Новое имя", "email": "new@mail.com" } — только изменённые поля.
// Ответ: { "user": { обновлённый профиль } }
func UpdateProfile(c *fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)

	// Парсим body — указатели в DTO позволяют отличить "не прислали" (nil) от "пустая строка"
	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невалидный JSON",
		})
	}

	// Проверяем что хоть что-то прислали
	if req.Name == nil && req.Email == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "нечего обновлять — передай name и/или email",
		})
	}

	// Валидация: если обновляется email, проверяем что он не занят другим пользователем
	if req.Email != nil && *req.Email != "" {
		existingUser, err := Repo.Account.GetUserByEmail(*req.Email)
		if err != nil && err != sql.ErrNoRows {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ошибка при проверке email",
			})
		}
		// Если email занят другим пользователем (не текущим)
		if existingUser != nil && existingUser.ID != userID {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "email уже используется другим пользователем",
			})
		}
	}

	// Обновляем юзера в БД и получаем обновлённую версию
	user, err := Repo.Account.UpdateUser(userID, &req)
	if err != nil {
		// Проверяем на duplicate key (email уже занят)
		if utils.IsDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": utils.FormatDuplicateError(err),
			})
		}
		// Если пользователь не найден
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "пользователь не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось обновить профиль",
		})
	}

	return c.JSON(fiber.Map{"user": user})
}
