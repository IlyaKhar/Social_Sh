package repository

import (
	"database/sql"
	"fmt"
	"socialsh/backend/internal/models"
)

// GallerySQLRepo — реализация GalleryRepository поверх PostgreSQL.
// Структура 1-в-1 как ProductSQLRepo, только таблица gallery_items и поля другие.
type GallerySQLRepo struct {
	db *sql.DB
}

func NewGallerySQLRepo(db *sql.DB) *GallerySQLRepo {
	return &GallerySQLRepo{db: db}
}

// ────────────────────────────────────────────────
// Публичный метод
// ────────────────────────────────────────────────

// ListByCategory — получить элементы галереи по категории.
//
// Как читается:
//  1. SELECT из gallery_items WHERE category = $1 ORDER BY sort_order ASC.
//     sort_order — порядок отображения (чтобы админ мог расставлять фотки руками).
//  2. Итерируем rows → Scan в models.GalleryItem.
//
// Отличия от products.List:
//   - Нет пагинации (галерея обычно маленькая, грузим всё разом).
//   - Нет jsonb полей — все колонки простые (string, int).
//   - Фильтр один: category, а не два булевых флага.
//
// TODO: реализуй по аналогии с products.List, но проще:
//
//	query := `SELECT id, category, title, image, sort_order
//	           FROM gallery_items WHERE category = $1 ORDER BY sort_order ASC`
//	rows, err := r.db.Query(query, category)
//	defer rows.Close()
//	for rows.Next() { rows.Scan(&item.ID, &item.Category, &item.Title, &item.Image, &item.Order) }
//	rows.Err()
func (r *GallerySQLRepo) ListByCategory(category string) ([]models.GalleryItem, error) {
	query := `SELECT id, category, title, image, sort_order
	           FROM gallery_items WHERE category = $1 ORDER BY sort_order ASC`

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("gallery.ListByCategory query: %w", err)
	}
	defer rows.Close()

	var items []models.GalleryItem
	for rows.Next() {
		var item models.GalleryItem
		err := rows.Scan(&item.ID, &item.Category, &item.Title, &item.Image, &item.Order)
		if err != nil {
			return nil, fmt.Errorf("gallery.ListByCategory scan: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gallery.ListByCategory rows: %w", err)
	}

	return items, nil
}

// ────────────────────────────────────────────────
// Админские методы
// ────────────────────────────────────────────────

// ListAll — все элементы галереи (для админ-панели).
//
// Как читается:
//
//	Тот же SELECT, что ListByCategory, но без WHERE category.
//	ORDER BY sort_order ASC — чтобы видеть общий порядок.
//
// TODO: аналогично products.ListAll, но колонки другие:
//
//	query := `SELECT id, category, title, image, sort_order
//	           FROM gallery_items ORDER BY sort_order ASC`
func (r *GallerySQLRepo) ListAll() ([]models.GalleryItem, error) {
	query := `SELECT id, category, title, image, sort_order
	           FROM gallery_items ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("gallery.ListAll query: %w", err)
	}
	defer rows.Close()

	var items []models.GalleryItem
	for rows.Next() {
		var item models.GalleryItem
		err := rows.Scan(&item.ID, &item.Category, &item.Title, &item.Image, &item.Order)
		if err != nil {
			return nil, fmt.Errorf("gallery.ListAll scan: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gallery.ListAll rows: %w", err)
	}

	return items, nil
}

// Create — добавить элемент галереи.
//
// Как читается:
//
//	INSERT INTO gallery_items (...) VALUES (...) RETURNING id
//	→ Scan(&item.ID) — подхватываем сгенерированный UUID.
//
// Отличия от products.Create:
//   - Нет jsonb (images), все поля простые → не нужен json.Marshal.
//   - sort_order — числовое поле, не забудь его в INSERT.
//
// TODO:
//
//	query := `INSERT INTO gallery_items (category, title, image, sort_order)
//	           VALUES ($1, $2, $3, $4) RETURNING id`
//	r.db.QueryRow(query, item.Category, item.Title, item.Image, item.Order).Scan(&item.ID)
func (r *GallerySQLRepo) Create(item *models.GalleryItem) error {
	query := `INSERT INTO gallery_items (category, title, image, sort_order)
	           VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(query, item.Category, item.Title, item.Image, item.Order).Scan(&item.ID)
	if err != nil {
		return fmt.Errorf("gallery.Create: %w", err)
	}
	return nil
}

// Update — обновить элемент галереи по id.
//
// Как читается:
//
//	UPDATE gallery_items SET ... WHERE id = $5 RETURNING ...
//	→ Scan в новый GalleryItem и возвращаем.
//
// Отличия от products.Update:
//   - Меньше полей (нет price, currency, images, is_new, is_on_sale).
//   - Нет jsonb → не нужен Marshal/Unmarshal.
//
// TODO:
//
//	query := `UPDATE gallery_items
//	           SET category = $1, title = $2, image = $3, sort_order = $4
//	           WHERE id = $5
//	           RETURNING id, category, title, image, sort_order`
//	r.db.QueryRow(query, ...).Scan(...)
func (r *GallerySQLRepo) Update(id string, item *models.GalleryItem) (*models.GalleryItem, error) {
	query := `UPDATE gallery_items
	           SET category = $1, title = $2, image = $3, sort_order = $4
	           WHERE id = $5
	           RETURNING id, category, title, image, sort_order`

	var updated models.GalleryItem
	err := r.db.QueryRow(query, item.Category, item.Title, item.Image, item.Order, id).Scan(
		&updated.ID, &updated.Category, &updated.Title, &updated.Image, &updated.Order,
	)
	if err != nil {
		return nil, fmt.Errorf("gallery.Update: %w", err)
	}

	return &updated, nil
}

// Delete — удалить элемент галереи по id.
//
// Как читается:
//
//	DELETE FROM gallery_items WHERE id = $1
//	→ Exec + RowsAffected проверка (ровно как в products.Delete).
//
// TODO: скопируй products.Delete и замени таблицу на gallery_items.
func (r *GallerySQLRepo) Delete(id string) error {
	query := `DELETE FROM gallery_items WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("gallery.Delete: %w", err)
	}

	// Проверяем, что реально удалили строку
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gallery.Delete rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("gallery.Delete: элемент галереи с id=%s не найден", id)
	}

	return nil
}
