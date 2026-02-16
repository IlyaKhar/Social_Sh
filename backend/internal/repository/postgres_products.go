package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"socialsh/backend/internal/models"
)

// ProductSQLRepo — реализация ProductRepository поверх PostgreSQL.
// Хранит в себе указатель на пул соединений (*sql.DB).
type ProductSQLRepo struct {
	db *sql.DB
}

// NewProductSQLRepo — конструктор: принимает пул и возвращает готовый репо.
func NewProductSQLRepo(db *sql.DB) *ProductSQLRepo {
	return &ProductSQLRepo{db: db}
}

// ────────────────────────────────────────────────
// Публичные методы (для витрины)
// ────────────────────────────────────────────────

// List — получить список товаров с фильтрами и пагинацией.
//
// Как читается:
//  1. Стартуем с базового SELECT.
//  2. Если newOnly=true → добавляем WHERE is_new = true.
//     Если saleOnly=true → добавляем WHERE is_on_sale = true.
//     Оба флага одновременно не предполагаются, но WHERE корректно
//     склеится через AND, если вдруг оба true.
//  3. Добавляем ORDER BY + LIMIT/OFFSET для пагинации.
//  4. Итерируем rows, сканируем каждую строку в models.Product.
//     Поле images в Postgres хранится как jsonb, поэтому считываем
//     его как []byte и десериализуем через json.Unmarshal.
//  5. После цикла проверяем rows.Err() — там могут быть ошибки,
//     которые не всплывают в rows.Next().
func (r *ProductSQLRepo) List(newOnly, saleOnly bool, page, limit int) ([]models.Product, error) {
	// Базовый запрос
	query := `SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale
	           FROM products WHERE 1=1`
	// args — слайс для параметризованных значений ($1, $2, ...)
	args := []interface{}{}
	argIdx := 1 // счётчик для $N плейсхолдеров

	// Динамические фильтры
	if newOnly {
		query += fmt.Sprintf(" AND is_new = $%d", argIdx)
		args = append(args, true)
		argIdx++
	}
	if saleOnly {
		query += fmt.Sprintf(" AND is_on_sale = $%d", argIdx)
		args = append(args, true)
		argIdx++
	}

	// Сортировка + пагинация
	// OFFSET = (page - 1) * limit → пропускаем уже просмотренные
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	// Выполняем запрос
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("products.List query: %w", err)
	}
	defer rows.Close() // ОБЯЗАТЕЛЬНО закрыть, иначе утечка соединений

	// Собираем результат
	var products []models.Product
	for rows.Next() {
		var p models.Product
		var imagesJSON []byte // images хранится как jsonb → читаем в сырые байты

		err := rows.Scan(
			&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.Price, &p.Currency, &imagesJSON,
			&p.IsNew, &p.IsOnSale,
		)
		if err != nil {
			return nil, fmt.Errorf("products.List scan: %w", err)
		}

		// Десериализуем jsonb → []string
		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
				return nil, fmt.Errorf("products.List unmarshal images: %w", err)
			}
		}

		products = append(products, p)
	}

	// Проверяем ошибки, которые могли возникнуть во время итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("products.List rows: %w", err)
	}

	return products, nil
}

// GetBySlug — получить один товар по его slug (URL-дружественный идентификатор).
//
// Как читается:
//  1. Делаем SELECT одной строки по slug.
//  2. QueryRow → Scan. Если строка не найдена, sql вернёт sql.ErrNoRows.
//  3. Оборачиваем ошибку, чтобы хендлер мог отличить «не найдено» от «БД упала».
func (r *ProductSQLRepo) GetBySlug(slug string) (*models.Product, error) {
	query := `SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale
	           FROM products WHERE slug = $1 LIMIT 1`

	var p models.Product
	var imagesJSON []byte

	err := r.db.QueryRow(query, slug).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Description,
		&p.Price, &p.Currency, &imagesJSON,
		&p.IsNew, &p.IsOnSale,
	)
	if err != nil {
		// sql.ErrNoRows — товар не найден, это не серверная ошибка
		return nil, fmt.Errorf("products.GetBySlug: %w", err)
	}

	if imagesJSON != nil {
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, fmt.Errorf("products.GetBySlug unmarshal images: %w", err)
		}
	}

	return &p, nil
}

// ────────────────────────────────────────────────
// Админские методы (CRUD)
// ────────────────────────────────────────────────

// ListAll — получить ВСЕ товары без фильтров (для админ-панели).
//
// Как читается:
//
//	Тот же SELECT, что и List, но без WHERE-фильтров и пагинации.
//	Админу нужно видеть всё.
func (r *ProductSQLRepo) ListAll() ([]models.Product, error) {
	query := `SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale
	           FROM products ORDER BY id DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("products.ListAll query: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var imagesJSON []byte

		err := rows.Scan(
			&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.Price, &p.Currency, &imagesJSON,
			&p.IsNew, &p.IsOnSale,
		)
		if err != nil {
			return nil, fmt.Errorf("products.ListAll scan: %w", err)
		}

		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
				return nil, fmt.Errorf("products.ListAll unmarshal images: %w", err)
			}
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("products.ListAll rows: %w", err)
	}

	return products, nil
}

// GetByID — получить товар по UUID (для админки, где оперируем id, а не slug).
//
// Как читается:
//
//	Аналогично GetBySlug, но ищем по id.
func (r *ProductSQLRepo) GetByID(id string) (*models.Product, error) {
	query := `SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale
	           FROM products WHERE id = $1 LIMIT 1`

	var p models.Product
	var imagesJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Description,
		&p.Price, &p.Currency, &imagesJSON,
		&p.IsNew, &p.IsOnSale,
	)
	if err != nil {
		return nil, fmt.Errorf("products.GetByID: %w", err)
	}

	if imagesJSON != nil {
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, fmt.Errorf("products.GetByID unmarshal images: %w", err)
		}
	}

	return &p, nil
}

// Create — вставить новый товар в БД.
//
// Как читается:
//  1. Сериализуем слайс Images в JSON (потому что в Postgres это jsonb).
//  2. INSERT ... RETURNING id — Postgres сам генерит UUID,
//     а мы его сразу подхватываем в product.ID через Scan.
//  3. pq.Array — НЕ используется здесь, т.к. images — это jsonb, а не text[].
//     Но если когда-нибудь перейдёшь на text[] — вот тебе импорт pq уже готов.
func (r *ProductSQLRepo) Create(product *models.Product) error {
	// Сериализуем images в JSON
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return fmt.Errorf("products.Create marshal images: %w", err)
	}

	query := `INSERT INTO products (slug, title, description, price, currency, images, is_new, is_on_sale)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	           RETURNING id`

	// Scan сразу пишет сгенерированный id в структуру
	err = r.db.QueryRow(query,
		product.Slug, product.Title, product.Description,
		product.Price, product.Currency, imagesJSON,
		product.IsNew, product.IsOnSale,
	).Scan(&product.ID)
	if err != nil {
		return fmt.Errorf("products.Create: %w", err)
	}

	return nil
}

// Update — обновить существующий товар по id.
//
// Как читается:
//  1. Сериализуем images → JSON (только если images не nil и не пустой).
//  2. UPDATE ... WHERE id = $9 RETURNING ... — обновляем строку и сразу
//     получаем обновлённые данные обратно (не делаем второй SELECT).
//  3. Scan в новый Product и возвращаем указатель.
//  4. Если строка не найдена (id не существует), вернётся sql.ErrNoRows.
func (r *ProductSQLRepo) Update(id string, product *models.Product) (*models.Product, error) {
	// Если images не передан (nil) или пустой массив, не обновляем его
	// Сначала получаем текущий товар, чтобы сохранить существующие images
	current, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("products.Update get current: %w", err)
	}

	// Если images не передан или пустой, используем текущие
	if product.Images == nil || len(product.Images) == 0 {
		product.Images = current.Images
	}

	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return nil, fmt.Errorf("products.Update marshal images: %w", err)
	}

	query := `UPDATE products
	           SET slug = $1, title = $2, description = $3,
	               price = $4, currency = $5, images = $6,
	               is_new = $7, is_on_sale = $8
	           WHERE id = $9
	           RETURNING id, slug, title, description, price, currency, images, is_new, is_on_sale`

	var updated models.Product
	var updatedImagesJSON []byte

	err = r.db.QueryRow(query,
		product.Slug, product.Title, product.Description,
		product.Price, product.Currency, imagesJSON,
		product.IsNew, product.IsOnSale,
		id,
	).Scan(
		&updated.ID, &updated.Slug, &updated.Title, &updated.Description,
		&updated.Price, &updated.Currency, &updatedImagesJSON,
		&updated.IsNew, &updated.IsOnSale,
	)
	if err != nil {
		return nil, fmt.Errorf("products.Update: %w", err)
	}

	if updatedImagesJSON != nil {
		if err := json.Unmarshal(updatedImagesJSON, &updated.Images); err != nil {
			return nil, fmt.Errorf("products.Update unmarshal images: %w", err)
		}
	}

	return &updated, nil
}

// Delete — удалить товар по id.
//
// Как читается:
//  1. Exec (не Query, потому что нам не нужны строки обратно).
//  2. RowsAffected — проверяем, удалилось ли что-то.
//     Если 0 — значит id не существует → возвращаем ошибку.
func (r *ProductSQLRepo) Delete(id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("products.Delete: %w", err)
	}

	// Проверяем, что реально удалили строку
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("products.Delete rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("products.Delete: товар с id=%s не найден", id)
	}

	return nil
}

// ────────────────────────────────────────────────
// Хелпер: scanProduct — вынести повторяющийся Scan в одно место.
// Пока не используется (всё inline), но если задолбает копипаста —
// раскомментируй и замени все Scan-блоки вызовом scanProduct.
// ────────────────────────────────────────────────

// func scanProduct(scanner interface{ Scan(dest ...any) error }) (*models.Product, error) {
// 	var p models.Product
// 	var imagesJSON []byte
// 	err := scanner.Scan(
// 		&p.ID, &p.Slug, &p.Title, &p.Description,
// 		&p.Price, &p.Currency, &imagesJSON,
// 		&p.IsNew, &p.IsOnSale,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if imagesJSON != nil {
// 		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return &p, nil
// }

// Search — поиск товаров по названию (без учёта регистра, частичное совпадение).
//
// Как читается:
//  1. Используем ILIKE для поиска без учёта регистра в PostgreSQL.
//  2. Ищем в title и description: WHERE title ILIKE '%query%' OR description ILIKE '%query%'
//  3. Добавляем пагинацию (LIMIT/OFFSET).
//  4. Сканируем результаты аналогично List().
func (r *ProductSQLRepo) Search(query string, page, limit int) ([]models.Product, error) {
	// Экранируем спецсимволы для ILIKE (простая защита)
	searchTerm := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"
	offset := (page - 1) * limit

	querySQL := `SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale
	              FROM products
	              WHERE title ILIKE $1 OR description ILIKE $1
	              ORDER BY 
	                CASE 
	                  WHEN title ILIKE $2 THEN 1
	                  WHEN description ILIKE $2 THEN 2
	                  ELSE 3
	                END,
	                id DESC
	              LIMIT $3 OFFSET $4`

	// $2 — точное совпадение в начале для приоритета
	exactStart := strings.ReplaceAll(query, "%", "\\%") + "%"

	rows, err := r.db.Query(querySQL, searchTerm, exactStart, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("products.Search query: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var imagesJSON []byte

		err := rows.Scan(
			&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.Price, &p.Currency, &imagesJSON,
			&p.IsNew, &p.IsOnSale,
		)
		if err != nil {
			return nil, fmt.Errorf("products.Search scan: %w", err)
		}

		if imagesJSON != nil {
			if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
				return nil, fmt.Errorf("products.Search unmarshal images: %w", err)
			}
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("products.Search rows: %w", err)
	}

	return products, nil
}

// NOTE: когда понадобится pq.Array для text[] колонок —
// сделай go get github.com/lib/pq и добавь импорт.
