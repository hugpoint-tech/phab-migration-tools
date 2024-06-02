CREATE TABLE IF NOT EXISTS bugs (
    id INTEGER PRIMARY KEY,
    CreationTime TEXT,
    Creator TEXT,
    Summary TEXT,
    OtherFieldsJSON TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    Creator TEXT NOT NULL UNIQUE
);