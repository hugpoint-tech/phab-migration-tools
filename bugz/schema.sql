CREATE TABLE IF NOT EXISTS bugs (
    id INTEGER PRIMARY KEY,
    CreationTime TEXT,
    Creator TEXT,
    Summary TEXT,
    OtherFieldsJSON TEXT
);