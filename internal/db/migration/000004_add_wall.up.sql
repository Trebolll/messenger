CREATE TABLE wall_posts (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE wall_attachments (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id    UUID REFERENCES wall_posts(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    filename   TEXT,
    mime_type  VARCHAR(100),
    size_bytes BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wall_posts_user_id ON wall_posts(user_id);
CREATE INDEX idx_wall_posts_created_at ON wall_posts(created_at DESC);
