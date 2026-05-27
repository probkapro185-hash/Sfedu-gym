-- 002_seed_admin.down.sql
DELETE FROM users WHERE email = 'admin@gmail.com' AND role = 'admin';
