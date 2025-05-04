------------------------------------------------------------------------ Tasks Table

CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT NOT NULL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    planned_time FLOAT NOT NULL,
    actual_time FLOAT, 
    tags integer[]
);

------------------------------------------------------------------------ Models Table

CREATE TABLE IF NOT EXISTS models (
    user_id BIGINT NOT NULL,
    model BYTEA NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT false
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