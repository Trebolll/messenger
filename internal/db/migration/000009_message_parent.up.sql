ALTER TABLE messages ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES messages(id) ON DELETE SET NULL;
CREATE INDEX idx_messages_parent_id ON messages(parent_id);
