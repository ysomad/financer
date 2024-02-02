-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS identities (
    id varchar(64) PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS identity_traits (
    identity_id varchar(64) NOT NULL UNIQUE REFERENCES identities (id),
    default_currency char(3) NOT NULL,
    telegram_uid bigint UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE identity_traits;
DROP TABLE identities;
-- +goose StatementEnd
