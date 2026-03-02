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
