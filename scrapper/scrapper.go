package scrapper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"lexicon/singapore-supreme-court-crawler/common"
	crawler_service "lexicon/singapore-supreme-court-crawler/crawler/services"
	"lexicon/singapore-supreme-court-crawler/scrapper/models"
	"lexicon/singapore-supreme-court-crawler/scrapper/services"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	mdp "github.com/JohannesKaufmann/html-to-markdown/plugin"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"

	gq "github.com/PuerkitoBio/goquery"

	"github.com/rs/zerolog/log"
)

func StartScraper() {
	// Fetch unscrapped url frontier from db
	list, err := crawler_service.GetUnscrappedUrlFrontiers(context.Background(), 100)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching unscrapped url frontier")
	}

	log.Info().Msg("Unscrapped URLs: " + strconv.Itoa(len(list)))

	q, err := queue.New(1, &queue.InMemoryQueueStorage{MaxSize: 100000})
	if err != nil {
		log.Error().Err(err).Msg("Error creating queue")
	}

	scrapper, err := buildScrapper()
	if err != nil {
		log.Error().Err(err).Msg("Error building scrapper")
	}

	for _, url := range list {
		// scrape url
		q.AddURL(url.Url)
	}
	q.Run(scrapper)
	scrapper.Wait()
}

func buildScrapper() (*colly.Collector, error) {
	var newExtraction models.Extraction
	c := colly.NewCollector(
		colly.AllowedDomains(common.CRAWLER_DOMAIN),
		// colly.Async(true),
		colly.MaxDepth(1),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  common.CRAWLER_DOMAIN,
		Delay:       time.Second * 2,
		RandomDelay: time.Second * 2,
	})

	c.SetRequestTimeout(time.Minute * 2)

	c.OnHTML("#divJudgement", func(e *colly.HTMLElement) {

		if e.ChildText("content") == "" {
			newExtraction = crawlOldTemplate(e)

		}
		newExtraction = crawlNewTemplate(e)

	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		var pdfUrl string
		if strings.Contains(url, "pdf") {
			pdfUrl = url
		}

		if pdfUrl != "" {
			pdfUrl = fmt.Sprintf("https://%s%s", common.CRAWLER_DOMAIN, pdfUrl)
			log.Info().Msg("Found PDF: " + pdfUrl)
			newExtraction.Metadata.PdfUrl = pdfUrl

		}
	})
	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	c.OnScraped(func(r *colly.Response) {
		frontierId := sha256.Sum256([]byte(r.Request.URL.String()))
		newExtraction.UrlFrontierId = hex.EncodeToString(frontierId[:])
		newExtraction.Id = hex.EncodeToString(frontierId[:])
		newExtraction.AddRawPageLink(r.Request.URL.String())
		newExtraction.UpdateUpdatedAt()

		if newExtraction.Metadata.PdfUrl != "" {
			log.Info().Msg("Uploading PDF to GCS: " + newExtraction.Metadata.PdfUrl)

			artifact, err := services.HandlePdf(newExtraction.Metadata, newExtraction.Metadata.PdfUrl, newExtraction.Metadata.Id+".pdf")
			if err != nil {
				log.Error().Err(err).Msg("Error handling pdf")
			}

			newExtraction.AddArtifactLink(artifact)
		}
		log.Info().Msg("Upserting extraction: " + newExtraction.Id)
		err := services.UpsertExtraction(newExtraction)
		if err != nil {
			log.Error().Err(err).Msg("Error upserting extraction")
		}
		log.Info().Msg("Updating url frontier status: " + newExtraction.UrlFrontierId)
		// err = crawler_service.UpdateUrlFrontierStatus(newExtraction.UrlFrontierId, crawler_model.URL_FRONTIER_STATUS_CRAWLED)
		// if err != nil {
		// 	log.Error().Err(err).Msg("Error updating url frontier status")
		// }

		log.Info().Msg("Scraped: " + r.Request.URL.String())
	})

	return c, nil
}

func crawlNewTemplate(e *colly.HTMLElement) models.Extraction {

	url := e.Request.URL.String()
	currentExtraction, err := services.GetCurrentExtraction(url)
	if err != nil {
		log.Error().Err(err).Msg("Error getting current extraction")
		return models.Extraction{}
	}

	e.ForEach("div.CaseNumber", func(i int, f *colly.HTMLElement) {
		currentExtraction.Metadata.Number = f.Text
	})

	title := currentExtraction.Metadata.Title
	splittedTitle := strings.Split(title, " v ")

	for _, part := range splittedTitle {
		if strings.Contains(strings.ToLower(part), "public prosecutor") {
			continue
		}

		currentExtraction.Metadata.Defendant = strings.TrimSpace(part)
	}

	e.ForEach("div.HN-Coram", func(i int, f *colly.HTMLElement) {

		splittedCoram := strings.Split(f.Text, "\n")

		for i, part := range splittedCoram {
			if i == 0 {
				courtSplit := strings.Split(part, "—")
				currentExtraction.Metadata.JudicalInstitution = strings.TrimSpace(courtSplit[0])
				continue
			}
			if i == 1 {
				currentExtraction.Metadata.Judges = strings.TrimSpace(part)
				continue
			}

		}
	})

	var rawVerdict []string
	var markdownVerdict []string
	converter := md.NewConverter("", true, nil)
	converter.Use(mdp.GitHubFlavored())

	converter.AddRules(
		md.Rule{
			Filter: []string{"div.Judg-Heading-1"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("## " + content)
			},
		},
		md.Rule{
			Filter: []string{"div.Judg-Heading-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("### " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judge-Quote-0"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("> " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judge-Quote-1"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String(">> " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judge-Quote-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String(">>> " + content)
			},
		},
	)
	e.ForEachWithBreak("#divJudgement > content > div.row > div.col.col-md-12.align-self-center", func(i int, f *colly.HTMLElement) bool {

		rawVerdict = append(rawVerdict, f.Text)
		html, err := f.DOM.Html()
		if err != nil {
			log.Error().Err(err).Msg("Error getting html")
			return false
		}
		md, err := converter.ConvertString(html)
		if err != nil {
			log.Error().Err(err).Msg("Error converting verdict")
			return false
		}

		markdownVerdict = append(markdownVerdict, md)

		return true
	})
	currentExtraction.Metadata.Verdict = strings.Join(rawVerdict, "\n")
	currentExtraction.Metadata.VerdictMarkdown = strings.Join(markdownVerdict, "\n")

	return currentExtraction

}

func crawlOldTemplate(e *colly.HTMLElement) models.Extraction {
	currentExtraction, err := services.GetCurrentExtraction(e.Request.URL.String())

	if err != nil {
		log.Error().Err(err).Msg("Error getting current extraction")
		return models.Extraction{}
	}
	e.ForEach("#info-table", func(i int, f *colly.HTMLElement) {
		f.ForEach("tr.info-row", func(i int, g *colly.HTMLElement) {
			key := g.ChildText("td.txt-label")
			value := g.ChildText("td.txt-body")

			if strings.Contains(key, "Tribunal/Court") {
				currentExtraction.Metadata.JudicalInstitution = value
			}
			if strings.Contains(key, "Counsel Name") {
				currentExtraction.Metadata.Judges = value
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
				currentExtraction.Metadata.Defendant = defendant
			}

			if strings.Contains(key, "Case Number") {
				currentExtraction.Metadata.Number = value
			}
		})
	})
	var rawVerdict []string
	var markdownVerdict []string
	converter := md.NewConverter("", true, nil)
	converter.Use(mdp.GitHubFlavored())

	converter.AddRules(
		md.Rule{
			Filter: []string{"p.Judg-Author"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("## " + content)
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
			Filter: []string{"p.Judg-Heading-1"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("## " + content)
			},
		},
		md.Rule{
			Filter: []string{"p.Judg-Heading-2"},
			Replacement: func(content string, selec *gq.Selection, options *md.Options) *string {

				content = strings.TrimSpace(content)
				return md.String("### " + content)
			},
		},
	)
	e.ForEachWithBreak("#divJudgement > div > p", func(i int, f *colly.HTMLElement) bool {
		if f.Attr("class") == "Footnote" {
			return false
		}

		rawVerdict = append(rawVerdict, f.Text)
		html, err := f.DOM.Html()
		if err != nil {
			log.Error().Err(err).Msg("Error getting html")
			return false
		}
		md, err := converter.ConvertString(html)
		if err != nil {
			log.Error().Err(err).Msg("Error converting verdict")
			return false
		}

		markdownVerdict = append(markdownVerdict, md)

		return true
	})

	currentExtraction.Metadata.Verdict = strings.Join(rawVerdict, "\n")
	currentExtraction.Metadata.VerdictMarkdown = strings.Join(markdownVerdict, "\n")

	return currentExtraction
}
