CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    planned_time FLOAT NOT NULL,    -- Планируемое время выполнения (в часах)
    actual_time FLOAT,              -- Фактическое время выполнения (в часах)
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_user_task FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- CREATE TABLE outbox (
--     id BIGSERIAL PRIMARY KEY,
--     user_id BIGINT NOT NULL,
--     event_type VARCHAR(50) NOT NULL,
--     payload JSONB NOT NULL,
--     created_at TIMESTAMP DEFAULT NOW(),
--     processed BOOLEAN DEFAULT FALSE,

--     CONSTRAINT fk_user_outbox FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
-- );
