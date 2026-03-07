-- ─── Миграция: сущность Feed (метрики + рекомендательный алгоритм) ────────────

-- 1. Сырые события взаимодействия
CREATE TABLE IF NOT EXISTS feed_events (
                                           id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                                           user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                           post_id       UUID        NOT NULL REFERENCES wall_posts(id) ON DELETE CASCADE,
                                           event_type    VARCHAR(32) NOT NULL CHECK (event_type IN ('view','like','comment','video_complete','skip')),
                                           watch_seconds INT         NOT NULL DEFAULT 0,
                                           created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_events_user     ON feed_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feed_events_post     ON feed_events(post_id);

-- 2. Агрегированный скор — обновляется при каждом событии
CREATE TABLE IF NOT EXISTS feed_scores (
                                           post_id         UUID    NOT NULL REFERENCES wall_posts(id) ON DELETE CASCADE,
                                           user_id         UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                           score           FLOAT   NOT NULL DEFAULT 0,
                                           view_count      INT     NOT NULL DEFAULT 0,
                                           like_count      INT     NOT NULL DEFAULT 0,
                                           comment_count   INT     NOT NULL DEFAULT 0,
                                           total_watch_sec INT     NOT NULL DEFAULT 0,
                                           last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                           PRIMARY KEY (post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_feed_scores_user ON feed_scores(user_id, score DESC);

-- 3. Предпочтения пользователя по типу контента (image / video / text)
--    Хранит нормализованный вес [0..1] для каждого типа.
--    Используется алгоритмом для буста нужного типа.
CREATE TABLE IF NOT EXISTS feed_preferences (
                                                user_id         UUID    PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                                                weight_image    FLOAT   NOT NULL DEFAULT 0.33,
                                                weight_video    FLOAT   NOT NULL DEFAULT 0.33,
                                                weight_text     FLOAT   NOT NULL DEFAULT 0.34,
                                                updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Индексы для оптимизации запроса GetPersonalFeed ─────────────────────────

-- first_att CTE: DISTINCT ON (post_id) ORDER BY post_id, created_at
CREATE INDEX IF NOT EXISTS idx_wall_att_post_created
    ON wall_attachments(post_id, created_at);

-- likes_agg: GROUP BY post_id
CREATE INDEX IF NOT EXISTS idx_wall_likes_post
    ON wall_post_likes(post_id);

-- is_liked: BOOL_OR(user_id = $1)
CREATE INDEX IF NOT EXISTS idx_wall_likes_user
    ON wall_post_likes(user_id);

-- comments_agg: JOIN messages ON chat_id
CREATE INDEX IF NOT EXISTS idx_messages_chat
    ON messages(chat_id);

-- user_score: WHERE user_id = $1
-- (уже есть idx_feed_scores_user, но добавим covering index)
CREATE INDEX IF NOT EXISTS idx_feed_scores_user_post
    ON feed_scores(user_id, post_id) INCLUDE (view_count, last_seen_at);

-- freshness filter: посты сортируются по created_at внутри ORDER BY final_score
-- partial index только на посты с медиа не нужен — JOIN с first_att уже фильтрует
CREATE INDEX IF NOT EXISTS idx_wall_posts_created
    ON wall_posts(created_at DESC);