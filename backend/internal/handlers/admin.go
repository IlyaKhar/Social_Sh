package handlers

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/models"
	"socialsh/backend/internal/utils"
)

// ═══════════════════════════════════════════════════════════════
// Админские хендлеры для CRUD операций.
// Все роуты защищены: Protected(jwtSecret) → AdminOnly().
// Если юзер не админ — 403 на уровне middleware, сюда не дойдёт.
// ═══════════════════════════════════════════════════════════════

// ──── Товары ────

// AdminListProducts — возвращает ВСЕ товары без фильтрации.
// GET /api/admin/products
// В отличие от публичного GetProducts (который фильтрует по new/sale и пагинирует),
// тут отдаём полный список — включая скрытые, черновики, без пагинации.
// Ответ: { "items": [ { "id", "slug", "title", "price", ... }, ... ] }
func AdminListProducts(c *fiber.Ctx) error {
	// TODO: Repo.Products.ListAll() — без фильтров new/sale, показать всё
	items, err := Repo.Products.ListAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ //500
			"error": "не удалось получить список товаров",
		})
	}

	return c.JSON(fiber.Map{"items": items})
}

// AdminCreateProduct — создание нового товара.
// POST /api/admin/products
// Body: { "slug": "hoodie-black", "title": "Худи чёрное", "price": 4990, ... }
// Ответ 201: { "item": { созданный товар с ID } }
//
// Как это читается:
//
//	Админ присылает JSON с данными нового товара.
//	Мы парсим body в models.Product через c.BodyParser.
//	Валидируем — slug и title обязательны, price > 0.
//	Вызываем Repo.Products.Create(&product) — он вставит строку в БД и заполнит product.ID.
//	Если Create вернул ошибку — 500.
//	Если всё ок — возвращаем 201 и созданный товар.
func AdminCreateProduct(c *fiber.Ctx) error {
	var product models.Product

	// TODO: парсим body — c.BodyParser читает JSON и заполняет структуру по тегам json:"..."
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "невалидный JSON",
		})
	}

	// TODO: валидация полей
	// Подумай что обязательно: slug (уникальный), title (не пустой), price (> 0).
	// Если slug пустой — вернуть 400 с понятным сообщением.
	// Если price <= 0 — тоже 400.
	if product.Slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "slug обязателен",
		})
	}
	if product.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "title обязателен",
		})
	}
	if product.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "price должен быть больше 0",
		})
	}

	// TODO: вызвать Repo.Products.Create(&product)
	// Create должен:
	//   - INSERT INTO products (slug, title, ...) VALUES ($1, $2, ...) RETURNING id
	//   - Записать сгенерированный id в product.ID
	//   - Вернуть nil если ок, error если дубликат slug или ошибка БД
	if err := Repo.Products.Create(&product); err != nil {
		// Проверяем на duplicate key (slug уже занят) → 409 Conflict
		if utils.IsDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": utils.FormatDuplicateError(err),
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось создать товар",
		})
	}

	// TODO: вернуть c.Status(201).JSON(fiber.Map{"item": product})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"item": product})
}

// AdminGetProduct — один товар по ID для формы редактирования.
// GET /api/admin/products/:id
// Ответ: { "item": { ... } }
//
// Как это читается:
//
//	Достаём id из URL через c.Params("id").
//	Вызываем Repo.Products.GetByID(id).
//	GetByID — это SELECT * FROM products WHERE id = $1.
//	Если вернулся nil — товар не найден, отдаём 404.
//	Если ошибка — 500.
//	Если нашёлся — отдаём { "item": product }.
func AdminGetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id обязателен",
		})
	}

	product, err := Repo.Products.GetByID(id)
	if err != nil {
		// Если товар не найден (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "товар не найден",
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при получении товара",
		})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "товар не найден",
		})
	}

	return c.JSON(fiber.Map{"item": product})
}

// AdminUpdateProduct — частичное обновление товара.
// PATCH /api/admin/products/:id
// Body: { "title": "Новое название", "price": 5990 } — только изменённые поля.
// Ответ: { "item": { обновлённый товар } }
//
// Как это читается:
//
//	Достаём id из URL: c.Params("id").
//	Парсим body в models.Product через c.BodyParser.
//	Вызываем Repo.Products.Update(id, &product).
//	Update должен:
//	  - Собрать SQL с SET только тех полей, которые не нулевые (или использовать COALESCE).
//	  - Выполнить UPDATE products SET title=$1, price=$2 WHERE id=$3 RETURNING *.
//	  - Вернуть обновлённый товар.
//	Если товар не найден (0 rows affected) — 404.
//	Если ошибка БД — 500.
//	Если всё ок — вернуть обновлённый товар.
//
// Подсказка: посмотри как сделан UpdateProfile в account.go — похожая логика.
func AdminUpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id обязателен",
		})
	}

	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невалидный JSON",
		})
	}

	updated, err := Repo.Products.Update(id, &product)
	if err != nil {
		// Если товар не найден (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "товар не найден",
			})
		}
		// Проверяем на duplicate key (slug уже занят другим товаром)
		if utils.IsDuplicateKeyError(err) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": utils.FormatDuplicateError(err),
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось обновить товар",
		})
	}
	if updated == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "товар не найден",
		})
	}

	return c.JSON(fiber.Map{"item": updated})
}

// AdminDeleteProduct — удаление товара по ID.
// DELETE /api/admin/products/:id
// Ответ: { "message": "ok" }
//
// Как это читается:
//
//	Достаём id: c.Params("id").
//	Вызываем Repo.Products.Delete(id).
//	Delete — это DELETE FROM products WHERE id = $1.
//	Проверяем RowsAffected: если 0 — товар не найден, 404.
//	Если ошибка — 500.
//	Если удалился — вернуть { "message": "ok" }.
func AdminDeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id обязателен",
		})
	}

	if err := Repo.Products.Delete(id); err != nil {
		// Проверяем на "не найден" (репозиторий возвращает ошибку если affected == 0)
		if strings.Contains(err.Error(), "не найден") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "товар не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось удалить товар",
		})
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

// ──── Галерея ────

// AdminListGalleryItems — все элементы галереи.
// GET /api/admin/gallery
// Ответ: { "items": [ ... ] }
//
// Как это читается:
//
//	Вызываем Repo.Gallery.ListAll() — SELECT * FROM gallery_items ORDER BY sort_order.
//	Если ошибка — 500.
//	Отдаём { "items": items }.
//	Один в один как AdminListProducts, только другой репозиторий.
func AdminListGalleryItems(c *fiber.Ctx) error {
	// TODO: реализуй — скопируй логику из AdminListProducts, замени Products на Gallery
	items, err := Repo.Gallery.ListAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ //500
			"error": "не удалось получить элементы галереи",
		})
	}

	return c.JSON(fiber.Map{"items": items})
}

// AdminCreateGalleryItem — добавить фото в галерею.
// POST /api/admin/gallery
// Body: { "category": "intro", "title": "Фото 1", "image": "/uploads/photo.jpg", "order": 1 }
// Ответ 201: { "item": { ... } }
//
// Как это читается:
//
//	Парсим body в models.GalleryItem.
//	Валидируем: category и image обязательны.
//	Repo.Gallery.Create(&item) — INSERT INTO gallery_items (...) RETURNING id.
//	Ошибка → 500, ок → 201 + созданный элемент.
//	Та же структура что AdminCreateProduct.
func AdminCreateGalleryItem(c *fiber.Ctx) error {
	// TODO: реализуй по описанию выше
	var item models.GalleryItem

	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "невалидный JSON",
		})
	}

	if item.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "category обязательна",
		})
	}
	if item.Image == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{ //400
			"error": "image обязателен",
		})
	}

	if err := Repo.Gallery.Create(&item); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ //500
			"error": "не удалось добавить элемент галереи",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"item": item}) //201
}

// AdminUpdateGalleryItem — обновить элемент галереи.
// PATCH /api/admin/gallery/:id
// Body: { "title": "Новое название", "order": 5 }
// Ответ: { "item": { ... } }
//
// Как это читается:
//
//	id из URL, body парсим в models.GalleryItem.
//	Repo.Gallery.Update(id, &item) — UPDATE gallery_items SET ... WHERE id = $1 RETURNING *.
//	Не нашёлся → 404, ошибка → 500, ок → обновлённый элемент.
func AdminUpdateGalleryItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id обязателен",
		})
	}

	var item models.GalleryItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невалидный JSON",
		})
	}

	updated, err := Repo.Gallery.Update(id, &item)
	if err != nil {
		// Если элемент не найден (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "элемент галереи не найден",
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось обновить элемент галереи",
		})
	}
	if updated == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "элемент галереи не найден",
		})
	}

	return c.JSON(fiber.Map{"item": updated})
}

// AdminDeleteGalleryItem — удалить элемент галереи.
// DELETE /api/admin/gallery/:id
// Ответ: { "message": "ok" }
//
// Как это читается:
//
//	id из URL → Repo.Gallery.Delete(id) → DELETE FROM gallery_items WHERE id = $1.
//	RowsAffected == 0 → 404, ошибка → 500, ок → { "message": "ok" }.
//	Один в один как AdminDeleteProduct.
func AdminDeleteGalleryItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id обязателен",
		})
	}

	if err := Repo.Gallery.Delete(id); err != nil {
		// Проверяем на "не найден"
		if strings.Contains(err.Error(), "не найден") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "элемент галереи не найден",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось удалить элемент галереи",
		})
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

// ──── Инфо-страницы ────

// AdminListPages — список всех статических страниц.
// GET /api/admin/pages
// Ответ: { "items": [ { "slug": "payment", "title": "Оплата", "content": "..." }, ... ] }
//
// Как это читается:
//
//	Repo.Pages.ListAll() — SELECT * FROM pages ORDER BY slug.
//	Ошибка → 500, ок → { "items": pages }.
func AdminListPages(c *fiber.Ctx) error {
	// TODO: реализуй — аналогично AdminListProducts/AdminListGalleryItems
	items, err := Repo.Pages.ListAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось получить список страниц",
		})
	}

	return c.JSON(fiber.Map{"items": items})
}

// AdminUpdatePage — обновить контент страницы по slug.
// PATCH /api/admin/pages/:slug
// Body: { "title": "Новый заголовок", "content": "<p>Новый текст</p>" }
// Ответ: { "item": { обновлённая страница } }
//
// Как это читается:
//
//	slug из URL через c.Params("slug") — НЕ id, а slug (payment, delivery, returns, contacts).
//	Парсим body в models.Page.
//	Repo.Pages.Update(slug, &page) — UPDATE pages SET title=$1, content=$2 WHERE slug=$3 RETURNING *.
//	Не нашлась → 404, ошибка → 500, ок → обновлённая страница.
//
// Отличие от товаров: тут slug вместо id, потому что у страниц slug — это первичный ключ.
func AdminUpdatePage(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "slug обязателен",
		})
	}

	var page models.Page
	if err := c.BodyParser(&page); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невалидный JSON",
		})
	}

	updated, err := Repo.Pages.Update(slug, &page)
	if err != nil {
		// Если страница не найдена (sql.ErrNoRows) — возвращаем 404
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "страница не найдена",
			})
		}
		// Иначе — серверная ошибка
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "не удалось обновить страницу",
		})
	}
	if updated == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "страница не найдена",
		})
	}

	return c.JSON(fiber.Map{"item": updated})
}
