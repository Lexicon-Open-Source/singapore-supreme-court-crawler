package scrapper

import (
	"errors"
	"fmt"
	"lexicon/singapore-supreme-court-crawler/common"
	"strings"

	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func extractTitle(e *rod.Element) (string, error) {
	title, err := e.Element("h1")
	if err != nil {
		log.Error().Err(err).Msg("Error getting title")
		return "", err
	}

	text, err := title.Text()
	if err != nil {
		log.Error().Err(err).Msg("Error getting title")
		return "", err
	}

	return text, nil
}

func extractPdfUrl(e *rod.Element) (string, error) {
	hrefs, err := e.Elements("a[href]")
	if err != nil {
		log.Error().Err(err).Msg("Error getting href")
		return "", err
	}

	for _, href := range hrefs {
		attr, err := href.Attribute("href")
		if err != nil {
			log.Error().Err(err).Msg("Error getting href")
			return "", err
		}
		if strings.Contains(*attr, "pdf") {
			pdfUrl := fmt.Sprintf("https://%s%s", common.CRAWLER_DOMAIN, *attr)
			log.Info().Msg("Found PDF: " + pdfUrl)
			return pdfUrl, nil
		}
	}

	return "", errors.New("no pdf url found")

}
