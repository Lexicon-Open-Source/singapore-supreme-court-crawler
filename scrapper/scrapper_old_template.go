package scrapper

import (
	"context"
	"errors"
	"lexicon/singapore-supreme-court-crawler/repository"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	mdp "github.com/JohannesKaufmann/html-to-markdown/plugin"
	gq "github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func scrapeOldTemplate(ctx context.Context, e *rod.Element, extraction *repository.Extraction, urlFrontier *repository.UrlFrontier) error {

	select {
	case <-ctx.Done():
		return errors.New("context canceled")
	default:
	}
	log.Info().Msgf("Scraping old template for url: %s", urlFrontier.Url)
	extraction.Metadata.CitationNumber = urlFrontier.Metadata.CitationNumber
	extraction.Metadata.Numbers = urlFrontier.Metadata.CaseNumbers
	extraction.Metadata.Classifications = urlFrontier.Metadata.Categories
	year, err := time.Parse(time.RFC3339, urlFrontier.Metadata.DecisionDate)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing year")
		return err
	}
	extraction.Metadata.Year = year.Format("2006")
	extraction.Metadata.DecisionDate = urlFrontier.Metadata.DecisionDate
	extraction.Metadata.Title = urlFrontier.Metadata.Title

	infoTable, err := e.Element("#info-table")
	if err != nil {
		log.Error().Err(err).Msg("Error finding info table")
		return err
	}

	row := infoTable.MustElements("tr.info-row")

	for _, r := range row {
		key := r.MustElement("td.txt-label").MustText()
		value := r.MustElement("td.txt-body").MustText()

		if strings.Contains(key, "Tribunal/Court") {
			extraction.Metadata.JudicalInstitution = value
		}
		if strings.Contains(key, "Coram") {
			extraction.Metadata.Judges = value
		}
		if strings.Contains(key, "Counsel Name") {
			extraction.Metadata.Counsel = value
		}
		if strings.Contains(key, "Parties") {
			extractedParties := strings.Split(value, "—")
			var defendant string
			for _, party := range extractedParties {
				if strings.Contains(strings.ToLower(party), "public prosecutor") {
					continue
				}
				defendant = strings.TrimSpace(party)
			}
			extraction.Metadata.Defendant = defendant
		}

	}

	var rawVerdict []string
	var markdownVerdict []string
	converter := md.NewConverter("", true, nil)
	converter.Use(mdp.GitHubFlavored())

	converter.AddRules(
		md.Rule{
			Filter: []string{"p.Judg-Author"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("** " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-Heading-1"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("# " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-Heading-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("## " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-1", "p.Judg-2", "p.Judg-List-1-Item", "p.Judg-List-2-Item"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String(content)
			},
		},

		md.Rule{
			Filter: []string{"p.Judge-Quote-1"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("> " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judge-Quote-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String(">> " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-QuoteList-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("- " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-QuoteList-3"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("  - " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Footnote"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("^" + content)

			},
		},
		md.Rule{
			Filter: []string{"p.Judg-List-1-No"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("1. " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-List-2-No"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("a. " + content)
			},
		},
	)
	content, err := e.Elements("div > p")
	if err != nil {
		log.Error().Err(err).Msg("Error getting content")
		return err
	}
	for _, c := range content {
		rawVerdict = append(rawVerdict, c.MustText())
		html, err := c.HTML()
		if err != nil {
			log.Error().Err(err).Msg("Error getting html")
			continue
		}
		md, err := converter.ConvertString(html)
		if err != nil {
			log.Error().Err(err).Msg("Error converting verdict")
			continue
		}

		markdownVerdict = append(markdownVerdict, md)

	}

	extraction.Metadata.Verdict = strings.Join(rawVerdict, "\n")
	extraction.Metadata.VerdictMarkdown = strings.Join(markdownVerdict, "\n")
	log.Info().Msgf("Scraped old template for url: %s", urlFrontier.Url)

	return nil
}
