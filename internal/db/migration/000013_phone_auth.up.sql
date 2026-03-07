-- Снимаем NOT NULL с email и password — для регистрации по телефону
ALTER TABLE users ALTER COLUMN email    DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Уникальный индекс на phone (частичный — только для непустых значений)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL AND phone != '';

-- Уникальный индекс на email тоже делаем частичным — чтобы NULL не конфликтовали
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL AND email != '';
