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

-- name: ListReleasesByPackageNameSimple :many
SELECT
    version,
    file_name,
    file_type,
    md5_digest,
    sha256_digest,
    blake2_256_digest,
    requires_python
FROM
    release
WHERE
    package_name = $1
ORDER BY
    created_at DESC;
