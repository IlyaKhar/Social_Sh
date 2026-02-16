-- Наполнение базы тестовыми данными

-- 1. Админа нужно создать через API: POST /api/auth/sign-up
--    {"email":"admin@socialsh.ru","password":"admin123","name":"Администратор"}
--    Затем обновить роль: UPDATE users SET role = 'admin' WHERE email = 'admin@socialsh.ru';
--    
--    Или создать вручную через psql после регистрации:
--    psql -d socialsh -c "UPDATE users SET role = 'admin' WHERE email = 'admin@socialsh.ru';"

-- 2. Добавляем тестовые товары
INSERT INTO products (id, slug, title, description, price, currency, images, is_new, is_on_sale) VALUES
('10000000-0000-0000-0000-000000000001', 'hoodie-black', 'Худи чёрное', 'Классическое чёрное худи из премиального хлопка. Удобный крой, капюшон с регулировкой.', 4990, 'RUB', '["/images/hoodie-black-1.jpg", "/images/hoodie-black-2.jpg"]'::jsonb, true, false),
('10000000-0000-0000-0000-000000000002', 'hoodie-white', 'Худи белое', 'Минималистичное белое худи. Идеально для повседневной носки.', 4990, 'RUB', '["/images/hoodie-white-1.jpg"]'::jsonb, true, false),
('10000000-0000-0000-0000-000000000003', 't-shirt-black', 'Футболка чёрная', 'Базовая чёрная футболка из органического хлопка. Экологично и стильно.', 1990, 'RUB', '["/images/tshirt-black-1.jpg"]'::jsonb, false, true),
('10000000-0000-0000-0000-000000000004', 't-shirt-white', 'Футболка белая', 'Классическая белая футболка. Универсальный базовый элемент гардероба.', 1990, 'RUB', '["/images/tshirt-white-1.jpg"]'::jsonb, false, true),
('10000000-0000-0000-0000-000000000005', 'cap-black', 'Кепка чёрная', 'Чёрная кепка с вышитым логотипом. Защита от солнца и стильный аксессуар.', 1490, 'RUB', '["/images/cap-black-1.jpg"]'::jsonb, false, false),
('10000000-0000-0000-0000-000000000006', 'sweatshirt-grey', 'Свитшот серый', 'Уютный серый свитшот. Идеален для прохладной погоды.', 3990, 'RUB', '["/images/sweatshirt-grey-1.jpg"]'::jsonb, true, false)
ON CONFLICT (slug) DO NOTHING;

-- 3. Добавляем элементы галереи
INSERT INTO gallery_items (id, category, title, image, sort_order) VALUES
('20000000-0000-0000-0000-000000000001', 'intro', 'Главное фото 1', '/images/gallery/intro-1.jpg', 1),
('20000000-0000-0000-0000-000000000002', 'intro', 'Главное фото 2', '/images/gallery/intro-2.jpg', 2),
('20000000-0000-0000-0000-000000000003', 'intro', 'Главное фото 3', '/images/gallery/intro-3.jpg', 3),
('20000000-0000-0000-0000-000000000004', 'tattoo', 'Тату 1', '/images/gallery/tattoo-1.jpg', 1),
('20000000-0000-0000-0000-000000000005', 'tattoo', 'Тату 2', '/images/gallery/tattoo-2.jpg', 2),
('20000000-0000-0000-0000-000000000006', 'tattoo', 'Тату 3', '/images/gallery/tattoo-3.jpg', 3),
('20000000-0000-0000-0000-000000000007', 'tokyo', 'Токио 1', '/images/gallery/tokyo-1.jpg', 1),
('20000000-0000-0000-0000-000000000008', 'tokyo', 'Токио 2', '/images/gallery/tokyo-2.jpg', 2)
ON CONFLICT DO NOTHING;

-- 4. Обновляем контент страниц (более реалистичный текст)
UPDATE pages SET content = 'Мы принимаем оплату банковскими картами Visa, MasterCard, МИР. Также доступна оплата через СБП (Система быстрых платежей). Все платежи защищены и обрабатываются через безопасные платёжные системы.' WHERE slug = 'payment';
UPDATE pages SET content = 'Доставка по России осуществляется через СДЭК и Почту России. Срок доставки: 3-7 рабочих дней. Стоимость доставки рассчитывается при оформлении заказа. Самовывоз из нашего офиса в Москве - бесплатно.' WHERE slug = 'delivery';
UPDATE pages SET content = 'Вы можете вернуть товар в течение 14 дней с момента покупки, если он не был в употреблении, сохранены товарный вид и упаковка. Возврат денежных средств осуществляется в течение 5-10 рабочих дней на ту же карту, с которой была оплата.' WHERE slug = 'returns';
UPDATE pages SET content = 'Email: info@socialsh.ru
Телефон: +7 (999) 123-45-67
Адрес: Москва, ул. Примерная, д. 1
Время работы: Пн-Пт 10:00-20:00, Сб-Вс 12:00-18:00' WHERE slug = 'contacts';
