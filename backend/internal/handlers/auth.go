package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"socialsh/backend/internal/models"
)

// ──── Хелперы ────

// hashPassword — хеширует пароль через bcrypt.
// cost=10 — стандартный баланс скорости/безопасности.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// checkPassword — сравнивает хеш с паролем.
func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// generateTokens — генерирует пару access + refresh JWT.
// access: живёт 15 минут, claims: sub (userID), role.
// refresh: живёт 7 дней, claims: sub (userID).
func generateTokens(userID, role, jwtSecret, refreshSecret string) (*models.TokenResponse, error) {
	// Access-токен (15 мин)
	accessClaims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	access, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("не удалось подписать access-токен: %w", err)
	}

	// Refresh-токен (7 дней)
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refresh, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("не удалось подписать refresh-токен: %w", err)
	}

	return &models.TokenResponse{Access: access, Refresh: refresh}, nil
}

// ──── Хендлеры ────

// SignUp — регистрация нового пользователя.
// POST /api/auth/sign-up
// Body: { "email": "...", "password": "...", "name": "..." }
// Ответ 201: { "access": "<jwt>", "refresh": "<jwt>" }
//
// Логика:
//  1. Парсим body → SignUpRequest
//  2. Валидируем поля (email, password длина)
//  3. Проверяем что email свободен → Repo.Account.GetUserByEmail
//  4. Хешируем пароль → bcrypt
//  5. Создаём юзера → Repo.Account.CreateUser
//  6. Генерируем access + refresh JWT
//  7. Возвращаем токены
func SignUp(jwtSecret, refreshSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.SignUpRequest

		// 1. Парсим JSON-body
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "невалидный JSON"})
		}

		// 2. Валидация полей
		if req.Email == "" {
			return c.Status(400).JSON(fiber.Map{"error": "email обязателен"})
		}
		if len(req.Password) < 8 {
			return c.Status(400).JSON(fiber.Map{"error": "пароль должен содержать минимум 8 символов"})
		}
		if len(req.Password) > 72 {
			return c.Status(400).JSON(fiber.Map{"error": "пароль не должен превышать 72 символа"})
		}

		// 3. Проверяем что email не занят
		existingUser, err := Repo.Account.GetUserByEmail(req.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			// Ошибка БД (не "не найдено")
			return c.Status(500).JSON(fiber.Map{"error": "ошибка при проверке email"})
		}
		if existingUser != nil {
			return c.Status(409).JSON(fiber.Map{"error": "пользователь с таким email уже существует"})
		}

		// 4. Хешируем пароль
		hashed, err := hashPassword(req.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "не удалось захешировать пароль"})
		}

		// 5. Создаём юзера в БД
		user := &models.User{
			Email:        req.Email,
			Name:         req.Name,
			PasswordHash: hashed,
			Role:         "user", // по умолчанию — обычный юзер
		}

		if err := Repo.Account.CreateUser(user); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "не удалось создать пользователя"})
		}

		// 6. Генерируем JWT-токены
		tokens, err := generateTokens(user.ID, user.Role, jwtSecret, refreshSecret)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "не удалось сгенерировать токены"})
		}

		// 7. Возвращаем токены
		return c.Status(201).JSON(tokens)
	}
}

// SignIn — логин существующего пользователя.
// POST /api/auth/sign-in
// Body: { "email": "...", "password": "..." }
// Ответ 200: { "access": "<jwt>", "refresh": "<jwt>" }
//
// Логика:
//  1. Парсим body → SignInRequest
//  2. Ищем юзера по email
//  3. Сравниваем пароль через bcrypt
//  4. Генерируем access + refresh JWT
//  5. Возвращаем токены
func SignIn(jwtSecret, refreshSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.SignInRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "невалидный JSON"})
		}

		if req.Email == "" || req.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "email и пароль обязательны"})
		}

		// Ищем юзера по email
		user, err := Repo.Account.GetUserByEmail(req.Email)
		if err != nil || user == nil {
			// Не говорим конкретно "email не найден" — чтобы не палить существование аккаунтов
			return c.Status(401).JSON(fiber.Map{"error": "неверный email или пароль"})
		}

		// Сравниваем пароль с хешем
		if !checkPassword(user.PasswordHash, req.Password) {
			return c.Status(401).JSON(fiber.Map{"error": "неверный email или пароль"})
		}

		// Генерируем токены
		tokens, err := generateTokens(user.ID, user.Role, jwtSecret, refreshSecret)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "не удалось сгенерировать токены"})
		}

		return c.JSON(tokens)
	}
}

// RefreshToken — обмен refresh-токена на новый access.
// POST /api/auth/refresh
// Body: { "refresh": "<jwt>" }
// Ответ 200: { "access": "<новый jwt>", "refresh": "<тот же или новый>" }
func RefreshToken(jwtSecret, refreshSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.RefreshRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "невалидный JSON"})
		}

		if req.Refresh == "" {
			return c.Status(400).JSON(fiber.Map{"error": "refresh-токен обязателен"})
		}

		// Парсим и валидируем refresh-токен
		token, err := jwt.Parse(req.Refresh, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("неверный метод подписи")
			}
			return []byte(refreshSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "невалидный или просроченный refresh-токен"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "не удалось прочитать claims"})
		}

		userID, _ := claims["sub"].(string)

		// Проверяем что юзер ещё существует
		user, err := Repo.Account.GetUserByID(userID)
		if err != nil || user == nil {
			return c.Status(401).JSON(fiber.Map{"error": "пользователь не найден"})
		}

		// Генерируем новую пару токенов
		tokens, err := generateTokens(user.ID, user.Role, jwtSecret, refreshSecret)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "не удалось сгенерировать токены"})
		}

		return c.JSON(tokens)
	}
}

// Logout — инвалидация сессии.
// POST /api/auth/logout (через Protected middleware — userID в c.Locals)
// Ответ 200: { "message": "ok" }
func Logout(c *fiber.Ctx) error {
	// TODO: если хранишь refresh-токены в БД (таблица refresh_tokens) —
	// удали строку по userID := c.Locals("userID").(string)
	// Пока просто возвращаем ok (клиент удалит токены на своей стороне)
	return c.JSON(fiber.Map{"message": "ok"})
}

// IsAdmin — проверяет роль текущего юзера.
// GET /api/auth/is-admin (через Protected middleware — role в c.Locals)
// Ответ: { "isAdmin": true/false }
func IsAdmin(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	return c.JSON(fiber.Map{"isAdmin": role == "admin"})
}
