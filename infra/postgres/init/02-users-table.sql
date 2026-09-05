CREATE TABLE IF NOT EXISTS users.users (
    id          uuid PRIMARY KEY,
    status      varchar(32) NOT NULL DEFAULT 'ACTIVE',
    created_at  timestamp NOT NULL DEFAULT NOW(),
    updated_at  timestamp NOT NULL DEFAULT NOW(),
    create_user varchar(255) NOT NULL DEFAULT 'system',
    update_user varchar(255) NOT NULL DEFAULT 'system'
);
