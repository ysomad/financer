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

CREATE TABLE IF NOT EXISTS categories (
    name varchar(64) PRIMARY KEY NOT NULL,
    description text,
    author varchar(64) REFERENCES identities (id),
    created_at timestamptz NOT NULL
);

INSERT INTO categories(name, description, created_at) VALUES
    ('Rent', 'Payments for rent', CURRENT_TIMESTAMP),
    ('Public utilities', 'Public utilities payments, usually comes with rent', CURRENT_TIMESTAMP),
    ('Telecommunications', 'Payments for mobile plans and internet in general', CURRENT_TIMESTAMP),
    ('Sport', 'Gym memberships and coach expenses', CURRENT_TIMESTAMP),
    ('Groceries', 'Grocery stores expenses', CURRENT_TIMESTAMP),
    ('Home', 'Home decor, accessories etc', CURRENT_TIMESTAMP),
    ('Food service', 'Food delivery and restaraunt expenses', CURRENT_TIMESTAMP),
    ('Entertainment', 'Cinema, clubs, games, theaters etc', CURRENT_TIMESTAMP),
    ('Healthcare', 'Doctor consultions, bloodwork and other healthcare expenses', CURRENT_TIMESTAMP),
    ('Pharmacy', 'Medication purchases', CURRENT_TIMESTAMP),
    ('Selfcare', 'Barbers and expenses on selfcare products including fragrances', CURRENT_TIMESTAMP),
    ('Electronics', 'Expenses for any electronics', CURRENT_TIMESTAMP),
    ('Shopping', 'Clothes, shoes and accessories expenses', CURRENT_TIMESTAMP),
    ('Trips', 'Expenses on any long distance trips', CURRENT_TIMESTAMP),
    ('Public transport', 'Buses, trams and other public transport', CURRENT_TIMESTAMP),
    ('Taxi', 'Taxi expenses', CURRENT_TIMESTAMP),
    ('Gifts', 'Gifts expenses', CURRENT_TIMESTAMP),
    ('Education', 'Books, courses etc', CURRENT_TIMESTAMP),
    ('Online services', 'Online service subscriptions, like Spotify Premium and other', CURRENT_TIMESTAMP),
    ('Other', 'I dont know what I spent my money on', CURRENT_TIMESTAMP),
    ('Taxes', 'PLOTI NALOGI', CURRENT_TIMESTAMP),
    ('Earnings', 'did i lose more than earn?', CURRENT_TIMESTAMP);

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
DROP TABLE IF EXISTS identity_traits;
DROP TABLE IF EXISTS identities;
-- +goose StatementEnd
