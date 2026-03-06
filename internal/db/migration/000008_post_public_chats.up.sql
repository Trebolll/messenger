-- Добавляем тип 'public' в допустимые типы чатов
ALTER TABLE chats DROP CONSTRAINT IF EXISTS chats_type_check;
ALTER TABLE chats ADD CONSTRAINT chats_type_check CHECK (type IN ('private', 'group', 'public'));

-- Создаём публичные чаты для всех существующих постов без чата
DO $$
DECLARE
    r RECORD;
    new_chat_id UUID;
BEGIN
    FOR r IN
        SELECT wp.id AS post_id, wp.user_id, wp.content
        FROM wall_posts wp
        WHERE wp.chat_id IS NULL
    LOOP
        INSERT INTO chats(type, name, creator_id)
        VALUES ('public', NULLIF(r.content, ''), r.user_id)
        RETURNING id INTO new_chat_id;

        INSERT INTO chat_members(chat_id, user_id)
        VALUES (new_chat_id, r.user_id);

        UPDATE wall_posts SET chat_id = new_chat_id WHERE id = r.post_id;
    END LOOP;
END;
$$;
