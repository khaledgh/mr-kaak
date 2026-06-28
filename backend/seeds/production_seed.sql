-- Seed: Arabic Sweets & Kaak catalog + Super Admin User
-- Run: mysql -u <user> -p <database> < seeds/production_seed.sql
-- Safe to re-run: uses INSERT IGNORE / ON DUPLICATE KEY

-- ── 1. Super Admin User ───────────────────────────────────────────────────────
-- Default Credentials:
-- Email: admin@mrkaak.com
-- Password: adminpassword123

INSERT INTO users (name, email, phone_e164, password_hash, role, status, token_version, created_at, updated_at)
VALUES (
  'Admin MrKaak',
  'admin@mrkaak.com',
  '+16045550100',
  '$2a$12$CrKpHcTh5DcIvDfJ5loLhusPnIhdCeOSuErbz.p4dHtnlNXalffA.', -- bcrypt hash for: adminpassword123
  'super_admin',
  'active',
  0,
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE 
  role = 'super_admin',
  status = 'active';

-- ── 2. Categories ────────────────────────────────────────────────────────────

INSERT INTO categories (slug, sort_order, is_active) VALUES
  ('kaak',           1, 1),
  ('arabic-sweets',  2, 1),
  ('baklava',        3, 1),
  ('drinks',         4, 1)
-- If category already exists, update its sort order
ON DUPLICATE KEY UPDATE sort_order = VALUES(sort_order);

-- ── 3. Category translations ───────────────────────────────────────────────────

INSERT INTO translations (entity_type, entity_id, locale, field, value)
SELECT 'category', id, 'en', 'name',        'Kaak'           FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT 'category', id, 'ar', 'name',        'كعك'            FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT 'category', id, 'en', 'description', 'Traditional ring-shaped cookies' FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT 'category', id, 'ar', 'description', 'كعك تقليدي بأشكال متنوعة'     FROM categories WHERE slug = 'kaak'

UNION ALL
SELECT 'category', id, 'en', 'name',        'Arabic Sweets'  FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT 'category', id, 'ar', 'name',        'حلويات عربية'   FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT 'category', id, 'en', 'description', 'Handcrafted Lebanese sweets' FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT 'category', id, 'ar', 'description', 'حلويات لبنانية مصنوعة يدوياً' FROM categories WHERE slug = 'arabic-sweets'

UNION ALL
SELECT 'category', id, 'en', 'name',        'Baklava'        FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT 'category', id, 'ar', 'name',        'بقلاوة'         FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT 'category', id, 'en', 'description', 'Crispy layers of pastry and nuts' FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT 'category', id, 'ar', 'description', 'طبقات هشة من العجين والمكسرات' FROM categories WHERE slug = 'baklava'

UNION ALL
SELECT 'category', id, 'en', 'name',        'Drinks'         FROM categories WHERE slug = 'drinks'
UNION ALL
SELECT 'category', id, 'ar', 'name',        'مشروبات'        FROM categories WHERE slug = 'drinks'
ON DUPLICATE KEY UPDATE value = VALUES(value);

-- ── 4. Products ────────────────────────────────────────────────────────────────

INSERT INTO products (category_id, slug, pricing_mode, base_price_cents, is_available, sort_order)
SELECT id, 'kaak-plain',       'unit', 150, 1, 1 FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT id, 'kaak-sesame',      'unit', 175, 1, 2 FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT id, 'kaak-anise',       'unit', 175, 1, 3 FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT id, 'kaak-nut',         'unit', 250, 1, 4 FROM categories WHERE slug = 'kaak'
UNION ALL
SELECT id, 'maamoul-date',     'unit', 350, 1, 1 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'maamoul-pistachio','unit', 400, 1, 2 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'katayef-kashta',   'unit', 500, 1, 3 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'katayef-walnut',   'unit', 500, 1, 4 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'znoud-el-sit',     'unit', 450, 1, 5 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'halawet-el-jibn',  'weight', 1800, 1, 6 FROM categories WHERE slug = 'arabic-sweets'
UNION ALL
SELECT id, 'baklava-assorted', 'weight', 2200, 1, 1 FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT id, 'baklava-pistachio','weight', 2500, 1, 2 FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT id, 'baklava-cashew',   'weight', 2400, 1, 3 FROM categories WHERE slug = 'baklava'
UNION ALL
SELECT id, 'sahlab',           'unit', 500, 1, 1 FROM categories WHERE slug = 'drinks'
UNION ALL
SELECT id, 'jallab',           'unit', 450, 1, 2 FROM categories WHERE slug = 'drinks'
ON DUPLICATE KEY UPDATE base_price_cents = VALUES(base_price_cents);

-- ── 5. Product translations ────────────────────────────────────────────────────

INSERT INTO translations (entity_type, entity_id, locale, field, value)
-- Kaak plain
SELECT 'product', id, 'en', 'name',        'Plain Kaak'                   FROM products WHERE slug = 'kaak-plain'
UNION ALL SELECT 'product', id, 'ar', 'name',        'كعك سادة'            FROM products WHERE slug = 'kaak-plain'
UNION ALL SELECT 'product', id, 'en', 'description', 'Classic butter cookie rings, lightly sweetened' FROM products WHERE slug = 'kaak-plain'
UNION ALL SELECT 'product', id, 'ar', 'description', 'حلقات كعك زبدة كلاسيكية، محلاة قليلاً' FROM products WHERE slug = 'kaak-plain'
-- Kaak sesame
UNION ALL SELECT 'product', id, 'en', 'name',        'Sesame Kaak'                  FROM products WHERE slug = 'kaak-sesame'
UNION ALL SELECT 'product', id, 'ar', 'name',        'كعك سمسم'                     FROM products WHERE slug = 'kaak-sesame'
UNION ALL SELECT 'product', id, 'en', 'description', 'Crunchy ring cookies coated with toasted sesame' FROM products WHERE slug = 'kaak-sesame'
UNION ALL SELECT 'product', id, 'ar', 'description', 'كعك مقرمش مغطى بالسمسم المحمص' FROM products WHERE slug = 'kaak-sesame'
-- Kaak anise
UNION ALL SELECT 'product', id, 'en', 'name',        'Anise Kaak'                   FROM products WHERE slug = 'kaak-anise'
UNION ALL SELECT 'product', id, 'ar', 'name',        'كعك يانسون'                   FROM products WHERE slug = 'kaak-anise'
UNION ALL SELECT 'product', id, 'en', 'description', 'Fragrant cookies flavoured with anise and orange blossom' FROM products WHERE slug = 'kaak-anise'
UNION ALL SELECT 'product', id, 'ar', 'description', 'كعك عطري بنكهة اليانسون وماء الزهر' FROM products WHERE slug = 'kaak-anise'
-- Kaak nut
UNION ALL SELECT 'product', id, 'en', 'name',        'Nut-Stuffed Kaak'             FROM products WHERE slug = 'kaak-nut'
UNION ALL SELECT 'product', id, 'ar', 'name',        'كعك بالمكسرات'               FROM products WHERE slug = 'kaak-nut'
UNION ALL SELECT 'product', id, 'en', 'description', 'Butter cookies filled with spiced walnuts and cinnamon' FROM products WHERE slug = 'kaak-nut'
UNION ALL SELECT 'product', id, 'ar', 'description', 'كعك زبدة محشو بالجوز والقرفة' FROM products WHERE slug = 'kaak-nut'
-- Maamoul date
UNION ALL SELECT 'product', id, 'en', 'name',        'Date Maamoul'                 FROM products WHERE slug = 'maamoul-date'
UNION ALL SELECT 'product', id, 'ar', 'name',        'معمول تمر'                    FROM products WHERE slug = 'maamoul-date'
UNION ALL SELECT 'product', id, 'en', 'description', 'Semolina pastry filled with spiced date paste' FROM products WHERE slug = 'maamoul-date'
UNION ALL SELECT 'product', id, 'ar', 'description', 'معجنات سميد محشوة بعجينة التمر المتبلة' FROM products WHERE slug = 'maamoul-date'
-- Maamoul pistachio
UNION ALL SELECT 'product', id, 'en', 'name',        'Pistachio Maamoul'            FROM products WHERE slug = 'maamoul-pistachio'
UNION ALL SELECT 'product', id, 'ar', 'name',        'معمول فستق'                  FROM products WHERE slug = 'maamoul-pistachio'
UNION ALL SELECT 'product', id, 'en', 'description', 'Semolina pastry generously filled with ground pistachios' FROM products WHERE slug = 'maamoul-pistachio'
UNION ALL SELECT 'product', id, 'ar', 'description', 'معجنات سميد محشوة بالفستق المطحون' FROM products WHERE slug = 'maamoul-pistachio'
-- Katayef kashta
UNION ALL SELECT 'product', id, 'en', 'name',        'Katayef Kashta'               FROM products WHERE slug = 'katayef-kashta'
UNION ALL SELECT 'product', id, 'ar', 'name',        'قطايف قشطة'                  FROM products WHERE slug = 'katayef-kashta'
UNION ALL SELECT 'product', id, 'en', 'description', 'Fried pancakes stuffed with clotted cream and orange blossom syrup' FROM products WHERE slug = 'katayef-kashta'
UNION ALL SELECT 'product', id, 'ar', 'description', 'قطايف مقلية محشوة بالقشطة وشراب ماء الزهر' FROM products WHERE slug = 'katayef-kashta'
-- Katayef walnut
UNION ALL SELECT 'product', id, 'en', 'name',        'Katayef Walnut'               FROM products WHERE slug = 'katayef-walnut'
UNION ALL SELECT 'product', id, 'ar', 'name',        'قطايف بالجوز'                FROM products WHERE slug = 'katayef-walnut'
UNION ALL SELECT 'product', id, 'en', 'description', 'Fried pancakes stuffed with spiced ground walnuts' FROM products WHERE slug = 'katayef-walnut'
UNION ALL SELECT 'product', id, 'ar', 'description', 'قطايف مقلية محشوة بالجوز المطحون' FROM products WHERE slug = 'katayef-walnut'
-- Znoud el sit
UNION ALL SELECT 'product', id, 'en', 'name',        "Znoud El Sit"                 FROM products WHERE slug = 'znoud-el-sit'
UNION ALL SELECT 'product', id, 'ar', 'name',        'زنود الست'                   FROM products WHERE slug = 'znoud-el-sit'
UNION ALL SELECT 'product', id, 'en', 'description', 'Fried pastry rolls filled with kashta cream and soaked in syrup' FROM products WHERE slug = 'znoud-el-sit'
UNION ALL SELECT 'product', id, 'ar', 'description', 'لفائف عجين مقلية محشوة بالقشطة ومغموسة بالقطر' FROM products WHERE slug = 'znoud-el-sit'
-- Halawet el jibn
UNION ALL SELECT 'product', id, 'en', 'name',        "Halawet El Jibn"              FROM products WHERE slug = 'halawet-el-jibn'
UNION ALL SELECT 'product', id, 'ar', 'name',        'حلاوة الجبن'                 FROM products WHERE slug = 'halawet-el-jibn'
UNION ALL SELECT 'product', id, 'en', 'description', 'Stretchy semolina-cheese rolls filled with kashta — sold by weight' FROM products WHERE slug = 'halawet-el-jibn'
UNION ALL SELECT 'product', id, 'ar', 'description', 'لفائف السميد والجبنة المحشوة بالقشطة — بالوزن' FROM products WHERE slug = 'halawet-el-jibn'
-- Baklava assorted
UNION ALL SELECT 'product', id, 'en', 'name',        'Assorted Baklava'             FROM products WHERE slug = 'baklava-assorted'
UNION ALL SELECT 'product', id, 'ar', 'name',        'بقلاوة مشكلة'               FROM products WHERE slug = 'baklava-assorted'
UNION ALL SELECT 'product', id, 'en', 'description', 'A mix of crispy filo pastry, rose-water syrup, and mixed nuts — sold by weight' FROM products WHERE slug = 'baklava-assorted'
UNION ALL SELECT 'product', id, 'ar', 'description', 'مزيج من عجينة الفيلو الهشة وشراب ماء الورد والمكسرات — بالوزن' FROM products WHERE slug = 'baklava-assorted'
-- Baklava pistachio
UNION ALL SELECT 'product', id, 'en', 'name',        'Pistachio Baklava'            FROM products WHERE slug = 'baklava-pistachio'
UNION ALL SELECT 'product', id, 'ar', 'name',        'بقلاوة فستق'                FROM products WHERE slug = 'baklava-pistachio'
UNION ALL SELECT 'product', id, 'en', 'description', 'Fine filo layers packed with crushed pistachios and fragrant syrup — by weight' FROM products WHERE slug = 'baklava-pistachio'
UNION ALL SELECT 'product', id, 'ar', 'description', 'طبقات فيلو رفيعة بالفستق المطحون والقطر العطر — بالوزن' FROM products WHERE slug = 'baklava-pistachio'
-- Baklava cashew
UNION ALL SELECT 'product', id, 'en', 'name',        'Cashew Baklava'               FROM products WHERE slug = 'baklava-cashew'
UNION ALL SELECT 'product', id, 'ar', 'name',        'بقلاوة كاجو'                FROM products WHERE slug = 'baklava-cashew'
UNION ALL SELECT 'product', id, 'en', 'description', 'Buttery filo filled with whole cashews and honey syrup — by weight' FROM products WHERE slug = 'baklava-cashew'
UNION ALL SELECT 'product', id, 'ar', 'description', 'عجينة فيلو زبدية بالكاجو الكامل وشراب العسل — بالوزن' FROM products WHERE slug = 'baklava-cashew'
-- Sahlab
UNION ALL SELECT 'product', id, 'en', 'name',        'Sahlab'                       FROM products WHERE slug = 'sahlab'
UNION ALL SELECT 'product', id, 'ar', 'name',        'سحلب'                        FROM products WHERE slug = 'sahlab'
UNION ALL SELECT 'product', id, 'en', 'description', 'Warm creamy orchid-root drink topped with cinnamon and coconut' FROM products WHERE slug = 'sahlab'
UNION ALL SELECT 'product', id, 'ar', 'description', 'مشروب السحلب الدافئ مع القرفة وجوز الهند' FROM products WHERE slug = 'sahlab'
-- Jallab
UNION ALL SELECT 'product', id, 'en', 'name',        'Jallab'                       FROM products WHERE slug = 'jallab'
UNION ALL SELECT 'product', id, 'ar', 'name',        'جلاب'                        FROM products WHERE slug = 'jallab'
UNION ALL SELECT 'product', id, 'en', 'description', 'Sweet rose-grape drink with pine nuts and raisins' FROM products WHERE slug = 'jallab'
UNION ALL SELECT 'product', id, 'ar', 'description', 'مشروب الجلاب الحلو بالصنوبر والزبيب' FROM products WHERE slug = 'jallab'
ON DUPLICATE KEY UPDATE value = VALUES(value);

-- ── 6. Variants: Kaak sold by box size ────────────────────────────────────────

INSERT INTO product_variants (product_id, label, price_cents, is_default, sort_order)
SELECT id, 'Small box (250g)',  350, 1, 1 FROM products WHERE slug = 'kaak-plain'
UNION ALL SELECT id, 'Large box (500g)',  650, 0, 2 FROM products WHERE slug = 'kaak-plain'
UNION ALL SELECT id, 'Small box (250g)',  400, 1, 1 FROM products WHERE slug = 'kaak-sesame'
UNION ALL SELECT id, 'Large box (500g)',  750, 0, 2 FROM products WHERE slug = 'kaak-sesame'
UNION ALL SELECT id, 'Small box (250g)',  400, 1, 1 FROM products WHERE slug = 'kaak-anise'
UNION ALL SELECT id, 'Large box (500g)',  750, 0, 2 FROM products WHERE slug = 'kaak-anise'
UNION ALL SELECT id, 'Box of 6',   900, 1, 1 FROM products WHERE slug = 'maamoul-date'
UNION ALL SELECT id, 'Box of 12', 1700, 0, 2 FROM products WHERE slug = 'maamoul-date'
UNION ALL SELECT id, 'Box of 6',  1100, 1, 1 FROM products WHERE slug = 'maamoul-pistachio'
UNION ALL SELECT id, 'Box of 12', 2100, 0, 2 FROM products WHERE slug = 'maamoul-pistachio'
ON DUPLICATE KEY UPDATE price_cents = VALUES(price_cents);
