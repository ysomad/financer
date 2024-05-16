-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id bigint PRIMARY KEY NOT NULL,
    currency char(3) NOT NULL,
    language char(2) NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE TYPE category_type AS ENUM ('EXPENSES', 'INCOME', 'OTHER');

CREATE TABLE IF NOT EXISTS categories (
    id uuid PRIMARY KEY NOT NULL,
    name varchar(64) NOT NULL,
    type category_type NOT NULL,
    author bigint REFERENCES users (id),
    created_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE INDEX idx_active_categories ON categories (deleted_at)
WHERE deleted_at is NULL;

INSERT INTO categories(id, name, type, created_at) VALUES
    ('fdb43add-1c82-48e6-b0e7-fc2068bda90a', '🛏️ Rent', 'EXPENSES', CURRENT_TIMESTAMP),
    ('0fee8fb2-1610-4b7d-9aab-d5ddd668cae6', '💡 Public utilities', 'EXPENSES', CURRENT_TIMESTAMP),
    ('15627f24-1488-45d1-b75c-6d753ba3a707', '📱 Mobile and internet', 'EXPENSES', CURRENT_TIMESTAMP),
    ('cbd7fae1-1ed5-4816-ba3b-243706849c0d', '💪 Sport', 'EXPENSES', CURRENT_TIMESTAMP),
    ('e601f59c-417d-43c3-b948-7b3c00a0e4e0', '🍏 Groceries', 'EXPENSES', CURRENT_TIMESTAMP),
    ('d94846dd-4445-4ac0-acc6-422f5e58f3e5', '🕯️ Home', 'EXPENSES', CURRENT_TIMESTAMP),
    ('731e3850-3292-4954-bab0-74db9f82077d', '🍕 Food service', 'EXPENSES', CURRENT_TIMESTAMP),
    ('0d7ae08c-c42d-4708-925e-034cb17ead43', '🎮 Entertainment', 'EXPENSES', CURRENT_TIMESTAMP),
    ('11bd8f86-fe5d-414e-9dec-0fe02a9a40b4', '🏥 Healthcare', 'EXPENSES', CURRENT_TIMESTAMP),
    ('c8e3a567-bd7a-4757-9303-d709036b7974', '💊 Pharmacy', 'EXPENSES', CURRENT_TIMESTAMP),
    ('a0069f1a-77fc-4fe8-852d-cb85529dec6a', '🛀 Selfcare', 'EXPENSES', CURRENT_TIMESTAMP),
    ('b1e10a8a-564b-4f00-8e88-bfded6de995f', '💻 Electronics', 'EXPENSES', CURRENT_TIMESTAMP),
    ('93e8172d-dbb6-456b-8efd-261c6ad7d13d', '🛍️ Shopping', 'EXPENSES', CURRENT_TIMESTAMP),
    ('21aa7ded-3e72-459f-9984-1893edf56dde', '✈️ Trips', 'EXPENSES', CURRENT_TIMESTAMP),
    ('00f3cc4f-6c08-4c45-9e8f-e3ed9072971c', '🚌 Public transport', 'EXPENSES', CURRENT_TIMESTAMP),
    ('54900743-9a9a-44b5-82ed-6ffe5868333d', '🚕 Taxi', 'EXPENSES', CURRENT_TIMESTAMP),
    ('6985154f-59ed-433b-88a7-754d6d4721c3', '🎁 Gifts', 'EXPENSES', CURRENT_TIMESTAMP),
    ('b9a9f57f-925a-41ab-a480-7f89255cc008', '📚 Education', 'EXPENSES', CURRENT_TIMESTAMP),
    ('94e37cf7-8dfd-485d-8258-bc4bdddcbc92', '📧 Online services', 'EXPENSES', CURRENT_TIMESTAMP),
    ('e399a511-9b62-41d5-9ca6-4d3a5337112e', '🏛️ Taxes', 'EXPENSES', CURRENT_TIMESTAMP),
    ('00000000-0000-0000-0000-000000000000', '🤷 Other', 'OTHER', CURRENT_TIMESTAMP),
    ('4f852705-245a-4c47-9f44-dd94ed243548', '💳 Salary', 'INCOME', CURRENT_TIMESTAMP),
    ('411afc38-3f5f-4b57-9834-3de07f830b74', '💰 Bonuses', 'INCOME', CURRENT_TIMESTAMP);

CREATE TABLE IF NOT EXISTS user_categories (
    user_id bigint NOT NULL REFERENCES users (id),
    category_id uuid NOT NULL REFERENCES categories (id),
    CONSTRAINT user_categories_pkey PRIMARY KEY (user_id, category_id)
);

CREATE TABLE IF NOT EXISTS operations (
    id uuid PRIMARY KEY NOT NULL,
    user_id bigint NOT NULL REFERENCES users (id),
    category_id uuid REFERENCES categories (id),
    name varchar(64) NOT NULL,
    currency char(3) NOT NULL,
    money int NOT NULL,
    occured_at date NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS user_keywords (
    operation varchar(64) NOT NULL,
    user_id bigint NOT NULL REFERENCES users (id),
    category_id uuid NOT NULL REFERENCES categories (id)
);

CREATE UNIQUE INDEX idx_user_keyword ON user_keywords (operation, user_id, category_id);

CREATE INDEX idx_active_operations ON operations (deleted_at)
WHERE deleted_at is NULL;

-- fulltext search (https://pgroonga.github.io/tutorial/)
CREATE EXTENSION IF NOT EXISTS pgroonga;

-- Disable sequential scan to ensure using pgroonga index
SET enable_seqscan = off;

CREATE INDEX idx_pgroonga_category_name ON categories USING pgroonga (name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_keywords;
DROP TABLE IF EXISTS user_categories;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS categories;
DROP TYPE IF EXISTS category_type;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
