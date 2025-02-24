package services

import (
	"context"
	"lexicon/singapore-supreme-court-crawler/common"
	"lexicon/singapore-supreme-court-crawler/crawler/models"
	"lexicon/singapore-supreme-court-crawler/repository"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func UpsertUrl(ctx context.Context, urlFrontier []repository.UrlFrontier) error {

	tx, err := common.Pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	queries := common.Queries.WithTx(tx)

	toModel := lo.Map(urlFrontier, func(url repository.UrlFrontier, _ int) repository.UpsertUrlFrontiersParams {
		return repository.UpsertUrlFrontiersParams{
			ID:        url.ID,
			Domain:    url.Domain,
			Url:       url.Url,
			Crawler:   url.Crawler,
			Status:    int16(url.Status),
			Metadata:  url.Metadata,
			CreatedAt: url.CreatedAt,
			UpdatedAt: url.UpdatedAt,
		}
	})

	res := queries.UpsertUrlFrontiers(ctx, toModel)
	res.Exec(func(i int, err error) {
		if err != nil {

			log.Err(err).Msg("failed to upsert url frontier")
			return
		}
	})

	return tx.Commit(ctx)
}

func UpdateFrontierStatuses(ctx context.Context, statuses []lo.Tuple2[string, int16]) error {
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("failed to begin transaction")
		return err
	}
	defer tx.Rollback(ctx)

	queries := common.Queries.WithTx(tx)

	res := queries.UpdateUrlFrontierStatus(ctx, lo.Map(statuses, func(status lo.Tuple2[string, int16], _ int) repository.UpdateUrlFrontierStatusParams {
		return repository.UpdateUrlFrontierStatusParams{
			ID:        status.A,
			Status:    status.B,
			UpdatedAt: time.Now(),
		}
	}))

	res.Exec(func(i int, err error) {
		if err != nil {
			log.Err(err).Msg("failed to update url frontier status")
			return
		}
	})

	return tx.Commit(ctx)
}

func GetUrlFrontierByUrl(ctx context.Context, url string) (repository.UrlFrontier, error) {
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("failed to begin transaction")

		return repository.UrlFrontier{}, err
	}

	queries := common.Queries.WithTx(tx)

	urlFrontier, err := queries.GetUrlFrontierByUrl(ctx, url)
	if err != nil {
		log.Err(err).Msg("failed to get url frontier by url")
		return repository.UrlFrontier{}, err
	}

	return repository.UrlFrontier{
		ID:        urlFrontier.ID,
		Domain:    urlFrontier.Domain,
		Url:       urlFrontier.Url,
		Crawler:   urlFrontier.Crawler,
		Status:    urlFrontier.Status,
		Metadata:  urlFrontier.Metadata,
		CreatedAt: urlFrontier.CreatedAt,
		UpdatedAt: urlFrontier.UpdatedAt,
	}, nil
}

func GetUrlFrontierById(ctx context.Context, id string) (repository.UrlFrontier, error) {
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("failed to begin transaction")

		return repository.UrlFrontier{}, err
	}

	queries := common.Queries.WithTx(tx)

	urlFrontier, err := queries.GetUrlFrontierById(ctx, id)
	if err != nil {
		log.Err(err).Msg("failed to get url frontier by id")
		return repository.UrlFrontier{}, err
	}

	return repository.UrlFrontier{
		ID:        urlFrontier.ID,
		Domain:    urlFrontier.Domain,
		Url:       urlFrontier.Url,
		Crawler:   urlFrontier.Crawler,
		Status:    urlFrontier.Status,
		Metadata:  urlFrontier.Metadata,
		CreatedAt: urlFrontier.CreatedAt,
		UpdatedAt: urlFrontier.UpdatedAt,
	}, nil
}

func GetUnscrappedUrlFrontiers(ctx context.Context, limit int32) ([]repository.UrlFrontier, error) {
	tx, err := common.Pool.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("failed to begin transaction")

		return nil, err
	}
	queries := common.Queries.WithTx(tx)

	urlFrontiers, err := queries.GetUnscrappedUrlFrontiers(ctx, repository.GetUnscrappedUrlFrontiersParams{
		Crawler: common.CRAWLER_NAME,
		Status:  models.URL_FRONTIER_STATUS_NEW,
		Limit:   limit,
	})
	if err != nil {
		log.Err(err).Msg("failed to get unscrapped url frontiers")
		return nil, err
	}

	res := lo.Map(urlFrontiers, func(u repository.GetUnscrappedUrlFrontiersRow, _ int) repository.UrlFrontier {
		return repository.UrlFrontier{
			ID:        u.ID,
			Domain:    u.Domain,
			Url:       u.Url,
			Crawler:   u.Crawler,
			Status:    u.Status,
			Metadata:  u.Metadata,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	})

	return res, nil
}
