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
    telegram_uid bigint UNIQUE,
    updated_at timestamptz
);

CREATE TYPE category_type AS ENUM ('EXPENSES', 'EARNINGS');

CREATE TABLE IF NOT EXISTS categories (
    name varchar(64) PRIMARY KEY NOT NULL,
    type category_type NOT NULL,
    author varchar(64) REFERENCES identities (id),
    created_at timestamptz NOT NULL
);

INSERT INTO categories(name, type, created_at) VALUES
    ('🛏️ Rent', 'EXPENSES', CURRENT_TIMESTAMP),
    ('💡 Public utilities', 'EXPENSES', CURRENT_TIMESTAMP),
    ('📱 Mobile and internet', 'EXPENSES', CURRENT_TIMESTAMP),
    ('💪 Sport', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🍏 Groceries', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🕯️ Home', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🍕 Food service', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🎮 Entertainment', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🏥 Healthcare', 'EXPENSES', CURRENT_TIMESTAMP),
    ('💊 Pharmacy', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🛀 Selfcare', 'EXPENSES', CURRENT_TIMESTAMP),
    ('💻 Electronics', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🛍️ Shopping', 'EXPENSES', CURRENT_TIMESTAMP),
    ('✈️ Trips', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🚌 Public transport', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🚕 Taxi', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🎁 Gifts', 'EXPENSES', CURRENT_TIMESTAMP),
    ('📚 Education', 'EXPENSES', CURRENT_TIMESTAMP),
    ('📧 Online services', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🏛️ Taxes', 'EXPENSES', CURRENT_TIMESTAMP),
    ('🤷 Other', 'EXPENSES', CURRENT_TIMESTAMP),
    ('💳 Salary', 'EARNINGS', CURRENT_TIMESTAMP),
    ('💰 Bonuses', 'EARNINGS', CURRENT_TIMESTAMP);

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
