-- name: UpsertUrlFrontier :exec
INSERT INTO url_frontiers (id, domain, url, crawler, status, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET
  domain = $2,
  url = $3,
  crawler = $4,
  metadata = $6,
  updated_at = $7;


-- name: UpsertUrlFrontiers :batchexec
INSERT INTO url_frontiers (id, domain, url, crawler, status, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET
  domain = $2,
  url = $3,
  crawler = $4,
  metadata = $6,
  updated_at = $7;

-- name: UpdateUrlFrontierStatus :batchexec
UPDATE url_frontiers
SET
  status = $2,
  updated_at = $3
WHERE id = $1;

-- name: GetUnscrappedUrlFrontiers :many
SELECT id, domain, url, crawler, status, metadata, created_at, updated_at
FROM url_frontiers
WHERE
  crawler = $1
  AND status = $2
ORDER BY created_at ASC LIMIT $3;

-- name: UpsertExtraction :batchexec
INSERT INTO extractions (id, url_frontier_id, site_content, raw_page_link, language, page_hash, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE
SET
  url_frontier_id = $2,
  site_content = $3,
  raw_page_link = $4,
  language = $5,
  page_hash = $6,
  metadata = $7,
  updated_at = $8;


-- name: GetUrlFrontierByUrl :one
SELECT id, domain, url, crawler, status, metadata, created_at, updated_at
FROM url_frontiers
WHERE url = $1
LIMIT 1;

-- name: GetUrlFrontierById :one
SELECT id, domain, url, crawler, status, metadata, created_at, updated_at
FROM url_frontiers
WHERE id = $1
LIMIT 1;
