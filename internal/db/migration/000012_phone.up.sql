-- Добавляем уникальный индекс на phone (если колонка уже есть в users)
-- Если phone ещё нет — добавляем колонку
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL;
