-- 002_seed_admin.up.sql
-- Начальный администратор системы
-- Пароль: Admin@2024 (bcrypt hash cost=12, нужно сменить после первого входа)

INSERT INTO users (full_name, phone, email, password_hash, role, gender)
VALUES (
    'Администратор SFEDU',
    '+70000000000',
    'admin@gmail.com',
    -- Это хеш строки 'Admin@2024' с bcrypt cost=12
    -- Сгенерируйте реальный хеш перед деплоем!
    '$2a$12$obOWT7.qHQyoG2qLCUuvW.IybKvUfxobSFa3Ps2VdxtJIBJPmDVfy',
    'admin',
    'male'
);
