package models

import "time"

// Здесь только формы структур, без логики.

// Product описывает товар в магазине.
type Product struct {
	ID          string   `json:"id"          db:"id"`
	Slug        string   `json:"slug"        db:"slug"`
	Title       string   `json:"title"       db:"title"`
	Description string   `json:"description" db:"description"`
	Price       int64    `json:"price"       db:"price"`
	Currency    string   `json:"currency"    db:"currency"`
	Images      []string `json:"images"      db:"images"` // в Postgres — jsonb
	IsNew       bool     `json:"isNew"       db:"is_new"`
	IsOnSale    bool     `json:"isOnSale"    db:"is_on_sale"`
}

// GalleryItem — элемент галереи (фото, кадр и т.п.).
type GalleryItem struct {
	ID       string `json:"id"       db:"id"`
	Category string `json:"category" db:"category"` // intro, tattoo, tokyo, ...
	Title    string `json:"title"    db:"title"`
	Image    string `json:"image"    db:"image"`
	Order    int    `json:"order"    db:"sort_order"`
}

// Page — статическая страница (оплата, доставка, возврат, контакты).
type Page struct {
	Slug    string `json:"slug"    db:"slug"` // payment, delivery, returns, contacts
	Title   string `json:"title"   db:"title"`
	Content string `json:"content" db:"content"` // Markdown/HTML/текст
}

// User — аккаунт пользователя.
type User struct {
	ID           string `json:"id"            db:"id"`
	Email        string `json:"email"         db:"email"`
	Name         string `json:"name"          db:"name"`
	PasswordHash string `json:"-"             db:"password_hash"` // json:"-" — не отдаём наружу
	Role         string `json:"role"          db:"role"`          // "user" или "admin"
}

// Order — заказ пользователя.
// В БД таблица orders хранит мета-инфу (кто, когда, статус, сумма).
// Позиции заказа лежат в отдельной таблице order_items (связь 1:N через order_id).
type Order struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"userId" db:"user_id"`
	Status    string      `json:"status" db:"status"` // pending | paid | shipped | delivered | cancelled
	Total     int64       `json:"total" db:"total"`   // итого в копейках (4990 = 49.90 ₽)
	CreatedAt time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time   `json:"updatedAt" db:"updated_at"`
	Items     []OrderItem `json:"items" db:"-"` // db:"-" — не колонка, подгружаем отдельным запросом
}

// OrderItem — одна позиция в заказе (какой товар, сколько штук, по какой цене).
// Хранится в таблице order_items, связана с orders через order_id.
type OrderItem struct {
	ID        string `json:"id" db:"id"`
	OrderID   string `json:"orderId" db:"order_id"`
	ProductID string `json:"productId" db:"product_id"`
	Title     string `json:"title" db:"title"`       // название товара НА МОМЕНТ покупки (не ссылка)
	Price     int64  `json:"price" db:"price"`       // цена за 1 шт. на момент покупки
	Quantity  int    `json:"quantity" db:"quantity"` // количество
}

// ──── Request/Response DTO для auth ────

// SignUpRequest — тело запроса на регистрацию.
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SignInRequest — тело запроса на логин.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest — тело запроса на обновление токена.
type RefreshRequest struct {
	Refresh string `json:"refresh"`
}

// TokenResponse — ответ с JWT-токенами.
type TokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// UpdateProfileRequest — тело запроса на обновление профиля.
// Указатели нужны чтобы отличить "поле не прислали" (nil) от "прислали пустую строку" ("").
type UpdateProfileRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}
