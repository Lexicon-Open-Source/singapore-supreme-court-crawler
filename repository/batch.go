// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: batch.go

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	crawlerModel "lexicon/singapore-supreme-court-crawler/crawler/models"
	scrapperModel "lexicon/singapore-supreme-court-crawler/scrapper/models"
)

var (
	ErrBatchAlreadyClosed = errors.New("batch already closed")
)

const updateUrlFrontierStatus = `-- name: UpdateUrlFrontierStatus :batchexec
UPDATE url_frontiers
SET
  status = $2,
  updated_at = $3
WHERE id = $1
`

type UpdateUrlFrontierStatusBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type UpdateUrlFrontierStatusParams struct {
	ID        string
	Status    int16
	UpdatedAt time.Time
}

func (q *Queries) UpdateUrlFrontierStatus(ctx context.Context, arg []UpdateUrlFrontierStatusParams) *UpdateUrlFrontierStatusBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.ID,
			a.Status,
			a.UpdatedAt,
		}
		batch.Queue(updateUrlFrontierStatus, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &UpdateUrlFrontierStatusBatchResults{br, len(arg), false}
}

func (b *UpdateUrlFrontierStatusBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *UpdateUrlFrontierStatusBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const upsertExtraction = `-- name: UpsertExtraction :batchexec
INSERT INTO extractions (id, url_frontier_id, site_content, artifact_link, raw_page_link, language, page_hash, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE
SET
  url_frontier_id = $2,
  site_content = $3,
  artifact_link = $4,
  raw_page_link = $5,
  language = $6,
  page_hash = $7,
  metadata = $8,
  updated_at = $9
`

type UpsertExtractionBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type UpsertExtractionParams struct {
	ID            string
	UrlFrontierID string
	SiteContent   *string
	ArtifactLink  *string
	RawPageLink   *string
	Language      string
	PageHash      *string
	Metadata      scrapperModel.Metadata
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (q *Queries) UpsertExtraction(ctx context.Context, arg []UpsertExtractionParams) *UpsertExtractionBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.ID,
			a.UrlFrontierID,
			a.SiteContent,
			a.ArtifactLink,
			a.RawPageLink,
			a.Language,
			a.PageHash,
			a.Metadata,
			a.CreatedAt,
			a.UpdatedAt,
		}
		batch.Queue(upsertExtraction, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &UpsertExtractionBatchResults{br, len(arg), false}
}

func (b *UpsertExtractionBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *UpsertExtractionBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const upsertUrlFrontiers = `-- name: UpsertUrlFrontiers :batchexec
INSERT INTO url_frontiers (id, domain, url, crawler, status, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET
  domain = $2,
  url = $3,
  crawler = $4,
  metadata = $6,
  updated_at = $7
`

type UpsertUrlFrontiersBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type UpsertUrlFrontiersParams struct {
	ID        string
	Domain    string
	Url       string
	Crawler   string
	Status    int16
	Metadata  crawlerModel.UrlFrontierMetadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) UpsertUrlFrontiers(ctx context.Context, arg []UpsertUrlFrontiersParams) *UpsertUrlFrontiersBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.ID,
			a.Domain,
			a.Url,
			a.Crawler,
			a.Status,
			a.Metadata,
			a.CreatedAt,
			a.UpdatedAt,
		}
		batch.Queue(upsertUrlFrontiers, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &UpsertUrlFrontiersBatchResults{br, len(arg), false}
}

func (b *UpsertUrlFrontiersBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *UpsertUrlFrontiersBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}
