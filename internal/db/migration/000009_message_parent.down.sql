DROP INDEX IF EXISTS idx_messages_parent_id;
ALTER TABLE messages DROP COLUMN IF EXISTS parent_id;
