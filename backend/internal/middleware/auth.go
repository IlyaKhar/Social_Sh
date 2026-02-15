package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected — middleware для проверки JWT-токена.
// Достаёт токен из заголовка Authorization: Bearer <token>,
// валидирует его, и кладёт userID + role в c.Locals.
// Если токен невалидный или отсутствует — возвращает 401.
//
// Использование в routes:
//
//	api.Group("/account", middleware.Protected(cfg.JwtSecret))
func Protected(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Достаём заголовок Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{ //401
				"error": "отсутствует токен авторизации",
			})
		}

		// Ожидаем формат "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{ //401
				"error": "неверный формат токена, ожидается Bearer <token>",
			})
		}

		tokenString := parts[1]

		// Парсим и валидируем JWT
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// Проверяем метод подписи — только HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "неверный метод подписи токена")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "невалидный или просроченный токен",
			})
		}

		// Достаём claims (userID и role)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "не удалось прочитать claims из токена",
			})
		}

		// Кладём данные в контекст — хендлеры достанут через c.Locals("userID")
		userID, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)

		c.Locals("userID", userID)
		c.Locals("role", role)

		// Передаём управление следующему хендлеру
		return c.Next()
	}
}

// AdminOnly — middleware, проверяющий что у юзера роль admin.
// Вешается ПОСЛЕ Protected(), чтобы role уже лежал в c.Locals.
//
// Использование:
//
//	api.Group("/admin", middleware.Protected(secret), middleware.AdminOnly())
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "доступ только для администраторов",
			})
		}
		return c.Next()
	}
}
