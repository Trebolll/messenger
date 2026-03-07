ALTER TABLE users ALTER COLUMN email    SET NOT NULL;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;

DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX idx_users_email ON users (email);
