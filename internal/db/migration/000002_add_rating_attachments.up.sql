-- Добавляем колонки которых не было в старой схеме
ALTER TABLE users    ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE users    ADD COLUMN IF NOT EXISTS rating     INT NOT NULL DEFAULT 0;
ALTER TABLE chats    ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE chats    ADD COLUMN IF NOT EXISTS creator_id UUID REFERENCES users(id);
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited_at  TIMESTAMP;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS likes      INT NOT NULL DEFAULT 0;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS dislikes   INT NOT NULL DEFAULT 0;

-- Заполняем creator_id для существующих групп
UPDATE chats SET creator_id = (
    SELECT user_id FROM chat_members
    WHERE chat_id = chats.id
    ORDER BY joined_at ASC LIMIT 1
) WHERE creator_id IS NULL AND type = 'group';

-- Создаём таблицы которых не было
CREATE TABLE IF NOT EXISTS attachments (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id    UUID REFERENCES chats(id),
    sender_id  UUID REFERENCES users(id),
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    filename   TEXT,
    mime_type  VARCHAR(100),
    size_bytes BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS message_votes (
    message_id UUID     REFERENCES messages(id) ON DELETE CASCADE,
    voter_id   UUID     REFERENCES users(id)    ON DELETE CASCADE,
    vote       SMALLINT NOT NULL CHECK (vote IN (1, -1)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, voter_id)
);

CREATE INDEX IF NOT EXISTS idx_message_votes_message ON message_votes(message_id);
CREATE INDEX IF NOT EXISTS idx_message_votes_voter   ON message_votes(voter_id);
