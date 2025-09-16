-- Write your migrate up statements here

CREATE TABLE package (
    name             TEXT PRIMARY KEY,
    summary          TEXT,

    latest_version   TEXT,

    created_at       TIMESTAMPTZ DEFAULT NOW(),
    last_uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE release (
    version                  TEXT PRIMARY KEY,
    package_name             TEXT REFERENCES package(name) ON DELETE CASCADE,
    summary                  TEXT,
    description              TEXT,
    description_content_type TEXT,

    file_name                TEXT,
    file_type                TEXT,

    pyversion                TEXT,
    requires_python          TEXT,
    requires_dist            TEXT[],

    md5_digest               TEXT,
    sha256_digest            TEXT,
    blake2_256_digest        TEXT,

    created_at    TIMESTAMPTZ DEFAULT NOW(),
    yanked        BOOLEAN DEFAULT FALSE,
    yanked_reason TEXT,
    yanked_at     TIMESTAMPTZ
);

---- create above / drop below ----

DROP TABLE IF EXISTS release;
DROP TABLE IF EXISTS package;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
