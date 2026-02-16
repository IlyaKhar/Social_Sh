package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UploadDir — директория для хранения загруженных файлов
const UploadDir = "./uploads"

// MaxFileSize — максимальный размер файла (10MB)
const MaxFileSize = 10 * 1024 * 1024

// AllowedImageTypes — разрешённые типы изображений
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// init создаёт директорию для загрузок если её нет
func init() {
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		fmt.Printf("WARN: не удалось создать директорию uploads: %v\n", err)
	}

	// Создаём поддиректорию для товаров
	if err := os.MkdirAll(filepath.Join(UploadDir, "products"), 0755); err != nil {
		fmt.Printf("WARN: не удалось создать директорию uploads/products: %v\n", err)
	}

	// Создаём поддиректорию для галереи
	if err := os.MkdirAll(filepath.Join(UploadDir, "gallery"), 0755); err != nil {
		fmt.Printf("WARN: не удалось создать директорию uploads/gallery: %v\n", err)
	}
}

// UploadProductImage — загрузка изображения товара.
// POST /api/admin/upload/product
// FormData: file (изображение)
// Ответ: { "url": "/uploads/products/xxx.jpg" }
func UploadProductImage(c *fiber.Ctx) error {
	return uploadImage(c, "products")
}

// UploadGalleryImage — загрузка изображения для галереи.
// POST /api/admin/upload/gallery
// FormData: file (изображение)
// Ответ: { "url": "/uploads/gallery/xxx.jpg" }
func UploadGalleryImage(c *fiber.Ctx) error {
	return uploadImage(c, "gallery")
}

// uploadImage — общая функция загрузки изображения
func uploadImage(c *fiber.Ctx, subdir string) error {
	// Получаем файл из формы
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "файл не найден в запросе",
		})
	}

	// Проверяем размер файла
	if file.Size > MaxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("файл слишком большой (максимум %d MB)", MaxFileSize/(1024*1024)),
		})
	}

	// Открываем файл для чтения
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось открыть файл",
		})
	}
	defer src.Close()

	// Читаем первые 512 байт для определения типа файла
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось прочитать файл",
		})
	}

	// Определяем MIME тип по первым байтам
	contentType := detectImageType(buffer)
	if contentType == "" || !AllowedImageTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "разрешены только изображения (JPEG, PNG, WebP, GIF)",
		})
	}

	// Возвращаемся в начало файла
	src.Seek(0, 0)

	// Генерируем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		// Определяем расширение по MIME типу
		switch contentType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".jpg"
		}
	}

	// Генерируем имя файла: timestamp_random.ext
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), generateRandomString(8), ext)
	filePath := filepath.Join(UploadDir, subdir, filename)

	// Создаём файл на диске
	dst, err := os.Create(filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось сохранить файл",
		})
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, src); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось сохранить файл",
		})
	}

	// Возвращаем URL файла (относительный путь для фронтенда)
	url := fmt.Sprintf("/uploads/%s/%s", subdir, filename)
	return c.JSON(fiber.Map{"url": url})
}

// detectImageType определяет тип изображения по первым байтам
func detectImageType(buffer []byte) string {
	if len(buffer) < 4 {
		return ""
	}

	// JPEG: FF D8 FF
	if buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG: 89 50 4E 47
	if buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "image/png"
	}

	// GIF: 47 49 46 38
	if buffer[0] == 0x47 && buffer[1] == 0x49 && buffer[2] == 0x46 && buffer[3] == 0x38 {
		return "image/gif"
	}

	// WebP: RIFF...WEBP
	if len(buffer) >= 12 &&
		buffer[0] == 0x52 && buffer[1] == 0x49 && buffer[2] == 0x46 && buffer[3] == 0x46 &&
		buffer[8] == 0x57 && buffer[9] == 0x45 && buffer[10] == 0x42 && buffer[11] == 0x50 {
		return "image/webp"
	}

	return ""
}

// generateRandomString генерирует случайную строку для имени файла
func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback на timestamp если rand не работает
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
