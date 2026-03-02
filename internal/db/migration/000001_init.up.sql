CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(50)  UNIQUE NOT NULL,
    email      VARCHAR(100) UNIQUE NOT NULL,
    password   TEXT         NOT NULL,
    phone      VARCHAR(20),
    full_name  VARCHAR(255),
    birth_date DATE,
    location   VARCHAR(255),
    status     VARCHAR(255),
    avatar_url TEXT,
    rating     INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE chats (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type       VARCHAR(10) NOT NULL CHECK (type IN ('private', 'group')),
    name       VARCHAR(100),
    avatar_url TEXT,
    creator_id UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_members (
    chat_id   UUID REFERENCES chats(id)   ON DELETE CASCADE,
    user_id   UUID REFERENCES users(id)   ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id    UUID REFERENCES chats(id)  ON DELETE CASCADE,
    sender_id  UUID REFERENCES users(id)  ON DELETE CASCADE,
    content    TEXT NOT NULL,
    likes      INT NOT NULL DEFAULT 0,
    dislikes   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at    TIMESTAMP,
    edited_at  TIMESTAMP
);

CREATE TABLE attachments (
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

CREATE TABLE message_votes (
    message_id UUID     REFERENCES messages(id) ON DELETE CASCADE,
    voter_id   UUID     REFERENCES users(id)    ON DELETE CASCADE,
    vote       SMALLINT NOT NULL CHECK (vote IN (1, -1)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, voter_id)
);

CREATE INDEX idx_users_username              ON users       (username);
CREATE INDEX idx_users_email                 ON users       (email);
CREATE INDEX idx_chat_members_user_id        ON chat_members(user_id);
CREATE INDEX idx_messages_chat_id_created_at ON messages    (chat_id, created_at ASC);
CREATE INDEX idx_message_votes_message       ON message_votes(message_id);
CREATE INDEX idx_message_votes_voter         ON message_votes(voter_id);
