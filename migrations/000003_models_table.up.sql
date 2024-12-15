CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    model BYTEA NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT fk_user_models FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_id_models ON models(user_id);

-- Создаём функцию для обновления поля model_updated_at
CREATE OR REPLACE FUNCTION upd_model_updated_at() 
RETURNS TRIGGER 
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаём триггер, который срабатывает на обновление поля model
CREATE TRIGGER update_model_updated_at
BEFORE UPDATE OF model ON models
FOR EACH ROW 
EXECUTE PROCEDURE upd_model_updated_at();
