CREATE TABLE IF NOT EXISTS phab_users 
(
    id              INTEGER PRIMARY KEY,
    type            TEXT,
    phid            TEXT,
    username        TEXT,
    real_name       TEXT,
    date_created    INTEGER,
    date_modified   INTEGER,
    roles           TEXT,
    policy_edit     TEXT,
    policy_view     TEXT
);
