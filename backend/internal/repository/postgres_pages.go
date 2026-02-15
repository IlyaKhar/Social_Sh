package repository

import (
	"database/sql"
	"socialsh/backend/internal/models"
)

// PageSQLRepo — реализация PageRepository поверх PostgreSQL.
//
// ВАЖНОЕ ОТЛИЧИЕ ОТ PRODUCTS И GALLERY:
//   - Нет Create / Delete. Страницы (оплата, доставка, возврат, контакты) —
//     статические сущности, их набор фиксирован. Админ может только
//     РЕДАКТИРОВАТЬ содержимое, но не создавать/удалять страницы.
//   - Первичный ключ — slug (не id). Это строка вроде "payment", "delivery".
//     В WHERE ищем по slug, а не по UUID.
type PageSQLRepo struct {
	db *sql.DB
}

func NewPageSQLRepo(db *sql.DB) *PageSQLRepo {
	return &PageSQLRepo{db: db}
}

// ────────────────────────────────────────────────
// Публичный метод
// ────────────────────────────────────────────────

// GetBySlug — получить страницу по slug.
//
// Как читается:
//
//	SELECT slug, title, content FROM pages WHERE slug = $1 LIMIT 1
//	→ QueryRow → Scan в models.Page.
//	Если не найдено → sql.ErrNoRows (хендлер вернёт 404).
//
// Отличия от products.GetBySlug:
//   - У Page нет id, нет images, нет price — всего 3 поля (slug, title, content).
//   - Scan проще, нет jsonb десериализации.
//
// TODO:
//
//	query := `SELECT slug, title, content FROM pages WHERE slug = $1 LIMIT 1`
//	var p models.Page
//	err := r.db.QueryRow(query, slug).Scan(&p.Slug, &p.Title, &p.Content)
//	if err != nil { return nil, fmt.Errorf("pages.GetBySlug: %w", err) }
//	return &p, nil
func (r *PageSQLRepo) GetBySlug(slug string) (*models.Page, error) {
	// TODO: реализовать
	return nil, nil
}

// ────────────────────────────────────────────────
// Админские методы
// ────────────────────────────────────────────────

// ListAll — получить все страницы (для админ-панели).
//
// Как читается:
//
//	SELECT slug, title, content FROM pages ORDER BY slug
//	→ rows.Next() → Scan → append.
//	Всё как products.ListAll, но проще (3 поля, нет jsonb).
//
// TODO:
//
//	query := `SELECT slug, title, content FROM pages ORDER BY slug`
//	rows, err := r.db.Query(query)
//	...
func (r *PageSQLRepo) ListAll() ([]models.Page, error) {
	// TODO: реализовать
	return nil, nil
}

// Update — обновить содержимое страницы по slug.
//
// Как читается:
//
//	UPDATE pages SET title = $1, content = $2 WHERE slug = $3
//	RETURNING slug, title, content
//	→ QueryRow → Scan → вернуть обновлённую Page.
//
// ВАЖНО: WHERE по slug, а не по id (как в products/gallery).
//
//	Это единственное место где ключ — строка, а не UUID.
//
// Отличия от products.Update:
//   - Меньше полей (только title + content).
//   - WHERE slug = $3, а не WHERE id = $9.
//   - Нет jsonb Marshal/Unmarshal.
//
// TODO:
//
//	query := `UPDATE pages SET title = $1, content = $2
//	           WHERE slug = $3
//	           RETURNING slug, title, content`
//	var updated models.Page
//	err := r.db.QueryRow(query, page.Title, page.Content, slug).Scan(
//	    &updated.Slug, &updated.Title, &updated.Content,
//	)
//	if err != nil { return nil, fmt.Errorf("pages.Update: %w", err) }
//	return &updated, nil
func (r *PageSQLRepo) Update(slug string, page *models.Page) (*models.Page, error) {
	// TODO: реализовать
	return nil, nil
}
