-- Лайки постов стены
CREATE TABLE wall_post_likes (
                                 post_id    UUID REFERENCES wall_posts(id) ON DELETE CASCADE,
                                 user_id    UUID REFERENCES users(id)      ON DELETE CASCADE,
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                 PRIMARY KEY (post_id, user_id)
);

-- Колонка для привязки чата к посту
ALTER TABLE wall_posts ADD COLUMN IF NOT EXISTS chat_id UUID REFERENCES chats(id) ON DELETE SET NULL;

CREATE INDEX idx_wall_post_likes_post ON wall_post_likes(post_id);
CREATE INDEX idx_wall_posts_chat_id   ON wall_posts(chat_id);