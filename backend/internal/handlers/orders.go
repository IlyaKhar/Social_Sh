package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CreateOrderRequest — структура запроса на создание заказа
type CreateOrderRequest struct {
	Items []struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
		Price     int64  `json:"price"`
	} `json:"items"`
	Customer struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone,omitempty"`
		Telegram string `json:"telegram,omitempty"`
		Address  string `json:"address,omitempty"`
	} `json:"customer"`
	Comment string `json:"comment,omitempty"`
	Total   int64  `json:"total"`
}

// CreateOrder — создание заказа с отправкой уведомления.
// POST /api/orders
// Body: CreateOrderRequest
// Ответ: { "message": "Заказ создан", "orderId": "..." }
//
// Логика:
//  1. Парсим body → CreateOrderRequest
//  2. Валидируем обязательные поля (name, email, items)
//  3. Формируем сообщение для отправки
//  4. Отправляем в Telegram (если настроен) или на email
//  5. Возвращаем успешный ответ
func CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest

	// 1. Парсим JSON-body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невалидный JSON",
		})
	}

	// 2. Валидация
	if req.Customer.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "имя обязательно",
		})
	}
	if req.Customer.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email обязателен",
		})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "корзина пуста",
		})
	}
	if req.Total <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "неверная сумма заказа",
		})
	}

	// 3. Формируем сообщение
	message := formatOrderMessage(req)

	// 4. Отправляем уведомление
	// Сначала пробуем Telegram, потом email
	sent := false
	if telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN"); telegramBotToken != "" {
		if chatID := os.Getenv("TELEGRAM_CHAT_ID"); chatID != "" {
			if err := sendTelegramMessage(telegramBotToken, chatID, message); err == nil {
				sent = true
			}
		}
	}

	// Если Telegram не сработал, пробуем email
	if !sent {
		if emailTo := os.Getenv("ORDER_EMAIL"); emailTo != "" {
			// Здесь можно добавить отправку через SMTP или сервис типа SendGrid
			// Пока просто логируем
			fmt.Printf("Order email (not implemented): %s\n", emailTo)
		}
	}

	// 5. Возвращаем успешный ответ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Заказ успешно оформлен. Мы свяжемся с вами в ближайшее время.",
	})
}

// formatOrderMessage форматирует сообщение о заказе
func formatOrderMessage(req CreateOrderRequest) string {
	var b strings.Builder

	b.WriteString("🛒 *НОВЫЙ ЗАКАЗ*\n\n")
	b.WriteString(fmt.Sprintf("👤 *Клиент:* %s\n", req.Customer.Name))
	b.WriteString(fmt.Sprintf("📧 *Email:* %s\n", req.Customer.Email))

	if req.Customer.Phone != "" {
		b.WriteString(fmt.Sprintf("📱 *Телефон:* %s\n", req.Customer.Phone))
	}
	if req.Customer.Telegram != "" {
		b.WriteString(fmt.Sprintf("💬 *Telegram:* %s\n", req.Customer.Telegram))
	}
	if req.Customer.Address != "" {
		b.WriteString(fmt.Sprintf("📍 *Адрес:* %s\n", req.Customer.Address))
	}

	b.WriteString("\n📦 *Товары:*\n")
	total := int64(0)
	for i, item := range req.Items {
		itemTotal := item.Price * int64(item.Quantity)
		total += itemTotal
		b.WriteString(fmt.Sprintf("%d. Товар ID: %s\n", i+1, item.ProductID))
		b.WriteString(fmt.Sprintf("   Количество: %d\n", item.Quantity))
		b.WriteString(fmt.Sprintf("   Цена: %d руб.\n", item.Price/100))
		b.WriteString(fmt.Sprintf("   Сумма: %d руб.\n\n", itemTotal/100))
	}

	b.WriteString(fmt.Sprintf("💰 *Итого:* %d руб.\n", req.Total/100))

	if req.Comment != "" {
		b.WriteString(fmt.Sprintf("\n💬 *Комментарий:*\n%s\n", req.Comment))
	}

	return b.String()
}

// sendTelegramMessage отправляет сообщение в Telegram через Bot API
func sendTelegramMessage(botToken, chatID, message string) error {
	// Экранируем специальные символы для Markdown
	message = strings.ReplaceAll(message, "_", "\\_")
	message = strings.ReplaceAll(message, "*", "\\*")
	message = strings.ReplaceAll(message, "[", "\\[")
	message = strings.ReplaceAll(message, "]", "\\]")

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":        message,
		"parse_mode":  "Markdown",
		"disable_web_page_preview": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram marshal: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: %s", string(body))
	}

	return nil
}
