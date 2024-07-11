CREATE TABLE eth_price_messages (
    id SERIAL PRIMARY KEY,
    message_id TEXT NOT NULL UNIQUE,
    price NUMERIC NOT NULL,
    publisher TEXT NOT NULL,
    writer TEXT NOT NULL,
    signers TEXT[] NOT NULL,
    signatures JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_messages_timestamp ON eth_price_messages (timestamp DESC);
