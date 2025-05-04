CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_by BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    planned_time FLOAT NOT NULL,
    actual_time FLOAT,
    created_at TIMESTAMP DEFAULT NOW(),
    tags INTEGER[]
);

CREATE TABLE IF NOT EXISTS tasks_groups (
    task_id BIGINT NOT NULL UNIQUE,
    group_id BIGINT NOT NULL,
    assigned_to BIGINT,

    CONSTRAINT fk_task_m2m_groups FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    task_id BIGINT NOT NULL,
    op VARCHAR(6) NOT NULL,
    processed BOOLEAN DEFAULT false,

    CONSTRAINT fk_task_outbox FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE,
    CHECK (op IN ('delete', 'update'))
)