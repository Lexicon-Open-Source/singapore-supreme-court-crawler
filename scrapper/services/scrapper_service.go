package services

import (
	"context"
	"lexicon/singapore-supreme-court-crawler/common"
	"lexicon/singapore-supreme-court-crawler/repository"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func UpsertExtraction(ctx context.Context, extractions []repository.Extraction) error {
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	queries := common.Query.WithTx(tx)

	br := queries.UpsertExtraction(ctx, lo.Map(extractions, func(extraction repository.Extraction, _ int) repository.UpsertExtractionParams {
		return repository.UpsertExtractionParams{
			ID:            extraction.ID,
			UrlFrontierID: extraction.UrlFrontierID,
			SiteContent:   extraction.SiteContent,
			ArtifactLink:  extraction.ArtifactLink,
			RawPageLink:   extraction.RawPageLink,
			Language:      extraction.Language,
			PageHash:      extraction.PageHash,
			Metadata:      extraction.Metadata,
			CreatedAt:     extraction.CreatedAt,
			UpdatedAt:     extraction.UpdatedAt,
		}
	}))

	br.Exec(func(int, error) {
		if err != nil {
			log.Error().Err(err).Msg("Error upserting extractions")
			return
		}
	})

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
