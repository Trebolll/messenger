CREATE TABLE walls (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    bio        TEXT,
    banner_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем записи для всех существующих пользователей
INSERT INTO walls (user_id)
SELECT id FROM users
ON CONFLICT (user_id) DO NOTHING;

CREATE INDEX idx_walls_user_id ON walls(user_id);
