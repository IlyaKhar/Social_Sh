## Общая идея бэка

Бэк — это Fiber‑API, которое:
- отдаёт данные для магазина (`/api/products`, `/api/products/:slug`);
- отдаёт данные для галереи (`/api/gallery`);
- отдаёт статические страницы “Информация” (`/api/pages/:slug`);
- отдаёт данные личного кабинета (`/api/account/me`, `/api/account/orders`).

Хранилище (БД) и конкретную реализацию ты выбираешь сам (Postgres, Mongo, файловый storage и т.д.). В коде уже заложены интерфейсы репозиториев и модели, чтобы было понятно, что нужно реализовать.

---

## Конфиг: `backend/internal/config/config.go`

**Что делает сейчас**
- Через `godotenv.Load()` подтягивает переменные окружения из `.env` (если файл есть).
- Читает и складывает в структуру `Config`:
  - `Port`
  - `DBHost`, `DBPort`, `DBUser`, `DBPassword`, `DBName`
  - `JwtSecret`, `RefreshSecret`
  - `BaseUrl`
- Даёт дефолты, если переменная не задана (например, `DB_PORT=5432`, `DB_NAME=postgres`).

**TODO (Postgres‑конфиг)**
- Добавить метод‑хелпер для сборки DSN:

  ```go
  func (c *Config) PostgresDSN() string {
    return fmt.Sprintf(
      "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
      c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
    )
  }
  ```

- Опционально:
  - поле `Env` (`dev` / `prod`) — чтобы включать/выключать дебаг и расширенный логгер;
  - поле `DBSSLMode`, если будешь ходить в облачный Postgres.

**Как это используется**
- В `main.go` вызывается `cfg := config.Load()`.
- Через `cfg.PostgresDSN()` собирается строка подключения к Postgres.
- Остальные поля идут в auth, формирование ссылок (`BaseUrl`) и т.п.

---

## Точка входа: `backend/cmd/api/main.go`

**Что делает сейчас**
- Загружает конфиг: `cfg := config.Load()`.
- Собирает DSN: `dsn := cfg.PostgresDSN()`.
- Открывает соединение к Postgres через `internal/db.OpenPostgres(dsn)`.
- Создаёт `store := repository.NewStore(sqlDB)` и присваивает его глобальной переменной `handlers.Repo`.
- Создаёт Fiber‑приложение: `app := fiber.New()`, вешает роуты `routes.Register(app)` и слушает порт `cfg.Port`.

**TODO — что дописать**
- Middleware:
  - логгер (`app.Use(logger.New(...))`);
  - recover от паник;
  - CORS (если фронт и бэк будут на разных доменах);
  - позже — auth middleware, который будет класть `userID` в `c.Locals`.

---

## Роутер: `backend/internal/routes/routes.go`

**Что тут написано и как это работает**

Файл `routes.go` — это «карта» всего API. Одна функция `Register()` вешает ВСЕ эндпоинты на Fiber-приложение. Роуты разбиты на 4 уровня доступа — от открытых до админских.

```go
// Register принимает Fiber-приложение и секрет для JWT.
// jwtSecret прокидывается из main.go → cfg.JwtSecret
func Register(app *fiber.App, jwtSecret string) {
    // Все эндпоинты живут под /api
    // Фронт ходит на /api/products, /api/auth/sign-in и т.д.
    api := app.Group("/api")
    // ...
}
```

> `app.Group("/api")` — создаёт префикс. Все роуты внутри автоматически получают `/api/...`.
> `jwtSecret` передаётся в middleware, чтобы он мог валидировать JWT-токены.

### Уровни доступа

| Уровень | Префикс | Middleware | Кто имеет доступ |
|---------|---------|-----------|-----------------|
| Публичные | `/api/products`, `/api/gallery`, `/api/pages` | нет | Все |
| Auth | `/api/auth/*` | нет (кроме logout/is-admin) | Все (sign-up, sign-in) |
| Личный кабинет | `/api/account/*` | `Protected(jwtSecret)` | Авторизованные юзеры |
| Админка | `/api/admin/*` | `Protected(jwtSecret)` + `AdminOnly()` | Только админы |

### 1. Публичные роуты ✅ (готовы)

Это эндпоинты, которые работают без авторизации — любой может их дёрнуть.

```go
// Магазин — товары с фильтрацией
api.Get("/products", handlers.GetProducts)
// → GET /api/products?new=true&sale=true&page=1&limit=20
// Хендлер читает query-параметры и фильтрует через Repo.Products.List()

api.Get("/products/:slug", handlers.GetProduct)
// → GET /api/products/hoodie-black
// :slug — динамический параметр, хендлер берёт его через c.Params("slug")

// Галерея — фотки с фильтром по категории
api.Get("/gallery", handlers.GetGalleryItems)
// → GET /api/gallery?category=intro

// Инфо-страницы — статический контент
api.Get("/pages/:slug", handlers.GetPage)
// → GET /api/pages/payment | delivery | returns | contacts
```

### 2. Auth-роуты ✅ (готовы, хендлеры — заглушки)

Отдельная функция `authRoutes()` — чтобы не засирать `Register()`.

```go
func authRoutes(api fiber.Router, jwtSecret string) {
    a := api.Group("/auth")
    // Теперь все роуты ниже будут /api/auth/...

    a.Post("/sign-up", handlers.SignUp)
    // → POST /api/auth/sign-up
    // Body: { "email": "test@mail.com", "password": "123456", "name": "Илья" }
    // Ответ: { "access": "<jwt>", "refresh": "<jwt>" }
    // Внутри: хешируем пароль bcrypt → создаём юзера в БД → генерим токены

    a.Post("/sign-in", handlers.SignIn)
    // → POST /api/auth/sign-in
    // Body: { "email": "...", "password": "..." }
    // Ответ: { "access": "<jwt>", "refresh": "<jwt>" }
    // Внутри: ищем юзера по email → сравниваем bcrypt → генерим токены

    a.Post("/refresh", handlers.RefreshToken)
    // → POST /api/auth/refresh
    // Body: { "refresh": "<old_refresh_jwt>" }
    // Ответ: { "access": "<new_jwt>" }
    // Внутри: парсим refresh-токен → проверяем не инвалидирован ли → выдаём новый access

    // Эти два — защищённые (нужен валидный JWT в заголовке Authorization)
    a.Post("/logout", middleware.Protected(jwtSecret), handlers.Logout)
    // → POST /api/auth/logout (+ заголовок Authorization: Bearer <token>)
    // Внутри: инвалидируем refresh-токен юзера в БД

    a.Get("/is-admin", middleware.Protected(jwtSecret), handlers.IsAdmin)
    // → GET /api/auth/is-admin (+ заголовок Authorization: Bearer <token>)
    // Ответ: { "isAdmin": true/false }
    // Внутри: читаем role из c.Locals (положил middleware) и сравниваем с "admin"
}
```

> **Зачем `middleware.Protected(jwtSecret)` на logout и is-admin?**
> Потому что logout без токена бессмысленен — не знаем чей refresh инвалидировать.
> А is-admin проверяет роль конкретного юзера.

### 3. Личный кабинет ✅ (готов, хендлеры — частично заглушки)

Вся группа `/account` обёрнута в `middleware.Protected()` — без JWT сюда не попасть.

```go
func accountRoutes(api fiber.Router, jwtSecret string) {
    acc := api.Group("/account", middleware.Protected(jwtSecret))
    // middleware.Protected() вызывается ДО любого хендлера в этой группе.
    // Он проверяет JWT и кладёт userID в c.Locals("userID").
    // Если токен битый — юзер получит 401 и до хендлера не дойдёт.

    acc.Get("/me", handlers.GetAccountMe)
    // → GET /api/account/me
    // Внутри: userID := c.Locals("userID") → Repo.Account.GetUserByID(userID)
    // Ответ: { "user": { "id", "email", "name" } }

    acc.Get("/orders", handlers.GetOrders)
    // → GET /api/account/orders
    // Внутри: userID из Locals → Repo.Account.ListOrdersByUser(userID)
    // Ответ: { "items": [ { "id", "status", "total" }, ... ] }

    acc.Patch("/me", handlers.UpdateProfile)
    // → PATCH /api/account/me
    // Body: { "name": "Новое имя" }  ← только изменённые поля
    // TODO: реализовать в хендлере
}
```

> **Почему `Protected` на всю группу, а не на каждый роут?**
> Потому что ВСЕ эндпоинты в `/account` требуют авторизации.
> `api.Group("/account", middleware)` — middleware применяется автоматически ко всем роутам группы.

### 4. Админка ✅ (готова, хендлеры — заглушки)

Два middleware подряд: сначала `Protected` (проверка JWT), потом `AdminOnly` (проверка роли).

```go
func adminRoutes(api fiber.Router, jwtSecret string) {
    adm := api.Group("/admin",
        middleware.Protected(jwtSecret),  // 1. Есть ли валидный JWT?
        middleware.AdminOnly(),           // 2. role == "admin"? Если нет → 403
    )

    // ── Товары (полный CRUD) ──
    adm.Get("/products", handlers.AdminListProducts)
    // Отличие от публичного GetProducts: тут показываем ВСЁ — скрытые, черновики и т.д.

    adm.Post("/products", handlers.AdminCreateProduct)
    // Body: { "slug", "title", "price", "images": [...], "isNew": true }
    // Создаёт новый товар в БД

    adm.Get("/products/:id", handlers.AdminGetProduct)
    // По ID (не slug!) — для формы редактирования в админке

    adm.Patch("/products/:id", handlers.AdminUpdateProduct)
    // Частичное обновление — присылаешь только изменённые поля

    adm.Delete("/products/:id", handlers.AdminDeleteProduct)
    // Удаление товара

    // ── Галерея ──
    adm.Get("/gallery", handlers.AdminListGalleryItems)
    adm.Post("/gallery", handlers.AdminCreateGalleryItem)
    adm.Patch("/gallery/:id", handlers.AdminUpdateGalleryItem)
    adm.Delete("/gallery/:id", handlers.AdminDeleteGalleryItem)

    // ── Инфо-страницы ──
    adm.Get("/pages", handlers.AdminListPages)
    adm.Patch("/pages/:slug", handlers.AdminUpdatePage)
    // Тут slug а не id — потому что страницы идентифицируются по slug (payment, delivery и т.д.)
}
```

> **Цепочка middleware**: запрос → `Protected` → `AdminOnly` → хендлер.
> Если JWT невалидный — `Protected` вернёт 401, до `AdminOnly` не дойдёт.
> Если JWT ок, но role != "admin" — `AdminOnly` вернёт 403.

---

## Middleware: `backend/internal/middleware/auth.go`

**Что тут написано и как это работает**

Два middleware — `Protected()` и `AdminOnly()`. Оба возвращают `fiber.Handler` (замыкание).

### Protected(jwtSecret string)

```go
func Protected(jwtSecret string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Достаём заголовок Authorization: "Bearer eyJhbGciOi..."
        authHeader := c.Get("Authorization")
        // Если пусто → 401

        // 2. Разбиваем на ["Bearer", "eyJhbGciOi..."]
        parts := strings.SplitN(authHeader, " ", 2)
        // Если формат не тот → 401

        // 3. Парсим JWT и проверяем подпись
        token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
            // Разрешаем только HMAC-подпись (HS256/HS384/HS512)
            // Чтобы никто не подсунул "alg: none" или RSA-ключ
            return []byte(jwtSecret), nil
        })
        // Если токен невалидный/просроченный → 401

        // 4. Достаём данные из токена (claims)
        claims := token.Claims.(jwt.MapClaims)
        userID := claims["sub"].(string)   // "sub" — стандартный claim для ID юзера
        role   := claims["role"].(string)  // кастомный claim — "user" или "admin"

        // 5. Кладём в контекст Fiber — хендлеры заберут через c.Locals("userID")
        c.Locals("userID", userID)
        c.Locals("role", role)

        // 6. Передаём дальше по цепочке
        return c.Next()
    }
}
```

> **Почему замыкание `func(...) fiber.Handler`?**
> Потому что middleware нужен `jwtSecret`, а сигнатура Fiber-middleware — `func(c *fiber.Ctx) error`.
> Замыкание захватывает `jwtSecret` и возвращает функцию с правильной сигнатурой.

### AdminOnly()

```go
func AdminOnly() fiber.Handler {
    return func(c *fiber.Ctx) error {
        role := c.Locals("role").(string)
        // Protected уже положил role в Locals
        // Если role != "admin" → 403 Forbidden
        // Если admin → c.Next() (пропускаем дальше)
    }
}
```

> Вешается СТРОГО ПОСЛЕ `Protected()` — потому что именно `Protected` кладёт `role` в `c.Locals`.

---

## Auth-хендлеры: `backend/internal/handlers/auth.go`

**Что тут написано и как это работает**

Пять хендлеров для работы с аутентификацией. Сейчас все (кроме `IsAdmin`) — заглушки, возвращающие 501.

### SignUp — регистрация

```go
func SignUp(c *fiber.Ctx) error {
    // Логика которую надо реализовать:
    // 1. c.BodyParser(&input) — парсим { email, password, name }
    // 2. Repo.Account.GetUserByEmail(email) — проверяем что email свободен
    // 3. bcrypt.GenerateFromPassword(password) — хешируем пароль
    // 4. Repo.Account.CreateUser(user) — сохраняем в БД
    // 5. Генерируем JWT: access (15 мин) + refresh (7 дней)
    //    claims: { "sub": userID, "role": "user", "exp": ... }
    // 6. Возвращаем { "access": "...", "refresh": "..." }
}
```

### SignIn — логин

```go
func SignIn(c *fiber.Ctx) error {
    // 1. Парсим { email, password }
    // 2. Repo.Account.GetUserByEmail(email)
    // 3. bcrypt.CompareHashAndPassword(user.PasswordHash, password)
    //    Если не совпадает → 401 "неверный email или пароль"
    // 4. Генерируем access + refresh JWT
    // 5. Возвращаем токены
}
```

### RefreshToken — обновление access-токена

```go
func RefreshToken(c *fiber.Ctx) error {
    // 1. Парсим { "refresh": "<old_jwt>" }
    // 2. jwt.Parse(refreshToken, refreshSecret) — валидируем
    // 3. Достаём userID из claims
    // 4. Генерируем новый access-токен
    //    (refresh НЕ перегенерируем — иначе бесконечная сессия)
    // 5. Возвращаем { "access": "<new_jwt>" }
}
```

### Logout — выход

```go
func Logout(c *fiber.Ctx) error {
    // Вызывается через Protected → userID уже в c.Locals
    // 1. Достаём userID
    // 2. Инвалидируем refresh-токен (удаляем из таблицы refresh_tokens)
    // 3. 200 OK
}
```

### IsAdmin — проверка роли (уже работает)

```go
func IsAdmin(c *fiber.Ctx) error {
    role := c.Locals("role").(string)  // Protected положил
    return c.JSON(fiber.Map{ "isAdmin": role == "admin" })
}
```

---

## Account-хендлеры: `backend/internal/handlers/account.go`

**Что тут написано и как это работает**

Три хендлера для личного кабинета. Все роуты защищены `Protected()` — `userID` уже в `c.Locals`.

### GetAccountMe — профиль ✅ (готов)

```go
func GetAccountMe(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)  // положил middleware Protected
    user, err := Repo.Account.GetUserByID(userID)
    // Если err → 500
    // Если user == nil → 404
    // Иначе → { "user": { "id", "email", "name" } }
}
```

### GetOrders — заказы ✅ (готов)

```go
func GetOrders(c *fiber.Ctx) error {
    userID := c.Locals("userID").(string)
    orders, err := Repo.Account.ListOrdersByUser(userID)
    // → { "items": [ ... ] }
}
```

### UpdateProfile — обновление профиля (заглушка)

```go
func UpdateProfile(c *fiber.Ctx) error {
    // TODO:
    // 1. userID из Locals
    // 2. c.BodyParser(&updates) — { "name": "...", "email": "..." }
    // 3. Repo.Account.UpdateUser(userID, updates)
    // 4. Вернуть обновлённого юзера
}
```

---

## Admin-хендлеры: `backend/internal/handlers/admin.go`

**Что тут написано и как это работает**

Все хендлеры — заглушки (501 Not Implemented). Защищены `Protected + AdminOnly`.

Структура одинаковая для товаров, галереи и страниц:

| Метод | Роут | Что делает |
|-------|------|-----------|
| `GET /admin/products` | `AdminListProducts` | Список ВСЕХ товаров (включая скрытые) |
| `POST /admin/products` | `AdminCreateProduct` | Создать товар: BodyParser → валидация → Repo.Create |
| `GET /admin/products/:id` | `AdminGetProduct` | Один товар по ID (для формы редактирования) |
| `PATCH /admin/products/:id` | `AdminUpdateProduct` | Частичное обновление (только изменённые поля) |
| `DELETE /admin/products/:id` | `AdminDeleteProduct` | Удаление товара |

То же самое для галереи (`/admin/gallery`) и страниц (`/admin/pages`).

**Чтобы хендлеры заработали, нужно добавить методы в репозитории:**
- `ProductRepository`: `ListAll()`, `Create()`, `GetByID()`, `Update()`, `Delete()`
- `GalleryRepository`: `ListAll()`, `Create()`, `Update()`, `Delete()`
- `PageRepository`: `ListAll()`, `Update()`

---

## Модели: `backend/internal/models/models.go`

**Что описано**
- `Product` — товар:
  - `ID`, `Slug`, `Title`, `Description`, `Price`, `Currency`, `Images`, флаги `IsNew`, `IsOnSale`.
- `GalleryItem` — элемент галереи:
  - `Category` (`intro`, `tattoo`, `tokyo` и т.п.), `Title`, `Image`, `Order`.
- `Page` — статическая страница:
  - `Slug` (`payment`, `delivery`, `returns`, `contacts`), `Title`, `Content`.
- `User` — пользователь:
  - пока только `ID`, `Email`, `Name`.
- `Order` — заказ:
  - `ID`, `UserID` (+ TODO: состав заказа и статусы).

**TODO — как адаптировать под БД**
- Добавить теги под ORM / драйвер, который будешь использовать:
  - для SQL: `db:"column_name"`;  
  - для Mongo: `bson:"field_name"`.
- Продумать индексы:
  - `Product.Slug` — уникальный индекс (используется в `GET /products/:slug`);
  - `GalleryItem.Category, Order` — составной индекс.
- Для `Order`:
  - добавить `Items []OrderItem`, `Total`, `Status`, `CreatedAt`, `UpdatedAt` и т.п.

---

## Репозитории: `backend/internal/repository/`

### Что тут вообще происходит

Репозитории — это слой между хендлерами и БД. Хендлеры не знают как данные хранятся (Postgres? Mongo? файлы?). Они вызывают методы интерфейса (`Repo.Products.List(...)`) а конкретная реализация уже решает как ходить в БД.

**Файлы:**

| Файл | Что внутри |
|------|-----------|
| `repository.go` | Интерфейсы + Store (агрегатор) + конструктор NewStore |
| `postgres_products.go` | SQL-реализация ProductRepository |
| `postgres_gallery.go` | SQL-реализация GalleryRepository |
| `postgres_pages.go` | SQL-реализация PageRepository |
| `postgres_account.go` | SQL-реализация AccountRepository |

---

### `repository.go` — как это читается

```go
// Интерфейс — контракт. Говорит "кто бы ты ни был — умей делать вот это".
// Хендлеры работают с интерфейсом, а не с конкретной структурой.
type ProductRepository interface {
    // ── Публичные (для фронта) ──
    List(newOnly, saleOnly bool, page, limit int) ([]models.Product, error)
    // → SELECT * FROM products WHERE (is_new = $1 OR ...) ORDER BY ... LIMIT $2 OFFSET $3
    // newOnly/saleOnly — фильтры из query-параметров (?new=true&sale=true)
    // page/limit — пагинация: page=2, limit=20 → OFFSET 20

    GetBySlug(slug string) (*models.Product, error)
    // → SELECT * FROM products WHERE slug = $1 LIMIT 1
    // Возвращает указатель: если nil — товар не найден, хендлер вернёт 404

    // ── Админские (CRUD) ──
    ListAll() ([]models.Product, error)
    // → SELECT * FROM products ORDER BY created_at DESC
    // Без фильтров — показать ВСЁ для админки

    GetByID(id string) (*models.Product, error)
    // → SELECT * FROM products WHERE id = $1

    Create(product *models.Product) error
    // → INSERT INTO products (slug, title, ...) VALUES ($1, $2, ...) RETURNING id
    // Репозиторий ЗАПОЛНЯЕТ product.ID — чтобы хендлер вернул его клиенту

    Update(id string, product *models.Product) (*models.Product, error)
    // → UPDATE products SET title=$1, price=$2 WHERE id=$3 RETURNING *
    // Возвращает обновлённый товар (или nil если id не найден)

    Delete(id string) error
    // → DELETE FROM products WHERE id = $1
    // Если 0 rows affected — можно вернуть кастомную ошибку "not found"
}
```

Аналогично для остальных:

```go
type GalleryRepository interface {
    ListByCategory(category string) ([]models.GalleryItem, error)
    // Если category пустой — вернуть всё. Иначе WHERE category = $1
    // ORDER BY sort_order — чтобы фотки шли в правильном порядке

    ListAll() ([]models.GalleryItem, error)
    Create(item *models.GalleryItem) error
    Update(id string, item *models.GalleryItem) (*models.GalleryItem, error)
    Delete(id string) error
}

type PageRepository interface {
    GetBySlug(slug string) (*models.Page, error)
    // slug = "payment" | "delivery" | "returns" | "contacts"
    // Таблица маленькая (4 строки), slug — первичный ключ

    ListAll() ([]models.Page, error)
    Update(slug string, page *models.Page) (*models.Page, error)
    // Тут slug вместо id — потому что у страниц slug IS первичный ключ
}

type AccountRepository interface {
    GetUserByID(id string) (*models.User, error)
    // → SELECT id, email, name, role FROM users WHERE id = $1
    // password_hash НЕ селектим (json:"-" не спасёт если мы его вообще не достаём)

    GetUserByEmail(email string) (*models.User, error)
    // → SELECT * FROM users WHERE email = $1
    // Используется в SignUp (проверка "email занят?") и SignIn (поиск юзера)
    // Тут ДОСТАЁМ password_hash — он нужен для bcrypt.Compare

    CreateUser(user *models.User) error
    // → INSERT INTO users (email, name, password_hash, role) VALUES (...) RETURNING id
    // role по умолчанию "user" (хендлер ставит)

    UpdateUser(id string, req *models.UpdateProfileRequest) (*models.User, error)
    // Частичное обновление — меняем только те поля, которые не nil
    // → UPDATE users SET name = COALESCE($1, name), email = COALESCE($2, email) WHERE id = $3 RETURNING *

    ListOrdersByUser(id string) ([]models.Order, error)
    // → SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC
    // Потом для каждого order — подгрузить items:
    // → SELECT * FROM order_items WHERE order_id = $1
    // И положить в order.Items
}
```

### Store — агрегатор

```go
// Store собирает все репозитории в одну структуру.
// Хендлеры получают его через глобальную переменную handlers.Repo.
type Store struct {
    Products ProductRepository  // работа с товарами
    Gallery  GalleryRepository  // работа с галереей
    Pages    PageRepository     // работа со страницами
    Account  AccountRepository  // работа с юзерами и заказами
}

// NewStore — конструктор. Принимает *sql.DB (одно соединение к Postgres)
// и создаёт все SQL-репозитории.
// Вызывается в main.go: store := repository.NewStore(sqlDB)
func NewStore(db *sql.DB) *Store {
    return &Store{
        Products: NewProductSQLRepo(db),  // каждый репозиторий получает один и тот же db
        Gallery:  NewGallerySQLRepo(db),
        Pages:    NewPageSQLRepo(db),
        Account:  NewAccountSQLRepo(db),
    }
}
```

> **Зачем интерфейсы, а не просто структуры?**
> Потому что завтра ты можешь заменить Postgres на Mongo — и хендлеры НЕ изменятся.
> Меняешь только реализацию (новый файл `mongo_products.go`) и конструктор NewStore.

---

### SQL-реализации — как их писать

Каждый файл `postgres_*.go` — это структура с `*sql.DB` + методы интерфейса.

#### `postgres_products.go` — пример (частично написан)

```go
type ProductSQLRepo struct {
    db *sql.DB  // соединение к Postgres, общее на всё приложение
}

func NewProductSQLRepo(db *sql.DB) *ProductSQLRepo {
    return &ProductSQLRepo{db: db}
}
```

Как это читается: создаём структуру, сохраняем ссылку на БД. `New...` — конструктор, вызывается из `NewStore`.

#### Метод List — пример реализации

```go
func (r *ProductSQLRepo) List(newOnly, saleOnly bool, page, limit int) ([]models.Product, error) {
    // 1. Собираем SQL динамически — в зависимости от фильтров
    query := "SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale FROM products WHERE 1=1"
    args := []interface{}{}
    argIdx := 1

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

    // 2. Пагинация
    offset := (page - 1) * limit
    query += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
    args = append(args, limit, offset)

    // 3. Выполняем запрос
    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // 4. Сканируем строки в слайс моделей
    var items []models.Product
    for rows.Next() {
        var p models.Product
        if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description,
            &p.Price, &p.Currency, &p.Images, &p.IsNew, &p.IsOnSale); err != nil {
            return nil, err
        }
        items = append(items, p)
    }

    return items, rows.Err()
}
```

> **Паттерн одинаковый для всех List-методов:**
> 1. Собрать SQL (SELECT + WHERE + ORDER + LIMIT)
> 2. `r.db.Query(query, args...)` → получить `rows`
> 3. `defer rows.Close()` — ОБЯЗАТЕЛЬНО, иначе утечка соединений
> 4. `for rows.Next()` → `rows.Scan()` → собрать слайс
> 5. Вернуть слайс + `rows.Err()`

#### Метод GetBySlug — пример реализации

```go
func (r *ProductSQLRepo) GetBySlug(slug string) (*models.Product, error) {
    var p models.Product
    err := r.db.QueryRow(
        "SELECT id, slug, title, description, price, currency, images, is_new, is_on_sale FROM products WHERE slug = $1",
        slug,
    ).Scan(&p.ID, &p.Slug, &p.Title, &p.Description, &p.Price, &p.Currency, &p.Images, &p.IsNew, &p.IsOnSale)

    if err == sql.ErrNoRows {
        return nil, nil  // не найден — НЕ ошибка, просто nil
    }
    if err != nil {
        return nil, err  // реальная ошибка БД
    }

    return &p, nil
}
```

> **Паттерн для всех Get-методов (одна строка):**
> 1. `r.db.QueryRow(query, args...)` → одна строка
> 2. `.Scan(...)` → заполняем модель
> 3. `sql.ErrNoRows` → вернуть nil, nil (хендлер сам решит — 404 или что)
> 4. Другая ошибка → вернуть nil, err

#### Метод Create — пример реализации

```go
func (r *ProductSQLRepo) Create(product *models.Product) error {
    return r.db.QueryRow(
        `INSERT INTO products (slug, title, description, price, currency, images, is_new, is_on_sale)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
        product.Slug, product.Title, product.Description, product.Price,
        product.Currency, product.Images, product.IsNew, product.IsOnSale,
    ).Scan(&product.ID)
    // RETURNING id → БД генерирует id и мы сразу сканируем его в product.ID
    // Хендлер потом отдаст product клиенту — уже с ID
}
```

> **Паттерн для Create:**
> 1. `INSERT ... RETURNING id` — Postgres умеет возвращать сгенерированные поля
> 2. `.Scan(&product.ID)` — записываем ID прямо в переданную структуру

#### Метод Delete — пример реализации

```go
func (r *ProductSQLRepo) Delete(id string) error {
    result, err := r.db.Exec("DELETE FROM products WHERE id = $1", id)
    if err != nil {
        return err
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return sql.ErrNoRows  // ничего не удалилось — товар не найден
    }
    return nil
}
```

> **Паттерн для Delete:**
> 1. `r.db.Exec(...)` — не возвращает строки, только результат
> 2. `result.RowsAffected()` — сколько строк затронуто
> 3. Если 0 — товар не существовал

---

### Файлы которые нужно реализовать

| Файл | Что писать | Сколько методов |
|------|-----------|-----------------|
| `postgres_products.go` | List, GetBySlug, ListAll, GetByID, Create, Update, Delete | 7 |
| `postgres_gallery.go` | ListByCategory, ListAll, Create, Update, Delete | 5 |
| `postgres_pages.go` | GetBySlug, ListAll, Update | 3 |
| `postgres_account.go` | GetUserByID, GetUserByEmail, CreateUser, UpdateUser, ListOrdersByUser | 5 |

**Порядок действий:**
1. Начни с `postgres_products.go` — он самый жирный, но после него остальные пойдут по шаблону.
2. Потом `postgres_account.go` — нужен для auth (SignUp/SignIn используют GetUserByEmail/CreateUser).
3. `postgres_gallery.go` и `postgres_pages.go` — самые простые, копипаста с products.

---

## Хендлеры магазина: `backend/internal/handlers/shop.go`

**Как работают сейчас**
- Есть глобальная переменная `var Repo *repository.Store` — её нужно инициализировать в `main.go`.
- `GetProducts`:
  - читает query‑параметры:
    - `new` → `newOnly` (bool);
    - `sale` → `saleOnly` (bool);
    - `page` / `limit` → пагинация (по умолчанию 1 и 20).
  - вызывает `Repo.Products.List(newOnly, saleOnly, page, limit)`;
  - возвращает JSON: `{ "items": [...] }`.
- `GetProduct`:
  - читает `slug` из `c.Params("slug")`;
  - вызывает `Repo.Products.GetBySlug(slug)`;
  - если товар не найден (`item == nil`) → `404`;
  - иначе `{ "item": { ... } }`.

**TODO — что нужно сделать**
- Реализовать методы репозитория так, чтобы они реально ходили в БД.
- Добавить обработку ошибок:
  - если `Repo.Products.List(...)` вернул ошибку → `500` и JSON с описанием;
  - если `GetBySlug` вернул ошибку вида “не найдено” → `404`, остальные → `500`.
- При желании добавить кэширование популярных запросов (новые/скидки).

---

## Хендлеры галереи: `backend/internal/handlers/gallery.go`

**Как работают сейчас**
- Читают `category` из `?category=...` (если пусто — вернуть все или дефолтную категорию — на твой выбор).
- Вызывают `Repo.Gallery.ListByCategory(category)`.
- Возвращают `{ "items": [...] }`.

**TODO**
- В репозитории:
  - если `category == ""` → вернуть все или только “intro”;
  - добавить сортировку по полю `Order`.
- В хендлере:
  - обрабатывать ошибку репозитория (500).

---

## Хендлеры статических страниц: `backend/internal/handlers/pages.go`

**Как работают сейчас**
- Читают `slug` из `/:slug` (`payment`, `delivery`, `returns`, `contacts`).
- Вызывают `Repo.Pages.GetBySlug(slug)`.
- Если страница не найдена → `404`.
- Если найдена → возвращают структуру `Page` как JSON.

**TODO**
- В репозитории решить, где хранятся текстовые страницы:
  - в БД (таблица `pages`);
  - в файловой системе (например, Markdown‑файлы);
  - в памяти (если их мало и они редко меняются).
- Если выберешь Markdown‑файлы:
  - можно положить их в отдельную папку и подгружать при старте сервера или по запросу.

---

## Хендлеры аккаунта: `backend/internal/handlers/account.go`

**Как работают сейчас**
- Используют `Repo.Account`.
- `GetAccountMe`:
  - ожидает, что в `c.Locals("userID")` лежит ID пользователя (это должен сделать auth‑middleware, которого пока нет);
  - вызывает `Repo.Account.GetUserByID(userID)`;
  - если пользователь не найден → `401 Unauthorized`;
  - иначе `{ "user": { ... } }`.
- `GetOrders`:
  - берёт тот же `userID`;
  - вызывает `Repo.Account.ListOrdersByUser(userID)`;
  - возвращает `{ "items": [ ... ] }`.

**TODO**
- Реализовать авторизацию:
  - middleware, который:
    - читает токен из заголовка/куки;
    - валидирует его;
    - кладёт `userID` в `c.Locals("userID")`;
  - повесить middleware на группу `api.Group("/account")`.
- Реализация методов репозитория:
  - `GetUserByID` → выборка из `users`;
  - `ListOrdersByUser` → выборка из `orders` по `user_id` с сортировкой по дате.

---

## Связка с фронтом

Фронт уже ожидает следующие эндпоинты (через `TODO` в компонентах):

- Магазин:
  - `GET /api/products` — грид всех товаров (`app/shop/page.tsx`).
  - `GET /api/products?new=true` — новые (`/shop/new`).
  - `GET /api/products?sale=true` — скидки (`/shop/sale`).
- Галерея:
  - `GET /api/gallery?category=intro` — табы в `/gallery`.
- Информация:
  - `GET /api/pages/payment` — `/info/payment`.
  - `GET /api/pages/delivery` — `/info/delivery`.
  - `GET /api/pages/returns` — `/info/returns`.
  - `GET /api/pages/contacts` — `/contacts`.
- Аккаунт:
  - `GET /api/account/me` и `GET /api/account/orders` — `/account`.

Твоя задача по бэку — реализовать репозитории и бизнес‑логику так, чтобы все эти end‑point’ы начали отдавать реальные данные.

---

## Практический план: что именно тебе дописать (пример под Postgres)

Ниже — один возможный вариант реализации через Postgres. Если выберешь другую БД (Mongo, SQLite), структура файлов может быть такой же, просто меняется код внутри.

### 1. Подключение к БД

**Новые файлы, которые имеет смысл создать**

- `backend/internal/db/postgres.go`

**Содержимое (идея, а не готовый код)**

- Функция:

```go
package db

import (
  "database/sql"
  _ "github.com/lib/pq"
)

func OpenPostgres(dsn string) (*sql.DB, error) {
  db, err := sql.Open("postgres", dsn)
  if err != nil {
    return nil, err
  }
  if err := db.Ping(); err != nil {
    return nil, err
  }
  return db, nil
}
```

**Что надо сделать в `main.go`**

- В `config.Config` добавить поле `DBURL string`.
- В `config.Load()` читать `os.Getenv("DATABASE_URL")`.
- В `main.go`:
  - вызвать `db.OpenPostgres(cfg.DBURL)`;
  - на фатальной ошибке — `log.Fatalf`.

### 2. Реализация Store и репозиториев

**Файлы, которые стоит создать**

- `backend/internal/repository/postgres_products.go`
- `backend/internal/repository/postgres_gallery.go`
- `backend/internal/repository/postgres_pages.go`
- `backend/internal/repository/postgres_account.go`

**Пример: `postgres_products.go`**

Идея кода:

```go
type ProductSQLRepo struct {
  db *sql.DB
}

func NewProductSQLRepo(db *sql.DB) *ProductSQLRepo {
  return &ProductSQLRepo{db: db}
}

func (r *ProductSQLRepo) List(newOnly, saleOnly bool, page, limit int) ([]models.Product, error) {
  // TODO: собрать SQL с учётом флагов newOnly/saleOnly и пагинации.
  // SELECT ... FROM products WHERE ... ORDER BY created_at DESC LIMIT $1 OFFSET $2;
}

func (r *ProductSQLRepo) GetBySlug(slug string) (*models.Product, error) {
  // SELECT ... FROM products WHERE slug = $1 LIMIT 1;
}
```

**Пример конструктора Store: `repository.go`**

```go
func NewStore(db *sql.DB) *Store {
  return &Store{
    Products: NewProductSQLRepo(db),
    Gallery:  NewGallerySQLRepo(db),
    Pages:    NewPageSQLRepo(db),
    Account:  NewAccountSQLRepo(db),
  }
}
```

**Что нужно дописать в `main.go`**

- После открытия БД:

```go
store := repository.NewStore(db)
handlers.Repo = store
```

Так все хендлеры (`GetProducts`, `GetGalleryItems`, `GetPage`, `GetAccountMe`) начнут использовать одну и ту же реализацию репозиториев.

### 3. Минимальный SQL‑скиллсет, который нужен

- Таблица `products`:
  - `id`, `slug UNIQUE`, `title`, `description`, `price`, `currency`, `images` (можно как `jsonb`), `is_new`, `is_on_sale`.
- Таблица `gallery_items`:
  - `id`, `category`, `title`, `image`, `order`.
- Таблица `pages`:
  - `slug PRIMARY KEY`, `title`, `content`.
- Таблицы `users`, `orders` — по вкусу, можно добавить позже.


