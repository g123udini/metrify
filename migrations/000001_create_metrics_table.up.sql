CREATE TABLE IF NOT EXISTS metrics (
                               id               BIGSERIAL PRIMARY KEY,
                               created_at       TIMESTAMP NOT NULL DEFAULT NOW(),

                               name            CHAR(100),
                               value           DOUBLE PRECISION,
                               type            CHAR(100)
);

CREATE INDEX IF NOT EXISTS created_at_idx ON metrics(created_at);-- migrations/000001_create_metrics_table.up.sql