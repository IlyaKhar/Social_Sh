package repository

import (
	"database/sql"
	"fmt"
	"socialsh/backend/internal/models"
	"strings"
)

// AccountSQLRepo — реализация AccountRepository поверх PostgreSQL.
//
// САМЫЙ ХИТРЫЙ РЕПО из всех, потому что:
//  1. Работает с двумя таблицами: users + orders (+ order_items).
//  2. UpdateUser — частичное обновление (PATCH-семантика) с указателями.
//  3. ListOrdersByUser — нужен JOIN или два запроса (orders + order_items).
type AccountSQLRepo struct {
	db *sql.DB
}

func NewAccountSQLRepo(db *sql.DB) *AccountSQLRepo {
	return &AccountSQLRepo{db: db}
}

// ────────────────────────────────────────────────
// Методы для работы с юзерами
// ────────────────────────────────────────────────

// GetUserByID — найти юзера по UUID.
//
// Как читается:
//
//	SELECT id, email, name, password_hash, role FROM users WHERE id = $1
//	→ QueryRow → Scan в models.User.
//	Если не найдено → sql.ErrNoRows.
//
// Используется в: middleware.Protected (после парсинга JWT, подгружаем юзера),
//
//	handlers.GetAccountMe (отдаём профиль текущего юзера).
//
// TODO:
//
//	query := `SELECT id, email, name, password_hash, role FROM users WHERE id = $1`
//	var u models.User
//	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
//	if err != nil { return nil, fmt.Errorf("account.GetUserByID: %w", err) }
//	return &u, nil
func (r *AccountSQLRepo) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, email, name, password_hash, role 
				FROM users WHERE id = $1`

	var u models.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, fmt.Errorf("account.GetUserByID: %w", err)
	}
	return &u, nil
}

// GetUserByEmail — найти юзера по email (для логина).
//
// Как читается:
//
//	То же самое что GetUserByID, но WHERE email = $1.
//
// Используется в: handlers.SignIn (ищем юзера → сравниваем пароль).
//
//	handlers.SignUp (проверяем что email не занят — если нашли, то 409 Conflict).
//
// TODO:
//
//	query := `SELECT id, email, name, password_hash, role FROM users WHERE email = $1`
//	... (аналогично GetUserByID)
func (r *AccountSQLRepo) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, name, password_hash, role FROM users WHERE email = $1`

	var u models.User

	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, fmt.Errorf("users.GetByEmail: %w", err)
	}
	return &u, nil
}

// CreateUser — зарегистрировать нового юзера.
//
// Как читается:
//
//	INSERT INTO users (email, name, password_hash, role) VALUES (...)
//	RETURNING id
//	→ Scan(&user.ID) — подхватываем сгенерированный UUID.
//
// Отличие от products.Create:
//   - password_hash — важно: мы НЕ храним пароль в открытом виде.
//     Хеширование уже произошло в хендлере (bcrypt), сюда приходит хеш.
//   - role по умолчанию "user" (можно задать DEFAULT в SQL-схеме,
//     но лучше явно передавать).
//
// TODO:
//
//	query := `INSERT INTO users (email, name, password_hash, role)
//	           VALUES ($1, $2, $3, $4) RETURNING id`
//	err := r.db.QueryRow(query, user.Email, user.Name, user.PasswordHash, user.Role).Scan(&user.ID)
//	if err != nil { return fmt.Errorf("account.CreateUser: %w", err) }
//	return nil
func (r *AccountSQLRepo) CreateUser(user *models.User) error {
	query := `INSERT INTO users (email, name, password_hash, role)
	           VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(query, user.Email, user.Name, user.PasswordHash, user.Role).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("account.CreateUser: %w", err)
	}
	return nil
}

// UpdateUser — частичное обновление профиля (PATCH-семантика).
//
// Как читается:
//  1. Принимаем UpdateProfileRequest, где поля — указатели (*string).
//     nil = «не прислали это поле» → не трогаем в БД.
//     non-nil = «прислали» → обновляем.
//  2. Собираем SET-часть запроса динамически:
//     if req.Name != nil { добавляем "name = $N" }
//     if req.Email != nil { добавляем "email = $N" }
//  3. Если ни одного поля не прислали → можно вернуть текущего юзера без UPDATE.
//  4. Выполняем UPDATE ... WHERE id = $X RETURNING ... → Scan.
//
// ВАЖНО: это самый сложный метод из всех репозиториев.
//
//	В products.Update мы перезаписываем ВСЕ поля.
//	Здесь — только те, что пришли (частичный апдейт).
//
// Паттерн динамического SQL:
//
//	setClauses := []string{}
//	args := []interface{}{}
//	argIdx := 1
//
//	if req.Name != nil {
//	    setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
//	    args = append(args, *req.Name)
//	    argIdx++
//	}
//	if req.Email != nil {
//	    setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
//	    args = append(args, *req.Email)
//	    argIdx++
//	}
//
//	if len(setClauses) == 0 {
//	    return r.GetUserByID(id) // ничего не обновляем, просто возвращаем текущего
//	}
//
//	query := fmt.Sprintf(
//	    "UPDATE users SET %s WHERE id = $%d RETURNING id, email, name, password_hash, role",
//	    strings.Join(setClauses, ", "), argIdx,
//	)
//	args = append(args, id)
//
//	var u models.User
//	err := r.db.QueryRow(query, args...).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
//
// Не забудь импортировать "strings" и "fmt".
//
// TODO: реализовать по паттерну выше
func (r *AccountSQLRepo) UpdateUser(id string, req *models.UpdateProfileRequest) (*models.User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}

	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.GetUserByID(id)
	}

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d RETURNING id, email, name, password_hash, role",
		strings.Join(setClauses, ", "), argIdx,
	)
	args = append(args, id)

	var u models.User
	err := r.db.QueryRow(query, args...).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, fmt.Errorf("account.UpdateUser: %w", err)
	}

	return &u, nil
}

// ────────────────────────────────────────────────
// Методы для заказов
// ────────────────────────────────────────────────

// ListOrdersByUser — получить все заказы юзера с позициями.
//
// Как читается:
//
//	Это ДВУХЭТАПНЫЙ запрос:
//
//	Этап 1: Получаем заказы
//	  SELECT id, user_id, status, total, created_at, updated_at
//	  FROM orders WHERE user_id = $1 ORDER BY created_at DESC
//	  → Scan каждую строку в models.Order.
//
//	Этап 2: Для каждого заказа подгружаем позиции (order_items)
//	  SELECT id, order_id, product_id, title, price, quantity
//	  FROM order_items WHERE order_id = $1
//	  → Scan в models.OrderItem, пишем в order.Items.
//
//	АЛЬТЕРНАТИВА (если хочешь оптимальнее):
//	  Можно одним запросом через JOIN собрать всё, а потом в Go-коде
//	  группировать Items по OrderID. Но для старта двухэтапный вариант проще.
//
//	Паттерн:
//	  orders := []models.Order{}
//	  rows, _ := r.db.Query(ordersQuery, userID)
//	  for rows.Next() {
//	      var o models.Order
//	      rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
//	      orders = append(orders, o)
//	  }
//	  rows.Close()
//
//	  for i := range orders {
//	      itemRows, _ := r.db.Query(itemsQuery, orders[i].ID)
//	      for itemRows.Next() {
//	          var item models.OrderItem
//	          itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.Price, &item.Quantity)
//	          orders[i].Items = append(orders[i].Items, item)
//	      }
//	      itemRows.Close()
//	  }
//
//	Не забудь defer rows.Close() и проверки ошибок на каждом этапе.
//
// TODO: реализовать по паттерну выше
func (r *AccountSQLRepo) ListOrdersByUser(id string) ([]models.Order, error) {
	// Этап 1: Получаем все заказы юзера
	ordersQuery := `SELECT id, user_id, status, total, created_at, updated_at
	                 FROM orders WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ordersQuery, id)
	if err != nil {
		return nil, fmt.Errorf("account.ListOrdersByUser query orders: %w", err)
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("account.ListOrdersByUser scan order: %w", err)
		}
		// Инициализируем Items как пустой слайс для каждого заказа
		o.Items = []models.OrderItem{}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("account.ListOrdersByUser rows: %w", err)
	}

	// Этап 2: Для каждого заказа подгружаем позиции (order_items)
	itemsQuery := `SELECT id, order_id, product_id, title, price, quantity
	                FROM order_items WHERE order_id = $1`

	for i := range orders {
		itemRows, err := r.db.Query(itemsQuery, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("account.ListOrdersByUser query items: %w", err)
		}

		for itemRows.Next() {
			var item models.OrderItem
			err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Title, &item.Price, &item.Quantity)
			if err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("account.ListOrdersByUser scan item: %w", err)
			}
			orders[i].Items = append(orders[i].Items, item)
		}

		if err := itemRows.Err(); err != nil {
			itemRows.Close()
			return nil, fmt.Errorf("account.ListOrdersByUser itemRows: %w", err)
		}

		itemRows.Close()
	}

	return orders, nil
}
