CREATE TABLE IF NOT EXISTS groups (
    id BIGINT NOT NULL PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    owner_id BIGINT NOT NULL,
    name VARCHAR(128) NOT NULL,
    description VARCHAR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(10) DEFAULT 'member',
    
    PRIMARY KEY (group_id, user_id),
    CONSTRAINT no_group_member_role CHECK (role IN ('admin', 'editor', 'member')),
    
    CONSTRAINT fk_member_group FOREIGN KEY (group_id) REFERENCES groups(id)  ON DELETE CASCADE
);
