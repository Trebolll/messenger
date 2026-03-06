-- Откатываем: удаляем чаты созданные для постов стены
DELETE FROM chats
WHERE type = 'public'
  AND id IN (SELECT chat_id FROM wall_posts WHERE chat_id IS NOT NULL);

UPDATE wall_posts SET chat_id = NULL;

ALTER TABLE chats DROP CONSTRAINT IF EXISTS chats_type_check;
ALTER TABLE chats ADD CONSTRAINT chats_type_check CHECK (type IN ('private', 'group'));
