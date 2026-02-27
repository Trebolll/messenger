-- Создание таблицы пользователей

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
                                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Используем UUID v4
                                     username VARCHAR(50) UNIQUE NOT NULL,
                                     email VARCHAR(100) UNIQUE NOT NULL,
                                     password TEXT NOT NULL,
                                     phone VARCHAR(20),
                                     full_name VARCHAR(255),
                                     birth_date DATE,
                                     location VARCHAR(255),
                                     status VARCHAR(255),
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP
);

-- Создание таблицы чатов
CREATE TABLE IF NOT EXISTS chats (
                                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Используем UUID v4
                                     type VARCHAR(10) NOT NULL CHECK (type IN ('private', 'group')),
                                     name VARCHAR(100),
                                     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы участников чатов
CREATE TABLE IF NOT EXISTS chat_members (
                                            chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
                                            user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                            joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                            PRIMARY KEY (chat_id, user_id)
);

-- Создание таблицы сообщений
CREATE TABLE IF NOT EXISTS messages (
                                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
                                        sender_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                        content TEXT NOT NULL,
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                        read_at TIMESTAMP,
                                        edited_at TIMESTAMP
);

CREATE TABLE  IF NOT EXISTS attachments (
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

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_chat_members_user_id ON chat_members (user_id);
CREATE INDEX IF NOT EXISTS idx_messages_chat_id_created_at ON messages (chat_id, created_at ASC);
-- Добавляем edited_at если ещё нет (для существующих БД)
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited_at TIMESTAMP;
-- Аватар пользователя
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;
-- Аватар группового чата
ALTER TABLE chats ADD COLUMN IF NOT EXISTS avatar_url TEXT;