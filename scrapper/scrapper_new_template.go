package scrapper

import (
	"context"
	"lexicon/singapore-supreme-court-crawler/repository"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	mdp "github.com/JohannesKaufmann/html-to-markdown/plugin"
	gq "github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

func scrapeNewTemplate(ctx context.Context, e *rod.Element, extraction *repository.Extraction, urlFrontier *repository.UrlFrontier) error {

	log.Info().Msgf("Scraping new template for url: %s", urlFrontier.Url)

	extraction.Metadata.CitationNumber = urlFrontier.Metadata.CitationNumber
	extraction.Metadata.Numbers = urlFrontier.Metadata.CaseNumbers
	extraction.Metadata.Classifications = urlFrontier.Metadata.Categories
	year, err := time.Parse(time.RFC3339, urlFrontier.Metadata.DecisionDate)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing year")
		return err
	}
	extraction.Metadata.Year = year.Format(time.RFC3339)
	extraction.Metadata.DecisionDate = urlFrontier.Metadata.DecisionDate
	extraction.Metadata.Title = urlFrontier.Metadata.Title
	splittedTitle := strings.Split(extraction.Metadata.Title, " v ")

	for _, part := range splittedTitle {
		if strings.Contains(strings.ToLower(part), "public prosecutor") {
			continue
		}

		extraction.Metadata.Defendant = strings.TrimSpace(part)
	}
	corams, err := e.Elements("div.HN-Coram")
	if err != nil {
		log.Error().Err(err).Msg("Error getting coram")
		return err
	}

	for _, coram := range corams {
		splittedCoram := strings.Split(coram.MustText(), "\n")

		for i, part := range splittedCoram {
			if i == 0 {
				courtSplit := strings.Split(part, "—")
				extraction.Metadata.JudicalInstitution = strings.TrimSpace(courtSplit[0])
				continue
			}
			if i == 1 {
				extraction.Metadata.Judges = strings.TrimSpace(part)
				continue
			}

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

	verdicts, err := e.Elements("div.col.col-md-12.align-self-center")
	if err != nil {
		log.Error().Err(err).Msg("Error getting verdicts")
		return err
	}

	for _, verdict := range verdicts {

		rawVerdict = append(rawVerdict, verdict.MustText())
		html, err := verdict.HTML()
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
	log.Info().Msgf("Scraped new template for url: %s", urlFrontier.Url)

	return nil

}
