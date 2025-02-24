package scrapper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	crawler_model "lexicon/singapore-supreme-court-crawler/crawler/models"
	crawler_service "lexicon/singapore-supreme-court-crawler/crawler/services"
	"lexicon/singapore-supreme-court-crawler/repository"
	"lexicon/singapore-supreme-court-crawler/scrapper/services"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type Scrapper interface {
	Setup()
	Teardown()
	Scrape(ctx context.Context, url string) error
	ScrapeAll(ctx context.Context) error
	ScrapeAllSequential(ctx context.Context) error
	Consume(m jetstream.Msg) error
}

type ScrapperImpl struct {
	browser *rod.Browser
}

func (c *ScrapperImpl) Setup() {
	c.browser = rod.New().MustConnect()
}

func (c *ScrapperImpl) Teardown() {
	c.browser.MustClose()
}

func (c *ScrapperImpl) Consume(m jetstream.Msg) error {
	return nil
}

func (c *ScrapperImpl) Scrape(ctx context.Context, url string) error {
	return nil
}

func (c *ScrapperImpl) ScrapeAll(ctx context.Context) error {

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure all resources are cleaned up

	var unscrappedUrlFrontiers []repository.UrlFrontier
	// Fetch unscrapped url frontier from db

	pagePool := rod.NewPagePool(10)
	defer pagePool.Cleanup(func(p *rod.Page) {
		err := p.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing page")
		}
	})

	create := func() (*rod.Page, error) {
		incognito, err := c.browser.Incognito()
		if err != nil {
			log.Error().Err(err).Msg("Error creating incognito page")
			return nil, err
		}
		return incognito.MustPage(), nil
	}

	job := func(ctx context.Context, urlFrontier repository.UrlFrontier, c chan<- repository.Extraction) error {
		page, err := pagePool.Get(create)
		if err != nil {
			log.Error().Err(err).Msg("Error getting page")
			return err
		}
		defer pagePool.Put(page)
		extraction, err := scrapeUrlFrontiers(ctx, page, urlFrontier)
		if err != nil {
			log.Error().Err(err).Msg("Error scraping url frontier")
			return err
		}
		c <- extraction
		return nil
	}

	for ok := true; ok; ok = len(unscrappedUrlFrontiers) > 0 {

		// Check if context is cancelled before starting new chunk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var scraperErrors []error
		unscrappedUrlFrontiers, err := crawler_service.GetUnscrappedUrlFrontiers(ctx, 100)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching unscrapped url frontier")
		}
		log.Info().Msgf("Unscrapped URLs: %d", len(unscrappedUrlFrontiers))

		chunks := lo.Chunk(unscrappedUrlFrontiers, 10)

		for _, urlFrontiers := range chunks {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			wg := sync.WaitGroup{}
			extractionChan := make(chan repository.Extraction, len(urlFrontiers))
			errChan := make(chan error, len(urlFrontiers))

			for _, urlFrontier := range urlFrontiers {
				wg.Add(1)
				go func(urlFrontier repository.UrlFrontier) {
					defer wg.Done()

					select {
					case <-ctx.Done():
						errChan <- ctx.Err()
					default:
						if err := job(ctx, urlFrontier, extractionChan); err != nil {
							errChan <- fmt.Errorf("error crawling %s: %w", urlFrontier.Url, err)
						}
					}
				}(urlFrontier)
			}

			// Wait for all goroutines to finish
			wg.Wait()

			// Close channels after all goroutines are done
			close(errChan)
			close(extractionChan)

			// Collect errors and extractions
			for err := range errChan {
				scraperErrors = append(scraperErrors, err)
			}

			var extractions []repository.Extraction
			for extraction := range extractionChan {
				extractions = append(extractions, extraction)
			}

			if len(scraperErrors) > 0 {
				log.Error().Err(scraperErrors[0]).Msg("Error scraping url frontier")
			}
			log.Info().Msgf("Upserting extractions")
			err := services.UpsertExtraction(ctx, extractions)
			if err != nil {
				log.Error().Err(err).Msg("Error upserting extractions")
			}
			log.Info().Msgf("Updating url frontier statuses")
			err = crawler_service.UpdateFrontierStatuses(ctx, lo.Map(urlFrontiers, func(urlFrontier repository.UrlFrontier, _ int) lo.Tuple2[string, int16] {
				return lo.Tuple2[string, int16]{A: urlFrontier.ID, B: crawler_model.URL_FRONTIER_STATUS_CRAWLED}
			}))
			if err != nil {
				log.Error().Err(err).Msg("Error updating url frontier status")
			}
			log.Info().Msgf("Finished scraping chunk")

		}

	}
	log.Info().Msgf("Finished scraping all url frontiers")
	return nil
}

func scrapeUrlFrontiers(ctx context.Context, page *rod.Page, urlFrontier repository.UrlFrontier) (repository.Extraction, error) {

	select {
	case <-ctx.Done():
		return repository.Extraction{}, ctx.Err()
	default:
	}
	log.Info().Msgf("Scraping url: %s", urlFrontier.Url)

	rpCtx := page.Context(ctx)
	wait := rpCtx.MustWaitNavigation()
	err := rpCtx.Navigate(urlFrontier.Url)
	if err != nil {
		log.Error().Err(err).Msg("Error navigating to url")
		return repository.Extraction{}, err
	}
	wait()

	rp := rpCtx.MustWaitStable()

	judgement, err := rp.Element("#divJudgement")
	if err != nil {
		log.Error().Err(err).Msg("Error getting judgement")
		return repository.Extraction{}, err
	}
	now := time.Now()
	var extraction repository.Extraction
	extraction.ID = urlFrontier.ID
	extraction.UrlFrontierID = urlFrontier.ID
	extraction.Language = "en"
	extraction.CreatedAt = now
	extraction.UpdatedAt = now
	siteContent, err := judgement.HTML()
	if err != nil {
		log.Error().Err(err).Msg("Error getting site content")
		return repository.Extraction{}, err
	}
	extraction.SiteContent = &siteContent
	nav, err := rp.Element("nav")
	if err != nil {
		log.Error().Err(err).Msg("Error getting nav")
		return repository.Extraction{}, err
	}
	donwloadLink, err := extractPdfUrl(nav)
	if err != nil {
		log.Error().Err(err).Msg("Error extracting pdf url")
		return repository.Extraction{}, err
	}
	extraction.Metadata.PdfUrl = donwloadLink
	hashPage := sha256.Sum256([]byte(*extraction.SiteContent))
	hashPageString := hex.EncodeToString(hashPage[:])
	extraction.PageHash = &hashPageString

	content, err := judgement.Element("content")
	if err != nil {
		log.Info().Msg("Judgement is old")
		err = scrapeOldTemplate(ctx, judgement, &extraction, &urlFrontier)
		if err != nil {
			log.Error().Err(err).Msg("Error scraping old template")
			return repository.Extraction{}, err
		}
	} else {
		err = scrapeNewTemplate(ctx, content, &extraction, &urlFrontier)
		if err != nil {
			log.Error().Err(err).Msg("Error scraping new template")
			return repository.Extraction{}, err
		}
	}
	log.Info().Msgf("Handling pdf for url: %s", urlFrontier.Url)
	pdfUrl, err := services.HandlePdf(ctx, extraction.ID, extraction.Metadata.PdfUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error handling pdf")
		return repository.Extraction{}, err
	}
	log.Info().Msgf("Uploaded pdf for url: %s, to GCS: %s", urlFrontier.Url, pdfUrl.B)
	log.Info().Msgf("Downloading html for url: %s", urlFrontier.Url)
	htmlUrl, err := services.HandleHtml(ctx, extraction.ID, urlFrontier.Url)
	if err != nil {
		log.Error().Err(err).Msg("Error handling html")
		return repository.Extraction{}, err
	}
	log.Info().Msgf("Uploaded html for url: %s, to GCS: %s", urlFrontier.Url, htmlUrl.B)

	extraction.ArtifactLink = &pdfUrl.B
	extraction.RawPageLink = &htmlUrl.B
	log.Info().Msgf("Scraped template for url: %s", urlFrontier.Url)

	return extraction, nil

}
