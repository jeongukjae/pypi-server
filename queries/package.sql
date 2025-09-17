-- name: ListPackagesSimple :many
SELECT
    name
FROM
    package;

-- name: GetPackageByName :one
SELECT
    name,
    summary,
    latest_version,
    created_at,
    last_uploaded_at,
    updated_at
FROM
    package
WHERE
    name = $1;

-- name: UpdatePackageLatestVersion :exec
UPDATE
    package
SET
    summary = $2,
    latest_version = $3,
    last_uploaded_at = NOW(),
    updated_at = NOW()
WHERE
    name = $1;

-- name: ListReleaseFilesByPackageNameSimple :many
SELECT
    version,
    file_name,
    file_type,
    md5_digest,
    sha256_digest,
    blake2_256_digest,
    requires_python
FROM
    release_file
WHERE
    package_name = $1
ORDER BY
    created_at DESC;

-- name: GetPackage :one
SELECT
    name,
    summary,
    latest_version
FROM package
WHERE name = $1;

-- name: CreatePackage :one
INSERT INTO package (name, summary, latest_version)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdatePackageDescription :one
UPDATE package
SET summary = $2, updated_at = NOW()
WHERE name = $1
RETURNING *;

-- name: UpsertRelease :one
INSERT INTO release (
    version,
    package_name,
    metadata_version,
    summary,
    description,
    description_content_type
) VALUES (
    $1, $2, $3, $4, $5, $6
) ON CONFLICT (package_name, version) DO UPDATE SET
    metadata_version = EXCLUDED.metadata_version,
    summary = EXCLUDED.summary,
    description = EXCLUDED.description,
    description_content_type = EXCLUDED.description_content_type,
    updated_at = NOW()
RETURNING *;

-- name: CreateReleaseFile :one
INSERT INTO release_file (
    version,
    package_name,
    file_name,
    file_type,
    file_path,
    file_size_bytes,
    pyversion,
    requires_python,
    requires_dist,
    md5_digest,
    sha256_digest,
    blake2_256_digest
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetRelease :one
SELECT
    version,
    package_name
FROM
    release
WHERE
    package_name = $1 AND version = $2;

-- name: GetReleaseFileByName :one
SELECT
    version,
    file_name,
    file_path
FROM
    release_file
WHERE
    package_name = $1 AND file_name = $2;
