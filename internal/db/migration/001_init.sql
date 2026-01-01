-- Создание таблицы пользователей

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Используем UUID v4
username VARCHAR(50) UNIQUE NOT NULL,
email VARCHAR(100) UNIQUE NOT NULL,
password TEXT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
read_at TIMESTAMP
);