package models

import (
	"context"
	"encoding/json"

	"github.com/golang-module/carbon/v2"
	"github.com/guregu/null"
	"github.com/jackc/pgx/v5"
)

type Extraction struct {
	Id            string
	UrlFrontierId string
	SiteContent   null.String
	ArtifactLink  null.String
	RawPageLink   null.String
	Metadata      Metadata
	CreatedAt     carbon.DateTime
	UpdatedAt     carbon.DateTime
	Language      string
}

func (e *Extraction) AddSiteContent(content string) {
	e.SiteContent = null.StringFrom(content)
}

func (e *Extraction) AddArtifactLink(link string) {
	e.ArtifactLink = null.StringFrom(link)
}

func (e *Extraction) AddRawPageLink(link string) {
	e.RawPageLink = null.StringFrom(link)
}

func (e *Extraction) AddMetadata(metadata Metadata) {
	e.Metadata = metadata
}
func (e *Extraction) UpdateUpdatedAt() {
	e.UpdatedAt = carbon.Now().ToDateTimeStruct()
}
func NewExtraction() Extraction {
	return Extraction{
		Id:            "",
		UrlFrontierId: "",
		SiteContent:   null.StringFrom(""),
		ArtifactLink:  null.StringFrom(""),
		RawPageLink:   null.StringFrom(""),
		Metadata:      Metadata{},
		CreatedAt:     carbon.Now().ToDateTimeStruct(),
		UpdatedAt:     carbon.Now().ToDateTimeStruct(),
		Language:      "en",
	}
}

func UpsertExtraction(ctx context.Context, tx pgx.Tx, extraction Extraction) error {

	sql := `INSERT INTO extraction (id, url_frontier_id, site_content, artifact_link, raw_page_link, metadata, created_at, updated_at, language) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO UPDATE SET site_content = EXCLUDED.site_content, artifact_link = EXCLUDED.artifact_link, raw_page_link = EXCLUDED.raw_page_link, metadata = EXCLUDED.metadata, updated_at = EXCLUDED.updated_at, language = EXCLUDED.language`

	metadataJson, err := json.Marshal(extraction.Metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, extraction.Id, extraction.UrlFrontierId, extraction.SiteContent, extraction.ArtifactLink, extraction.RawPageLink, metadataJson, extraction.CreatedAt.ToIso8601String(), extraction.UpdatedAt.ToIso8601String(), extraction.Language)
	if err != nil {
		return err
	}

	return nil
}

func GetExistingExtraction(ctx context.Context, tx pgx.Tx, url string) (Extraction, error) {

	sql := `SELECT id, url_frontier_id, site_content, artifact_link, raw_page_link, metadata, created_at, updated_at, language FROM extraction WHERE raw_page_link = $1 LIMIT 1`

	var extraction Extraction

	rows, err := tx.Query(ctx, sql, url)
	if err != nil {
		return Extraction{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var stringMetadata string

		err = rows.Scan(&extraction.Id, &extraction.UrlFrontierId, &extraction.SiteContent, &extraction.ArtifactLink, &extraction.RawPageLink, &stringMetadata, &extraction.CreatedAt, &extraction.UpdatedAt, &extraction.Language)
		if err != nil {
			return Extraction{}, err
		}

		err = json.Unmarshal([]byte(stringMetadata), &extraction.Metadata)
		if err != nil {
			return Extraction{}, err
		}
	}

	return extraction, nil
}
