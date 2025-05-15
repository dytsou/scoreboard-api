CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS scoreboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- CREATE TABLE IF NOT EXISTS scoreboards_items (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     scoreboard_id UUID NOT NULL,
--     user_id UUID NOT NULL,
--     username VARCHAR(255) DEFAULT NULL,
--     score int DEFAULT 0,
--     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
--     updated_at TIMESTAMP NOT NULL DEFAULT NOW()
-- );