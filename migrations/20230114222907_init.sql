-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id      CHAR(32) PRIMARY KEY NOT NULL,
    balance BIGINT               NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions
(
    id        CHAR(64) PRIMARY KEY      NOT NULL,
    "from"    CHAR(32) REFERENCES users NOT NULL,
    "to"      CHAR(32) REFERENCES users NOT NULL,
    value     BIGINT                    NOT NULL,
    timestamp TIMESTAMP                 NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
