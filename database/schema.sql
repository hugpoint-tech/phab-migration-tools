CREATE TABLE IF NOT EXISTS bugs (
    id INTEGER PRIMARY KEY,
    CreationTime TEXT,
    Creator TEXT,
    OtherFieldsJSON TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    Email TEXT,
    Name TEXT,
    RealName TEXT
);

CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY,
    bug_id INTEGER,
	attachment_id INTEGER,
    creation_time TEXT,
    creator TEXT,
    text TEXT,
    FOREIGN KEY(bug_id) REFERENCES bugs(id)
);

CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY,
    bug_id INTEGER,
    creation_time TEXT,
    creator TEXT,
    summary TEXT,
    data BLOB,
    FOREIGN KEY(bug_id) REFERENCES bugs(id)

);