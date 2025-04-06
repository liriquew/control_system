CREATE TABLE IF NOT EXISTS graphs (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name varchar(128) NOT NULL,
    group_id BIGINT NOT NULL,
    created_by BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS nodes (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    graph_id BIGINT NOT NULL,
    task_id BIGINT NOT NULL,

    CONSTRAINT uq_task_id UNIQUE (task_id),
    CONSTRAINT fk_graph_node FOREIGN KEY (graph_id) REFERENCES graphs (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dependencies (
    from_node_id BIGINT NOT NULL,
    to_node_id BIGINT NOT NULL,
    graph_id BIGINT NOT NULL,

    PRIMARY KEY (from_node_id, to_node_id),

    CONSTRAINT fk_graph_dependency FOREIGN KEY (graph_id) REFERENCES graphs (id) ON DELETE CASCADE,
    CONSTRAINT fk_from_node_dependency FOREIGN KEY (from_node_id) REFERENCES nodes (id) ON DELETE CASCADE,
    CONSTRAINT fk_to_node_dependency FOREIGN KEY (to_node_id) REFERENCES nodes (id) ON DELETE CASCADE,

    CONSTRAINT no_self_dependency CHECK (from_node_id <> to_node_id)
);

CREATE INDEX idx_dependencies_pair ON dependencies(from_node_id, to_node_id);