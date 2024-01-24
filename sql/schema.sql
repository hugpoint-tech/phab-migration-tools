PRAGMA foreign_keys = ON;

-- Global object registry is needed to enforce correctness of
-- relationships. I.e. to have cascade deletes and other sql goodies
-- also holds original json values obtained from APIs
CREATE TABLE IF NOT EXISTS nodes (
    id        TEXT PRIMARY KEY,
    type      TEXT,
    eid       INTEGER,         -- eintity id
    raw       TEXT,            -- json
    UNIQUE (type, eid)
);

CREATE TABLE IF NOT EXISTS users (
    id        TEXT PRIMARY KEY REFERENCES nodes(id),
    email     TEXT NOT NULL,
    real_name TEXT,
    name      TEXT
);

CREATE TABLE IF NOT EXISTS bugs (
    id         TEXT PRIMARY KEY REFERENCES nodes(id),
    summary    TEXT,
    created_at TEXT,
    updated_at TEXT
);

-- Label is a key-value property of a node.
-- Lables do not carry realationship information
CREATE TABLE IF NOT EXISTS labels (
    obj TEXT,
    key TEXT,
    val TEXT, -- json
    PRIMARY KEY (obj, key),
    FOREIGN KEY (obj) REFERENCES nodes(id)
);

CREATE TABLE IF NOT EXISTS edges (
    id   INTEGER PRIMARY KEY,
    name TEXT,
    data TEXT -- json
);

CREATE TABLE IF NOT EXISTS edge_data (
    rel INTEGER,
    obj TEXT,
    idx INTEGER,
    PRIMARY KEY (rel, obj),
    FOREIGN KEY (rel) REFERENCES edges(id)
    FOREIGN KEY (obj) REFERENCES nodes(id),
    UNIQUE (rel, obj)
);
