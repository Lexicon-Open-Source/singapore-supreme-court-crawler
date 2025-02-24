package scrapper

import (
	"context"
	"lexicon/singapore-supreme-court-crawler/repository"

	"github.com/go-rod/rod"
)

func scrapeCaseSummary(ctx context.Context, e *rod.Element) (repository.Extraction, error) {

	var currentExtraction repository.Extraction

	return currentExtraction, nil
}
