package scrapper

import (
	"fmt"
	"lexicon/singapore-supreme-court-crawler/common"
	"path"
	"strings"

	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func extractPdfUrl(e *rod.Element) (string, error) {
	hrefs, err := e.Elements("a[href]")
	if err != nil {
		return "", fmt.Errorf("failed to find href elements: %w", err)
	}

	for _, href := range hrefs {
		attr, err := href.Attribute("href")
		if err != nil {
			return "", fmt.Errorf("failed to get href attribute: %w", err)
		}
		if strings.Contains(*attr, "pdf") {
			// Sanitize the path to prevent path traversal
			cleanPath := path.Clean(*attr)
			if strings.HasPrefix(cleanPath, "..") {
				return "", fmt.Errorf("invalid path: %s", *attr)
			}
			pdfUrl := fmt.Sprintf("https://%s%s", common.CRAWLER_DOMAIN, cleanPath)
			log.Info().Msg("Found PDF: " + pdfUrl)
			return pdfUrl, nil
		}
	}

	return "", fmt.Errorf("no pdf url found in element")
}
