package routes

import (
	"github.com/gofiber/fiber/v2"

	"socialsh/backend/internal/handlers"
	"socialsh/backend/internal/middleware"
)

// Register — единая точка входа для регистрации ВСЕХ маршрутов приложения.
// Принимает Fiber-приложение и jwtSecret (нужен для middleware авторизации).
// Внутри создаёт группу /api и раскидывает роуты по уровням доступа.
func Register(app *fiber.App, jwtSecret, refreshSecret string) {
	// Все эндпоинты живут под /api — фронт ходит на /api/products, /api/auth/sign-in и т.д.
	api := app.Group("/api")

	// ──── 1. Публичные роуты (без авторизации) ────

	// Магазин — список товаров с фильтрацией через query-параметры
	api.Get("/products", handlers.GetProducts)      // GET /api/products?new=true&sale=true&page=1&limit=20
	api.Get("/products/:slug", handlers.GetProduct) // GET /api/products/hoodie-black → один товар по slug

	// Галерея — фотки с фильтром по категории
	api.Get("/gallery", handlers.GetGalleryItems) // GET /api/gallery?category=intro

	// Инфо-страницы — оплата, доставка, возврат, контакты
	api.Get("/pages/:slug", handlers.GetPage) // GET /api/pages/payment | delivery | returns | contacts

	// ──── 2. Auth-роуты (регистрация/логин/рефреш) ────
	authRoutes(api, jwtSecret, refreshSecret)

	// ──── 3. Личный кабинет (нужна авторизация) ────
	accountRoutes(api, jwtSecret)

	// ──── 4. Админка (авторизация + роль admin) ────
	adminRoutes(api, jwtSecret)
}

// authRoutes — группа маршрутов аутентификации.
// sign-up и sign-in — публичные (токена ещё нет).
// logout и is-admin — защищённые (нужен валидный JWT).
func authRoutes(api fiber.Router, jwtSecret, refreshSecret string) {
	a := api.Group("/auth")

	// sign-up/sign-in/refresh — публичные (токена ещё нет)
	a.Post("/sign-up", handlers.SignUp(jwtSecret, refreshSecret))       // регистрация → {access, refresh}
	a.Post("/sign-in", handlers.SignIn(jwtSecret, refreshSecret))       // логин → {access, refresh}
	a.Post("/refresh", handlers.RefreshToken(jwtSecret, refreshSecret)) // обновить access по refresh

	// logout и is-admin — защищённые (нужен валидный JWT)
	a.Post("/logout", middleware.Protected(jwtSecret), handlers.Logout)   // инвалидировать refresh
	a.Get("/is-admin", middleware.Protected(jwtSecret), handlers.IsAdmin) // {isAdmin: true/false}
}

// accountRoutes — личный кабинет пользователя.
// Вся группа /account защищена middleware.Protected — без токена сюда не попасть.
func accountRoutes(api fiber.Router, jwtSecret string) {
	acc := api.Group("/account", middleware.Protected(jwtSecret))

	acc.Get("/me", handlers.GetAccountMe)    // GET /api/account/me → профиль текущего юзера
	acc.Get("/orders", handlers.GetOrders)   // GET /api/account/orders → список заказов юзера
	acc.Patch("/me", handlers.UpdateProfile) // PATCH /api/account/me → обновить имя/email/и т.д.
}

// adminRoutes — админская панель, полный CRUD для контента.
// Два middleware подряд: Protected (проверка JWT) → AdminOnly (проверка role == "admin").
// Если юзер не админ — получит 403 Forbidden.
func adminRoutes(api fiber.Router, jwtSecret string) {
	adm := api.Group("/admin",
		middleware.Protected(jwtSecret), // сначала проверяем что JWT валидный
		middleware.AdminOnly(),          // потом что role == "admin"
	)

	// ── Товары (полный CRUD) ──
	adm.Get("/products", handlers.AdminListProducts)         // список всех товаров для админки
	adm.Post("/products", handlers.AdminCreateProduct)       // создать новый товар
	adm.Get("/products/:id", handlers.AdminGetProduct)       // один товар по ID (не slug!)
	adm.Patch("/products/:id", handlers.AdminUpdateProduct)  // обновить поля товара
	adm.Delete("/products/:id", handlers.AdminDeleteProduct) // удалить товар

	// ── Галерея ──
	adm.Get("/gallery", handlers.AdminListGalleryItems)         // список элементов галереи
	adm.Post("/gallery", handlers.AdminCreateGalleryItem)       // добавить фото в галерею
	adm.Patch("/gallery/:id", handlers.AdminUpdateGalleryItem)  // изменить элемент
	adm.Delete("/gallery/:id", handlers.AdminDeleteGalleryItem) // удалить элемент

	// ── Инфо-страницы ──
	adm.Get("/pages", handlers.AdminListPages)          // список всех страниц
	adm.Patch("/pages/:slug", handlers.AdminUpdatePage) // обновить контент страницы
}
