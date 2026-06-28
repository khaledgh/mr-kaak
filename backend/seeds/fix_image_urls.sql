-- SQL Script to fix existing image URLs by replacing localhost and incorrect domains with the production API domain
-- Run this on your server's MySQL database.

-- 1. Update Categories
UPDATE categories 
SET image_url = REPLACE(image_url, 'http://127.0.0.1:8080', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'http://127.0.0.1:8080%';

UPDATE categories 
SET image_url = REPLACE(image_url, 'https://mr-kaak-api.linksbridge.top', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'https://mr-kaak-api.linksbridge.top%';

-- 2. Update Products
UPDATE products 
SET image_url = REPLACE(image_url, 'http://127.0.0.1:8080', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'http://127.0.0.1:8080%';

UPDATE products 
SET image_url = REPLACE(image_url, 'https://mr-kaak-api.linksbridge.top', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'https://mr-kaak-api.linksbridge.top%';

-- 3. Update Banners
UPDATE banners 
SET image_url = REPLACE(image_url, 'http://127.0.0.1:8080', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'http://127.0.0.1:8080%';

UPDATE banners 
SET image_url = REPLACE(image_url, 'https://mr-kaak-api.linksbridge.top', 'https://mrkaak-api.linksbridge.top') 
WHERE image_url LIKE 'https://mr-kaak-api.linksbridge.top%';

-- 4. Update Media Library
UPDATE media 
SET url = REPLACE(url, 'http://127.0.0.1:8080', 'https://mrkaak-api.linksbridge.top'),
    thumb_url = REPLACE(thumb_url, 'http://127.0.0.1:8080', 'https://mrkaak-api.linksbridge.top') 
WHERE url LIKE 'http://127.0.0.1:8080%';

UPDATE media 
SET url = REPLACE(url, 'https://mr-kaak-api.linksbridge.top', 'https://mrkaak-api.linksbridge.top'),
    thumb_url = REPLACE(thumb_url, 'https://mr-kaak-api.linksbridge.top', 'https://mrkaak-api.linksbridge.top') 
WHERE url LIKE 'https://mr-kaak-api.linksbridge.top%';
