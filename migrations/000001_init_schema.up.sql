------------------------------------------------------------------------ Users Table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username varchar(50) UNIQUE,
    password varchar(50)
);

------------------------------------------------------------------------ Tasks Table

CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    planned_time FLOAT NOT NULL,
    actual_time FLOAT,
    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_user_task FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_user ON tasks(user_id);

------------------------------------------------------------------------ Groups Tables

CREATE TABLE IF NOT EXISTS groups (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT fk_group_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(10) DEFAULT 'member',
    
    PRIMARY KEY (group_id, user_id),
    CONSTRAINT no_group_member_role CHECK (role IN ('admin', 'editor', 'member', 'viewer')),
    
    CONSTRAINT fk_member_group FOREIGN KEY (group_id) REFERENCES groups(id)  ON DELETE CASCADE,
    CONSTRAINT fk_member_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_group_members_user ON group_members(user_id);

------------------------------------------------------------------------ Graphs Tables

CREATE TABLE IF NOT EXISTS graphs (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name varchar(128) NOT NULL,
    group_id BIGINT NOT NULL,
    created_by BIGINT NOT NULL,
    
    CONSTRAINT fk_graph_group FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
    CONSTRAINT fk_graph_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS nodes (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    graph_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,
    assigned_to BIGINT,

    CONSTRAINT fk_graph_node FOREIGN KEY (graph_id) REFERENCES graphs (id) ON DELETE CASCADE,
    CONSTRAINT fk_task_node FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dependencies (
    from_node_id BIGINT NOT NULL,
    to_node_id BIGINT NOT NULL,
    graph_id BIGINT NOT NULL,

    PRIMARY KEY (to_node_id, graph_id),

    CONSTRAINT fk_graph_dependency FOREIGN KEY (graph_id) REFERENCES graphs (id) ON DELETE CASCADE,
    CONSTRAINT fk_from_node_dependency FOREIGN KEY (from_node_id) REFERENCES nodes (id) ON DELETE CASCADE,
    CONSTRAINT fk_to_node_dependency FOREIGN KEY (to_node_id) REFERENCES nodes (id) ON DELETE CASCADE,

    CONSTRAINT no_self_dependency CHECK (from_node_id <> to_node_id)
);

CREATE INDEX idx_dependencies_pair ON dependencies(from_node_id, to_node_id);

------------------------------------------------------------------------ Models Table

CREATE TABLE IF NOT EXISTS models (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    model BYTEA NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT false,
    
    CONSTRAINT fk_user_models FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_id_models ON models(user_id);

-- функция для обновления поля model_updated_at
CREATE OR REPLACE FUNCTION upd_model_updated_at() 
RETURNS TRIGGER 
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- триггер, который срабатывает на обновление поля model
CREATE TRIGGER update_model_updated_at
BEFORE UPDATE OF model ON models
FOR EACH ROW 
EXECUTE PROCEDURE upd_model_updated_at();