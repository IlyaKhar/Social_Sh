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
- `GetGalleryItems`:
  - читает query‑параметр `category` из `?category=intro` (или `tattoo`, `tokyo`, и т.д.);
  - если `category` пустой (`""`), нужно решить что возвращать:
    - **Вариант 1**: вернуть все элементы галереи (все категории);
    - **Вариант 2**: вернуть дефолтную категорию (например, `"intro"`).
  - вызывает `Repo.Gallery.ListByCategory(category)`;
  - возвращает JSON: `{ "items": [...] }`.

**TODO — что нужно сделать**

### 1. В репозитории (`postgres_gallery.go`): метод `ListByCategory`

**Проблема:** сейчас метод `ListByCategory` всегда требует `category` и делает `WHERE category = $1`. Если `category == ""`, запрос не найдёт ничего или вернёт ошибку.

**Решение — два варианта:**

#### Вариант А: Если `category == ""` → вернуть все элементы

```go
func (r *GallerySQLRepo) ListByCategory(category string) ([]models.GalleryItem, error) {
    var query string
    var args []interface{}
    
    if category == "" {
        // Если категория не указана — возвращаем все элементы
        query = `SELECT id, category, title, image, sort_order
                 FROM gallery_items 
                 ORDER BY sort_order ASC`
        // args остаётся пустым
    } else {
        // Если категория указана — фильтруем по ней
        query = `SELECT id, category, title, image, sort_order
                 FROM gallery_items 
                 WHERE category = $1 
                 ORDER BY sort_order ASC`
        args = []interface{}{category}
    }
    
    rows, err := r.db.Query(query, args...)
    // ... остальной код как обычно
}
```

**Как это читается:**
- Если `category == ""` → делаем SELECT без WHERE, получаем все элементы.
- Если `category != ""` → делаем SELECT с `WHERE category = $1`, получаем только элементы этой категории.
- В обоих случаях сортируем по `sort_order ASC` (чтобы админ мог расставлять порядок вручную).

#### Вариант Б: Если `category == ""` → вернуть только дефолтную категорию (например, "intro")

```go
func (r *GallerySQLRepo) ListByCategory(category string) ([]models.GalleryItem, error) {
    // Если категория не указана — используем дефолтную
    if category == "" {
        category = "intro"
    }
    
    query := `SELECT id, category, title, image, sort_order
              FROM gallery_items 
              WHERE category = $1 
              ORDER BY sort_order ASC`
    
    rows, err := r.db.Query(query, category)
    // ... остальной код
}
```

**Как это читается:**
- Если `category == ""` → подставляем `"intro"` как дефолт.
- Всегда делаем `WHERE category = $1` (теперь `category` гарантированно не пустой).
- Сортируем по `sort_order ASC`.

**Рекомендация:** используй **Вариант А** (вернуть все), если фронтенд может показывать галерею без фильтра по категории. Если фронтенд всегда требует категорию — используй **Вариант Б** (дефолт "intro").

**Важно:** сортировка по `sort_order ASC` уже есть в текущем коде репозитория, но убедись что она работает корректно. Поле `sort_order` в БД соответствует полю `Order` в структуре `models.GalleryItem`.

### 2. В хендлере (`gallery.go`): обработка ошибок

**Проблема:** сейчас ошибка из `Repo.Gallery.ListByCategory(category)` игнорируется (`items, _ := ...`). Если БД упадёт или произойдёт другая ошибка, клиент получит пустой массив вместо понятного сообщения об ошибке.

**Решение:**

```go
func GetGalleryItems(c *fiber.Ctx) error {
    category := c.Query("category", "")
    
    items, err := Repo.Gallery.ListByCategory(category)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "ошибка при получении элементов галереи",
        })
    }
    
    // Если items == nil, возвращаем пустой массив (не ошибка)
    if items == nil {
        items = []models.GalleryItem{}
    }
    
    return c.JSON(fiber.Map{"items": items})
}
```

**Как это читается:**
1. Читаем `category` из query (если не указан — будет `""`).
2. Вызываем репозиторий и **проверяем ошибку**.
3. Если ошибка → возвращаем `500` с JSON `{"error": "..."}`.
4. Если ошибки нет, но `items == nil` → возвращаем пустой массив `[]models.GalleryItem{}` (это нормально, если в БД нет элементов).
5. Иначе возвращаем `{"items": [...]}`.

**Не забудь:**
- Добавить импорт `socialsh/backend/internal/models` в `gallery.go` (для типа `[]models.GalleryItem{}`).
- По аналогии с `shop.go` и `account.go` — используй `fiber.Map{"error": "..."}` для ошибок.

### 3. Дополнительно (опционально)

- **Валидация категории:** если фронтенд знает список допустимых категорий (`intro`, `tattoo`, `tokyo`, и т.д.), можно добавить проверку:
  ```go
  validCategories := map[string]bool{
      "intro": true,
      "tattoo": true,
      "tokyo": true,
  }
  if category != "" && !validCategories[category] {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
          "error": "недопустимая категория",
      })
  }
  ```
- **Кэширование:** если галерея редко меняется, можно закэшировать результаты на уровне хендлера (например, в памяти или Redis).

---

## Хендлеры статических страниц: `backend/internal/handlers/pages.go`

**Как работают сейчас**
- `GetPage`:
  - читает `slug` из URL‑параметра `/:slug` (например, `/api/pages/payment`, `/api/pages/delivery`, `/api/pages/returns`, `/api/pages/contacts`);
  - вызывает `Repo.Pages.GetBySlug(slug)`;
  - если страница не найдена (`page == nil`) → возвращает `404`;
  - если найдена → возвращает структуру `Page` как JSON напрямую (без обёртки `{"page": ...}`).

**TODO — что нужно сделать**

### 1. Решить, где хранить текстовые страницы

**Варианты:**

#### Вариант А: В БД (таблица `pages`) — **РЕКОМЕНДУЕТСЯ**

**Плюсы:**
- Админ может редактировать страницы через админ‑панель (без деплоя).
- Единая точка хранения (всё в БД).
- Легко версионировать через `updated_at`.

**Минусы:**
- Нужно создать таблицу и миграции.
- Длинные тексты занимают место в БД.

**Реализация:**

1. **SQL‑схема:**
```sql
CREATE TABLE pages (
    slug VARCHAR(50) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Начальные данные
INSERT INTO pages (slug, title, content) VALUES
('payment', 'Оплата', 'Здесь будет текст про оплату...'),
('delivery', 'Доставка', 'Здесь будет текст про доставку...'),
('returns', 'Возврат', 'Здесь будет текст про возврат...'),
('contacts', 'Контакты', 'Здесь будет текст про контакты...');
```

2. **Репозиторий уже реализован** в `postgres_pages.go`:
   - `GetBySlug(slug)` — делает `SELECT slug, title, content FROM pages WHERE slug = $1`.
   - `ListAll()` — для админ‑панели.
   - `Update(slug, page)` — для редактирования через админку.

3. **Хендлер нужно доработать** (см. ниже).

#### Вариант Б: В файловой системе (Markdown‑файлы)

**Плюсы:**
- Версионирование через Git.
- Легко редактировать в любом редакторе.

**Минусы:**
- Нужен деплой для изменения текста.
- Нет админ‑панели для редактирования.

**Реализация:**

1. **Структура файлов:**
```
backend/
  content/
    pages/
      payment.md
      delivery.md
      returns.md
      contacts.md
```

2. **Реализация репозитория** (новый файл `file_pages.go`):
```go
package repository

import (
    "os"
    "path/filepath"
    "socialsh/backend/internal/models"
    "github.com/russross/blackfriday/v2" // для Markdown → HTML
)

type PageFileRepo struct {
    contentDir string
}

func NewPageFileRepo(contentDir string) *PageFileRepo {
    return &PageFileRepo{contentDir: contentDir}
}

func (r *PageFileRepo) GetBySlug(slug string) (*models.Page, error) {
    filePath := filepath.Join(r.contentDir, "pages", slug+".md")
    
    content, err := os.ReadFile(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, sql.ErrNoRows // страница не найдена
        }
        return nil, err
    }
    
    // Парсим Markdown → HTML (опционально)
    html := blackfriday.Run(content)
    
    return &models.Page{
        Slug:    slug,
        Title:   extractTitle(content), // первая строка как заголовок
        Content: string(html),
    }, nil
}
```

3. **В `main.go`** инициализировать:
```go
pagesRepo := repository.NewPageFileRepo("./content")
store := &repository.Store{
    Pages: pagesRepo,
    // ...
}
```

#### Вариант В: В памяти (hardcoded)

**Плюсы:**
- Самый простой вариант для старта.

**Минусы:**
- Нужен редеплой для изменения текста.
- Не масштабируется.

**Реализация:**

```go
// В postgres_pages.go или отдельном файле
type PageMemoryRepo struct {
    pages map[string]*models.Page
}

func NewPageMemoryRepo() *PageMemoryRepo {
    return &PageMemoryRepo{
        pages: map[string]*models.Page{
            "payment": {
                Slug:    "payment",
                Title:   "Оплата",
                Content: "Здесь текст про оплату...",
            },
            "delivery": {
                Slug:    "delivery",
                Title:   "Доставка",
                Content: "Здесь текст про доставку...",
            },
            // ...
        },
    }
}

func (r *PageMemoryRepo) GetBySlug(slug string) (*models.Page, error) {
    page, ok := r.pages[slug]
    if !ok {
        return nil, sql.ErrNoRows
    }
    return page, nil
}
```

**Рекомендация:** используй **Вариант А (БД)**, потому что:
- Репозиторий уже реализован в `postgres_pages.go`.
- Админ сможет редактировать страницы через админ‑панель.
- Это стандартный подход для CMS.

### 2. Доработать хендлер (`pages.go`)

**Проблема:** сейчас ошибка из `Repo.Pages.GetBySlug(slug)` игнорируется (`page, _ := ...`). Если страница не найдена, репозиторий вернёт `sql.ErrNoRows`, но хендлер проверяет только `if page == nil`.

**Решение:**

```go
package handlers

import (
    "database/sql"
    "github.com/gofiber/fiber/v2"
    "socialsh/backend/internal/repository"
)

func GetPage(c *fiber.Ctx) error {
    slug := c.Params("slug")
    
    // Валидация slug
    if slug == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "slug обязателен",
        })
    }
    
    page, err := Repo.Pages.GetBySlug(slug)
    if err != nil {
        // Если страница не найдена (sql.ErrNoRows) — возвращаем 404
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "страница не найдена",
            })
        }
        // Иначе — серверная ошибка
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "ошибка при получении страницы",
        })
    }
    
    // Если page == nil (на всякий случай)
    if page == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "страница не найдена",
        })
    }
    
    // Возвращаем страницу как JSON (без обёртки)
    return c.JSON(page)
}
```

**Как это читается:**
1. Читаем `slug` из URL‑параметра.
2. Проверяем что `slug` не пустой → если пустой, возвращаем `400`.
3. Вызываем репозиторий и **проверяем ошибку**.
4. Если `sql.ErrNoRows` → страница не найдена, возвращаем `404`.
5. Если другая ошибка → серверная ошибка, возвращаем `500`.
6. Если всё ОК → возвращаем `page` как JSON.

**Не забудь:**
- Добавить импорт `database/sql` для проверки `sql.ErrNoRows`.
- По аналогии с `shop.go` и `account.go` — используй `fiber.Map{"error": "..."}` для ошибок.

### 3. Дополнительно (опционально)

- **Валидация slug:** можно добавить whitelist допустимых slug'ов:
  ```go
  validSlugs := map[string]bool{
      "payment": true,
      "delivery": true,
      "returns": true,
      "contacts": true,
  }
  if !validSlugs[slug] {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
          "error": "недопустимый slug",
      })
  }
  ```

- **Кэширование:** если страницы редко меняются, можно закэшировать их в памяти на уровне хендлера или репозитория.

- **Markdown поддержка:** если хранишь контент в БД как Markdown, можно парсить его в HTML на лету (используй библиотеку типа `github.com/russross/blackfriday/v2`).

---

## Хендлеры аккаунта: `backend/internal/handlers/account.go`

**Как работают сейчас**
- Используют `Repo.Account`.
- `GetAccountMe`:
  - ожидает, что в `c.Locals("userID")` лежит ID пользователя (это должен сделать auth‑middleware);
  - вызывает `Repo.Account.GetUserByID(userID)`;
  - если пользователь не найден → `404 Not Found`;
  - иначе возвращает `{ "user": { "id", "email", "name", "role" } }`.
- `GetOrders`:
  - берёт тот же `userID` из `c.Locals("userID")`;
  - вызывает `Repo.Account.ListOrdersByUser(userID)`;
  - возвращает `{ "items": [ { "id", "userId", "status", "total", "createdAt", "items": [...] }, ... ] }`.
- `UpdateProfile`:
  - принимает PATCH‑запрос с частичными данными (`name` и/или `email`);
  - вызывает `Repo.Account.UpdateUser(userID, &req)`;
  - возвращает обновлённого пользователя.

**TODO — что нужно сделать**

### 1. Реализовать middleware авторизации (`middleware/auth.go`)

**Проблема:** хендлеры `GetAccountMe`, `GetOrders`, `UpdateProfile` ожидают что `userID` уже лежит в `c.Locals("userID")`, но middleware который это делает — ещё не реализован.

**Решение:**

Middleware уже реализован в `backend/internal/middleware/auth.go`! Проверь что он работает так:

```go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
)

// Protected — middleware для проверки JWT токена.
// Читает токен из заголовка Authorization: Bearer <token>.
// Если токен валидный — кладёт userID в c.Locals("userID").
// Если токен невалидный или отсутствует — возвращает 401.
func Protected(jwtSecret string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Читаем токен из заголовка
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "токен не предоставлен",
            })
        }
        
        // Убираем префикс "Bearer "
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "неверный формат токена",
            })
        }
        
        // Парсим и валидируем токен
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "невалидный токен",
            })
        }
        
        // Извлекаем claims (userID и role)
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "неверный формат токена",
            })
        }
        
        userID, ok := claims["userID"].(string)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "токен не содержит userID",
            })
        }
        
        // Кладём userID и role в контекст
        c.Locals("userID", userID)
        c.Locals("role", claims["role"])
        
        return c.Next()
    }
}
```

**Как это читается:**
1. Читаем заголовок `Authorization: Bearer <token>`.
2. Убираем префикс `"Bearer "`.
3. Парсим JWT токен с секретом из конфига.
4. Если токен валидный → извлекаем `userID` и `role` из claims.
5. Кладём их в `c.Locals("userID")` и `c.Locals("role")`.
6. Вызываем `c.Next()` — передаём управление следующему хендлеру.
7. Если на любом этапе ошибка → возвращаем `401 Unauthorized`.

**Важно:** middleware должен быть зарегистрирован в `routes.go` на группу `/api/account`:

```go
// В routes.go
func accountRoutes(app *fiber.App, jwtSecret string) {
    account := app.Group("/api/account")
    account.Use(middleware.Protected(jwtSecret)) // ← middleware здесь
    
    account.Get("/me", handlers.GetAccountMe)
    account.Get("/orders", handlers.GetOrders)
    account.Patch("/me", handlers.UpdateProfile)
}
```

### 2. Реализация методов репозитория

**Хорошая новость:** все методы уже реализованы в `postgres_account.go`! Проверь что они работают так:

#### `GetUserByID` — выборка пользователя по ID

```go
func (r *AccountSQLRepo) GetUserByID(id string) (*models.User, error) {
    query := `SELECT id, email, name, password_hash, role 
              FROM users WHERE id = $1`
    
    var u models.User
    err := r.db.QueryRow(query, id).Scan(
        &u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role,
    )
    if err != nil {
        return nil, fmt.Errorf("account.GetUserByID: %w", err)
    }
    return &u, nil
}
```

**Как это читается:**
- Делаем `SELECT` одной строки по `id`.
- Сканируем все поля в структуру `models.User`.
- Если не найдено → вернётся `sql.ErrNoRows`.

#### `ListOrdersByUser` — выборка заказов пользователя с позициями

**Это самый сложный метод**, потому что нужно:
1. Получить все заказы пользователя из таблицы `orders`.
2. Для каждого заказа подгрузить позиции из таблицы `order_items`.

```go
func (r *AccountSQLRepo) ListOrdersByUser(id string) ([]models.Order, error) {
    // Этап 1: Получаем все заказы
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
        err := rows.Scan(
            &o.ID, &o.UserID, &o.Status, &o.Total,
            &o.CreatedAt, &o.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("account.ListOrdersByUser scan order: %w", err)
        }
        o.Items = []models.OrderItem{} // инициализируем пустой слайс
        orders = append(orders, o)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("account.ListOrdersByUser rows: %w", err)
    }
    
    // Этап 2: Для каждого заказа подгружаем позиции
    itemsQuery := `SELECT id, order_id, product_id, title, price, quantity
                   FROM order_items WHERE order_id = $1`
    
    for i := range orders {
        itemRows, err := r.db.Query(itemsQuery, orders[i].ID)
        if err != nil {
            return nil, fmt.Errorf("account.ListOrdersByUser query items: %w", err)
        }
        
        for itemRows.Next() {
            var item models.OrderItem
            err := itemRows.Scan(
                &item.ID, &item.OrderID, &item.ProductID,
                &item.Title, &item.Price, &item.Quantity,
            )
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
```

**Как это читается:**
1. **Этап 1:** делаем `SELECT` всех заказов пользователя, сортируем по `created_at DESC` (новые сверху).
2. Сканируем каждый заказ в `models.Order`, инициализируем `Items` как пустой слайс.
3. **Этап 2:** для каждого заказа делаем отдельный `SELECT` позиций из `order_items`.
4. Сканируем каждую позицию в `models.OrderItem` и добавляем в `orders[i].Items`.
5. Возвращаем полный список заказов с позициями.

**Альтернатива (более оптимально):** можно сделать один запрос с `JOIN`, а потом группировать в Go:

```sql
SELECT 
    o.id, o.user_id, o.status, o.total, o.created_at, o.updated_at,
    oi.id, oi.order_id, oi.product_id, oi.title, oi.price, oi.quantity
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
WHERE o.user_id = $1
ORDER BY o.created_at DESC, oi.id
```

Но для старта двухэтапный вариант проще и понятнее.

#### `UpdateUser` — частичное обновление профиля

**Уже реализован** в `postgres_account.go`. Работает через динамический SQL:

```go
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
        return r.GetUserByID(id) // ничего не обновляем
    }
    
    query := fmt.Sprintf(
        "UPDATE users SET %s WHERE id = $%d RETURNING id, email, name, password_hash, role",
        strings.Join(setClauses, ", "), argIdx,
    )
    args = append(args, id)
    
    var u models.User
    err := r.db.QueryRow(query, args...).Scan(
        &u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role,
    )
    if err != nil {
        return nil, fmt.Errorf("account.UpdateUser: %w", err)
    }
    
    return &u, nil
}
```

**Как это читается:**
- Если `req.Name != nil` → добавляем `"name = $1"` в SET.
- Если `req.Email != nil` → добавляем `"email = $2"` в SET.
- Если ничего не прислали → просто возвращаем текущего пользователя.
- Выполняем `UPDATE ... RETURNING ...` и сканируем обновлённые данные.

### 3. Проверить что хендлеры правильно обрабатывают ошибки

**Хендлеры уже реализованы** в `account.go`, но убедись что они работают так:

```go
func GetAccountMe(c *fiber.Ctx) error {
    userID, _ := c.Locals("userID").(string)
    
    user, err := Repo.Account.GetUserByID(userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "ошибка при получении профиля",
        })
    }
    if user == nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "пользователь не найден",
        })
    }
    
    return c.JSON(fiber.Map{"user": user})
}
```

**Важно:** хендлеры должны проверять что `userID` не пустой (на случай если middleware не сработал):

```go
userID, ok := c.Locals("userID").(string)
if !ok || userID == "" {
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
        "error": "не авторизован",
    })
}
```

### 4. SQL‑схема для таблиц `users`, `orders`, `order_items`

**Если ещё не создал таблицы, вот пример схемы:**

```sql
-- Таблица пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user', -- 'user' или 'admin'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица заказов
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending | paid | shipped | delivered | cancelled
    total BIGINT NOT NULL, -- сумма в копейках (4990 = 49.90 ₽)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица позиций заказа
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL, -- ссылка на products.id (можно добавить FOREIGN KEY)
    title VARCHAR(255) NOT NULL, -- название товара НА МОМЕНТ покупки
    price BIGINT NOT NULL, -- цена за 1 шт. на момент покупки (в копейках)
    quantity INTEGER NOT NULL DEFAULT 1
);

-- Индексы для производительности
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
```

### 5. Дополнительно (опционально)

- **Валидация email:** перед обновлением профиля можно проверить формат email:
  ```go
  if req.Email != nil {
      if !isValidEmail(*req.Email) {
          return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
              "error": "невалидный email",
          })
      }
  }
  ```

- **Логирование:** можно добавить логирование действий пользователя (кто и когда обновил профиль, посмотрел заказы).

- **Пагинация для заказов:** если заказов много, можно добавить пагинацию в `GetOrders` (аналогично `GetProducts`).

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


