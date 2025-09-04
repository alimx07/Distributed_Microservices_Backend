CREATE TABLE IF NOT EXISTS outbox(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic TEXT NOT NULL ,
    kafka_key TEXT NOT NULL,
    payload BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT Now(),
    status INT NOT NULL DEFAULT 0 ; -- 0: Pending , 1:Done , 2:Failed
)