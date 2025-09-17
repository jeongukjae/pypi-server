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
    package_name             TEXT REFERENCES package(name) ON DELETE CASCADE NOT NULL,
    metadata_version         TEXT NOT NULL,

    summary                  TEXT,
    description              TEXT,
    description_content_type TEXT,

    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    yanked        BOOLEAN DEFAULT FALSE,
    yanked_reason TEXT,
    yanked_at     TIMESTAMPTZ,

    UNIQUE(package_name, version)
);

CREATE TABLE release_file (
    version      TEXT REFERENCES release(version) ON DELETE CASCADE NOT NULL,
    package_name TEXT REFERENCES package(name) ON DELETE CASCADE NOT NULL,

    file_name       TEXT NOT NULL,
    file_type       TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    file_size_bytes INTEGER NOT NULL,

    pyversion                TEXT,
    requires_python          TEXT,
    requires_dist            TEXT[],
    md5_digest               TEXT,
    sha256_digest            TEXT,
    blake2_256_digest        TEXT,

    created_at  TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (package_name, version, file_name)
);

CREATE INDEX idx_release_file_package_name ON release_file(package_name);
CREATE INDEX idx_release_file_package_name_file_name ON release_file(package_name, file_name);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_release_file_package_name_file_name;
DROP INDEX IF EXISTS idx_release_file_package_name;

DROP TABLE IF EXISTS release_file;
DROP TABLE IF EXISTS release;
DROP TABLE IF EXISTS package;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
