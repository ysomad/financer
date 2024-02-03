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
    currency char(3) NOT NULL,
    telegram_uid bigint UNIQUE
);

CREATE TYPE category_type AS ENUM ('Expenses', 'Earnings');

CREATE TABLE IF NOT EXISTS categories (
    name varchar(64) PRIMARY KEY NOT NULL,
    type category_type NOT NULL,
    author varchar(64) REFERENCES identities (id),
    created_at timestamptz NOT NULL
);

INSERT INTO categories(name, type, created_at) VALUES
    ('🛏️Rent', 'Expenses', CURRENT_TIMESTAMP),
    ('💡Public utilities', 'Expenses', CURRENT_TIMESTAMP),
    ('🌐Telecommunications', 'Expenses', CURRENT_TIMESTAMP),
    ('💪Sport', 'Expenses', CURRENT_TIMESTAMP),
    ('🍏Groceries', 'Expenses', CURRENT_TIMESTAMP),
    ('🕯️Home', 'Expenses', CURRENT_TIMESTAMP),
    ('🍕Food service', 'Expenses', CURRENT_TIMESTAMP),
    ('🎮Entertainment', 'Expenses', CURRENT_TIMESTAMP),
    ('🏥Healthcare', 'Expenses', CURRENT_TIMESTAMP),
    ('💊Pharmacy', 'Expenses', CURRENT_TIMESTAMP),
    ('🛀Selfcare', 'Expenses', CURRENT_TIMESTAMP),
    ('💻Electronics', 'Expenses', CURRENT_TIMESTAMP),
    ('🛍️Shopping', 'Expenses', CURRENT_TIMESTAMP),
    ('✈️Trips', 'Expenses', CURRENT_TIMESTAMP),
    ('🚌Public transport', 'Expenses', CURRENT_TIMESTAMP),
    ('🚕Taxi', 'Expenses', CURRENT_TIMESTAMP),
    ('🎁Gifts', 'Expenses', CURRENT_TIMESTAMP),
    ('📚Education', 'Expenses', CURRENT_TIMESTAMP),
    ('📧Online services', 'Expenses', CURRENT_TIMESTAMP),
    ('🏛️Taxes', 'Expenses', CURRENT_TIMESTAMP),
    ('🤷Other', 'Expenses', CURRENT_TIMESTAMP),
    ('💳Salary', 'Earnings', CURRENT_TIMESTAMP),
    ('💰Bonuses', 'Earnings', CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS identity_categories (
    identity_id varchar(64) NOT NULL REFERENCES identities (id),
    category varchar(64) NOT NULL REFERENCES categories (name),
    CONSTRAINT identity_categories_pkey PRIMARY KEY (identity_id, category) -- implement many2many
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity_categories;
DROP TABLE IF EXISTS categories;
DROP TYPE IF EXISTS category_type;
DROP TABLE IF EXISTS identity_traits;
DROP TABLE IF EXISTS identities;
-- +goose StatementEnd
