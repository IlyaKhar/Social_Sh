package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// IsDuplicateKeyError проверяет, является ли ошибка нарушением уникального ключа.
// PostgreSQL возвращает код ошибки 23505 для duplicate key violations.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	// Проверяем на pq.Error (PostgreSQL драйвер)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" // unique_violation
	}

	// Проверяем строковое представление ошибки (fallback)
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "23505")
}

// GetDuplicateKeyField пытается извлечь название поля из ошибки дубликата.
// Например: "duplicate key value violates unique constraint \"products_slug_key\""
// вернёт "slug".
func GetDuplicateKeyField(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	
	// Пытаемся найти поле в сообщении об ошибке
	if strings.Contains(errStr, "slug") {
		return "slug"
	}
	if strings.Contains(errStr, "email") {
		return "email"
	}

	return ""
}

// FormatDuplicateError форматирует ошибку дубликата в понятное сообщение.
func FormatDuplicateError(err error) string {
	field := GetDuplicateKeyField(err)
	if field != "" {
		return fmt.Sprintf("%s уже используется", field)
	}
	return "значение уже существует"
}
