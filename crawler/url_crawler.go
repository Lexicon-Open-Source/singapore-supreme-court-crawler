package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"lexicon/singapore-supreme-court-crawler/common"
	"lexicon/singapore-supreme-court-crawler/crawler/models"
	"lexicon/singapore-supreme-court-crawler/crawler/services"

	scrapperModel "lexicon/singapore-supreme-court-crawler/scrapper/models"
	scrapperService "lexicon/singapore-supreme-court-crawler/scrapper/services"
	stdUrl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/golang-module/carbon/v2"
	"github.com/guregu/null"
	"github.com/rs/zerolog/log"
)

func StartCrawler() {

	startPage := 1

	url := "https://www.elitigation.sg/gd/Home/Index?filter=SUPCT&yearOfDecision=All&sortBy=DateOfDecision&currentPage=" + strconv.Itoa(startPage) + "&sortAscending=False&searchPhrase=CatchWords:%22Corruption%22&verbose=False"

	lastPage := getLastPage(url)

	totalData := (10 * lastPage) - ((startPage - 1) * 10)
	log.Info().Msg("Total Data: " + strconv.Itoa(totalData))
	for i := startPage; i <= lastPage; i++ {
		details := crawlUrl("https://www.elitigation.sg/gd/Home/Index?filter=SUPCT&yearOfDecision=All&sortBy=DateOfDecision&currentPage=" + strconv.Itoa(i) + "&sortAscending=False&searchPhrase=CatchWords:%22Corruption%22&verbose=False")

		log.Info().Msg("Details: " + strconv.Itoa(len(details)))
		allDetails := []models.UrlFrontier{}
		allExtraction := []scrapperModel.Extraction{}

		for _, detail := range details {

			id := sha256.Sum256([]byte(detail.Url))
			currentTime := carbon.Now().ToDateTimeStruct()
			allDetails = append(allDetails, models.UrlFrontier{
				Id:        hex.EncodeToString(id[:]),
				Url:       detail.Url,
				Domain:    common.CRAWLER_DOMAIN,
				Crawler:   common.CRAWLER_NAME,
				Status:    models.URL_FRONTIER_STATUS_NEW,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			})

			extractionData := scrapperModel.Extraction{
				Id:            hex.EncodeToString(id[:]),
				UrlFrontierId: hex.EncodeToString(id[:]),
				Language:      "en",
				Metadata: scrapperModel.Metadata{
					Id:             hex.EncodeToString(id[:]),
					Title:          detail.Metadata.Title,
					Number:         detail.Metadata.CaseNumber,
					CitationNumber: detail.Metadata.CitationNumber,
					Classification: detail.Metadata.CatchWords,
					Year:           strconv.Itoa(detail.Metadata.DecisionDate.Year()),
					DecisionDate:   detail.Metadata.DecisionDate.ToDateString(),
				},
				RawPageLink: null.StringFrom(detail.Url),
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			}
			allExtraction = append(allExtraction, extractionData)
		}

		err := services.UpsertUrl(allDetails)
		for _, ext := range allExtraction {
			err = scrapperService.UpsertExtraction(ext)

		}

		// insert temporary extraction data

		if err != nil {
			log.Error().Err(err).Msg("Error upserting url")
		}

		time.Sleep(time.Second * 2)
	}

}

func getLastPage(url string) int {
	lastPage := 0
	c := colly.NewCollector(
		colly.AllowedDomains("www.elitigation.sg"),
	)

	// find lastPage links
	c.OnHTML("#listview > div.row.justify-content-end > div > ul", func(e *colly.HTMLElement) {

		e.ForEach("li.page-item.page-link", func(i int, f *colly.HTMLElement) {

			lastUrl := f.ChildAttr("a", "href")
			if lastUrl == "" {
				return
			}

			u, err := stdUrl.Parse(lastUrl)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}

			q := u.Query()
			currentPageStr := q.Get("CurrentPage")
			currentPage, err := strconv.Atoi(currentPageStr)

			if err != nil {
				fmt.Println("Error converting CurrentPage to integer:", err)
				return
			}

			if lastPage < currentPage {
				lastPage = currentPage
			}
		})

	})

	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})
	c.Visit(url)
	return lastPage

}
func crawlUrl(url string) []models.CrawlerResult {
	log.Info().Msg("Crawling URL: " + url)
	c := colly.NewCollector(
		colly.AllowedDomains("www.elitigation.sg"),
	)
	c.SetRequestTimeout(time.Minute * 2)
	var detailUrls []models.CrawlerResult
	// find current page links
	c.OnHTML("#listview", func(e *colly.HTMLElement) {

		e.ForEach(".card.col-12", func(i int, f *colly.HTMLElement) {
			link := f.ChildAttr(" a.h5.gd-heardertext", "href")
			if link == "" {
				return
			}
			if !isDetailPage(link) {
				return
			}
			stringDate := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(f.ChildText("a.decision-date-link"), "Decision Date:", ""), "|", ""))
			decisionDate := carbon.ParseByFormat(stringDate, "j M Y").ToDateTimeStruct()
			detailUrls = append(detailUrls, models.CrawlerResult{
				Url: fmt.Sprintf("https://%s%s", common.CRAWLER_DOMAIN, link),
				Metadata: models.CrawlerMetadata{
					Title:          f.ChildText("a.h5.gd-heardertext"),
					CaseNumber:     f.ChildText("a.case-num-link"),
					CitationNumber: strings.TrimSpace(strings.ReplaceAll(f.ChildText("a.citation-num-link"), "|", "")),
					DecisionDate:   decisionDate,
					CatchWords:     f.ChildText("div.gd-catchword-container > a"),
				},
			})
		})

	})

	c.OnRequest(func(r *colly.Request) {
		log.Info().Msg("Visiting: " + r.URL.String())
	})

	c.Visit(url)

	return detailUrls
	//find pagination links

}

func isDetailPage(link string) bool {
	checker, err := regexp.Compile("/gd/s")
	if err != nil {
		log.Error().Err(err).Msg("Regex Compile Error")
		return false
	}

	return checker.MatchString(link)
}
