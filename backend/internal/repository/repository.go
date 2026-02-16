package repository

import (
	"database/sql"
	"socialsh/backend/internal/models"
)

// Здесь только интерфейсы, реализацию (Mongo/Postgres/файлы) выбираешь сам.

type ProductRepository interface {
	// Публичные
	List(newOnly, saleOnly bool, page, limit int) ([]models.Product, error)
	GetBySlug(slug string) (*models.Product, error)
	Search(query string, page, limit int) ([]models.Product, error) // поиск по названию
	// Админские
	ListAll() ([]models.Product, error)
	GetByID(id string) (*models.Product, error)
	Create(product *models.Product) error
	Update(id string, product *models.Product) (*models.Product, error)
	Delete(id string) error
}

type GalleryRepository interface {
	// Публичные
	ListByCategory(category string) ([]models.GalleryItem, error)
	// Админские
	ListAll() ([]models.GalleryItem, error)
	Create(item *models.GalleryItem) error
	Update(id string, item *models.GalleryItem) (*models.GalleryItem, error)
	Delete(id string) error
}

type PageRepository interface {
	// Публичные
	GetBySlug(slug string) (*models.Page, error)
	// Админские
	ListAll() ([]models.Page, error)
	Update(slug string, page *models.Page) (*models.Page, error)
}

type AccountRepository interface {
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(id string, req *models.UpdateProfileRequest) (*models.User, error)
	ListOrdersByUser(id string) ([]models.Order, error)
}

// Store агрегирует все репозитории, чтобы было удобно прокидывать зависимости.
type Store struct {
	Products ProductRepository
	Gallery  GalleryRepository
	Pages    PageRepository
	Account  AccountRepository
}

// TODO: сделай конструктор под свою реализацию, например:
// func NewStore(db *sql.DB) *Store { ... } или func NewStore(client *mongo.Client) *Store { ... }.

func NewStore(db *sql.DB) *Store {
	return &Store{
		Products: NewProductSQLRepo(db),
		Gallery:  NewGallerySQLRepo(db),
		Pages:    NewPageSQLRepo(db),
		Account:  NewAccountSQLRepo(db),
	}
}
