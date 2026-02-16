-- Таблица товаров
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    price BIGINT NOT NULL, -- цена в копейках (4990 = 49.90 ₽)
    currency VARCHAR(10) DEFAULT 'RUB',
    images JSONB DEFAULT '[]'::jsonb, -- массив URL изображений
    is_new BOOLEAN DEFAULT false,
    is_on_sale BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для товаров
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_is_new ON products(is_new);
CREATE INDEX idx_products_is_on_sale ON products(is_on_sale);

-- Таблица элементов галереи
CREATE TABLE gallery_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(50) NOT NULL, -- intro, tattoo, tokyo, ...
    title VARCHAR(255),
    image VARCHAR(500) NOT NULL, -- URL изображения
    sort_order INTEGER DEFAULT 0, -- порядок сортировки
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для галереи
CREATE INDEX idx_gallery_items_category ON gallery_items(category);
CREATE INDEX idx_gallery_items_sort_order ON gallery_items(sort_order);

-- Таблица статических страниц
CREATE TABLE pages (
    slug VARCHAR(50) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

-- Начальные данные
INSERT INTO pages (slug, title, content) VALUES
('payment', 'Оплата', 'Здесь будет текст про оплату...'),
('delivery', 'Доставка', 'Здесь будет текст про доставку...'),
('returns', 'Возврат', 'Здесь будет текст про возврат...'),
('contacts', 'Контакты', 'Здесь будет текст про контакты...');