package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"lexicon/singapore-supreme-court-crawler/common"
	"lexicon/singapore-supreme-court-crawler/crawler/models"
	"lexicon/singapore-supreme-court-crawler/crawler/services"
	"lexicon/singapore-supreme-court-crawler/repository"
	"sync"
	"time"

	stdUrl "net/url"
	"regexp"
	"strconv"

	"github.com/go-rod/rod"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type urlCrawler struct {
	BaseUrl        string
	Filter         string
	YearOfDecision string
	SortBy         string
	CurrentPage    int
	SortAscending  bool
	SearchPhrase   string
	Verbose        bool
}

func (u *urlCrawler) constructUrl() string {
	return fmt.Sprintf("%s?filter=%s&yearOfDecision=%s&sortBy=%s&currentPage=%d&sortAscending=%t&searchPhrase=%s&verbose=%t", u.BaseUrl, u.Filter, u.YearOfDecision, u.SortBy, u.CurrentPage, u.SortAscending, u.SearchPhrase, u.Verbose)
}

func (u *urlCrawler) copy() urlCrawler {
	return urlCrawler{
		BaseUrl:        u.BaseUrl,
		Filter:         u.Filter,
		YearOfDecision: u.YearOfDecision,
		SortBy:         u.SortBy,
		CurrentPage:    u.CurrentPage,
		SortAscending:  u.SortAscending,
		SearchPhrase:   u.SearchPhrase,
		Verbose:        u.Verbose,
	}
}

func newUrlCrawler(baseUrl string) (urlCrawler, error) {
	parsedUrl, err := stdUrl.Parse(baseUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing URL")
		return urlCrawler{}, err
	}

	base := fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, parsedUrl.Path)

	filter := parsedUrl.Query().Get("filter")
	yearOfDecision := parsedUrl.Query().Get("yearOfDecision")
	sortBy := parsedUrl.Query().Get("sortBy")
	sortAscending := parsedUrl.Query().Get("sortAscending")
	searchPhrase := parsedUrl.Query().Get("searchPhrase")
	verbose := parsedUrl.Query().Get("verbose")
	currentPage := parsedUrl.Query().Get("currentPage")

	currentPageInt, err := strconv.Atoi(currentPage)
	if err != nil {
		log.Error().Err(err).Msg("Error converting currentPage to integer")
		return urlCrawler{}, err
	}

	sortAscendingBool, err := strconv.ParseBool(sortAscending)
	if err != nil {
		log.Error().Err(err).Msg("Error converting sortAscending to boolean")
		return urlCrawler{}, err
	}

	verboseBool, err := strconv.ParseBool(verbose)
	if err != nil {
		log.Error().Err(err).Msg("Error converting verbose to boolean")
		return urlCrawler{}, err
	}

	return urlCrawler{
		BaseUrl:        base,
		Filter:         filter,
		YearOfDecision: yearOfDecision,
		SortBy:         sortBy,
		SortAscending:  sortAscendingBool,
		SearchPhrase:   searchPhrase,
		Verbose:        verboseBool,
		CurrentPage:    currentPageInt,
	}, nil

}

type Crawler interface {
	Setup()
	Teardown()
	CrawlAll(ctx context.Context) error
	Crawl(ctx context.Context, url string) error
	Consume(m jetstream.Msg) error
}

type CrawlerImpl struct {
	browser *rod.Browser
}

func (c *CrawlerImpl) Setup() {
	c.browser = rod.New().MustConnect()
}

func (c *CrawlerImpl) Teardown() {
	c.browser.MustClose()
}

func (c *CrawlerImpl) CrawlAll(ctx context.Context) error {
	startUrl := "https://www.elitigation.sg/gd/Home/Index?filter=SUPCT&yearOfDecision=All&sortBy=DateOfDecision&currentPage=1&sortAscending=False&searchPhrase=CatchWords:Corruption&verbose=False"

	pagePool := rod.NewPagePool(7)
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

	createListOfUrl := func(urlCrawler urlCrawler, startPage int, endPage int) []string {
		urls := []string{}
		for i := startPage; i <= endPage; i++ {
			newUrlCrawler := urlCrawler.copy()
			newUrlCrawler.CurrentPage = i
			urls = append(urls, newUrlCrawler.constructUrl())
		}
		return urls
	}

	job := func(urlPage string) error {

		page, err := pagePool.Get(create)
		if err != nil {
			log.Error().Err(err).Msg("Error getting page")
			return err
		}
		defer pagePool.Put(page)
		return c.crawlJudgement(ctx, page, urlPage)

	}
	parsedUrl, err := stdUrl.Parse(startUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing URL")
		return err
	}
	currentPage := parsedUrl.Query().Get("currentPage")
	if currentPage == "" {
		currentPage = "1"
	}
	currentPageInt, err := strconv.Atoi(currentPage)
	if err != nil {
		log.Error().Err(err).Msg("Error converting currentPage to integer")
		return err
	}

	rpLast, err := create()
	if err != nil {
		return fmt.Errorf("failed to create page for last page check: %w", err)
	}
	defer rpLast.Close()

	lastPage, err := getLastPage(ctx, rpLast, startUrl)
	if err != nil {
		return fmt.Errorf("failed to get last page: %w", err)
	}

	lastPageInt, totalResult := lastPage.Unpack()

	log.Info().Msg("Total result: " + strconv.Itoa(totalResult))

	urlCrawler, err := newUrlCrawler(startUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error creating url crawler")
		return err
	}

	urlList := createListOfUrl(urlCrawler, currentPageInt, lastPageInt)

	chunks := lo.Chunk(urlList, 7)

	var crawlErrors []error
	errChan := make(chan error, len(urlList)) // Channel to collect errors

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure all resources are cleaned up

	for _, urls := range chunks {
		// Check if context is cancelled before starting new chunk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wg := sync.WaitGroup{}

		// Process all URLs in the chunk concurrently
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()

				// Create a new context for this goroutine
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				default:
					if err := job(url); err != nil {
						errChan <- fmt.Errorf("error crawling %s: %w", url, err)
						// Optional: cancel other goroutines if you want to stop on first error
						// cancel()
					}
				}
			}(url)
		}

		wg.Wait()
	}

	// Close error channel
	close(errChan)

	// Collect any errors that occurred
	for err := range errChan {
		if err != nil {
			if err == context.Canceled {
				return fmt.Errorf("crawling was cancelled: %w", err)
			}
			crawlErrors = append(crawlErrors, err)
		}
	}

	// If there were any errors, return them combined
	if len(crawlErrors) > 0 {
		return fmt.Errorf("encountered %d errors during crawling: %v", len(crawlErrors), crawlErrors)
	}

	return nil
}

func (c *CrawlerImpl) Crawl(ctx context.Context, url string) error {
	page := c.browser.MustPage(url)

	c.crawlJudgement(ctx, page, url)

	return nil

}

func (c *CrawlerImpl) Consume(m jetstream.Msg) error {

	return nil
}

func (c *CrawlerImpl) crawlJudgement(ctx context.Context, rp *rod.Page, url string) error {
	log.Info().Msg("Crawling URL: " + url)

	// Check context before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	rpCtx := rp.Context(ctx)
	wait := rpCtx.MustWaitNavigation()
	err := rpCtx.Navigate(url)
	if err != nil {
		log.Error().Err(err).Msg("Error navigating to url")
		return err
	}
	wait()

	// Check context after navigation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	rp = rpCtx.MustWaitStable()

	elements, err := rp.Elements("#listview > div.row > div.card.col-12")
	if err != nil {
		log.Error().Err(err).Msg("Error getting elements")
		return err
	}

	log.Info().Msgf("Found %d elements", len(elements))

	detailUrls := []repository.UrlFrontier{}
	for _, element := range elements {

		crawlerResult, err := getElementContent(element)

		if err != nil {
			log.Error().Err(err).Msg("Error getting element content")
			continue
		}

		detailUrls = append(detailUrls, crawlerResult)
		log.Info().Msg("Crawler result: " + crawlerResult.Url)
	}

	allDetails := []repository.UrlFrontier{}

	for _, detail := range detailUrls {

		id := sha256.Sum256([]byte(detail.Url))
		currentTime := time.Now()
		allDetails = append(allDetails, repository.UrlFrontier{
			ID:        hex.EncodeToString(id[:]),
			Url:       detail.Url,
			Domain:    common.CRAWLER_DOMAIN,
			Crawler:   common.CRAWLER_NAME,
			Status:    models.URL_FRONTIER_STATUS_NEW,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Metadata:  detail.Metadata,
		})

	}
	log.Info().Msgf("Upserting %d urls", len(allDetails))
	err = services.UpsertUrl(ctx, allDetails)

	if err != nil {
		log.Error().Err(err).Msg("Error upserting url")
	}

	log.Info().Msgf("Crawling url: %s done!", url)
	return nil
}
func getElementContent(element *rod.Element) (repository.UrlFrontier, error) {
	link := element.MustElement("a.h5.gd-heardertext").MustAttribute("href")
	if link == nil {
		return repository.UrlFrontier{}, errors.New("link is empty")
	}
	if !isDetailPage(*link) {
		return repository.UrlFrontier{}, errors.New("link is not a detail page")
	}

	title, err := getTitle(element)
	if err != nil {
		log.Error().Err(err).Msg("Error getting title")
		return repository.UrlFrontier{}, err
	}
	caseNumbers, err := getCaseNumbers(element)
	if err != nil {
		log.Error().Err(err).Msg("Error getting case numbers")
		return repository.UrlFrontier{}, err
	}
	citationNumber, err := getCitationNumber(element)
	if err != nil {
		log.Error().Err(err).Msg("Error getting citation number")
		return repository.UrlFrontier{}, err
	}
	categories, err := getCategories(element)
	if err != nil {
		log.Error().Err(err).Msg("Error getting categories")
		return repository.UrlFrontier{}, err
	}

	decisionDate, err := getDecisionDate(element)
	if err != nil {
		log.Error().Err(err).Msg("Error getting decision date")
		return repository.UrlFrontier{}, err
	}
	detailUrls := repository.UrlFrontier{
		Url: fmt.Sprintf("https://%s%s", common.CRAWLER_DOMAIN, *link),
		Metadata: models.UrlFrontierMetadata{
			Title:          title,
			CaseNumbers:    caseNumbers,
			CitationNumber: citationNumber,
			DecisionDate:   decisionDate,
			Categories:     categories,
		},
	}

	return detailUrls, nil
}

func getLastPage(ctx context.Context, rp *rod.Page, url string) (lo.Tuple2[int, int], error) {
	lastPage := 0

	if err := rp.Context(ctx).Navigate(url); err != nil {
		if err == context.Canceled {
			return lo.Tuple2[int, int]{}, err
		}
		log.Error().Err(err).Msg("Error navigating to url")
		return lo.Tuple2[int, int]{}, err
	}
	totalResult := 0
	totalResultElement, err := rp.Element("#listview > div.row.justify-content-between.align-items-center > div.gd-csummary")
	if err != nil {
		log.Error().Err(err).Msg("Error getting total result element")
		return lo.Tuple2[int, int]{}, err
	}
	res, err := totalResultElement.HTML()
	if err != nil {
		log.Error().Err(err).Msg("Error getting total result element")
		return lo.Tuple2[int, int]{}, err
	}
	totalResult, err = strconv.Atoi(regexp.MustCompile(`\d+`).FindString(res))
	if err != nil {
		log.Error().Err(err).Msg("Error converting total result to integer")
		return lo.Tuple2[int, int]{}, err
	}

	elements, err := rp.Elements("#listview > div.row.justify-content-end > div > ul > li.page-item.page-link> a")
	if err != nil {
		log.Error().Err(err).Msg("Error getting elements")
		return lo.Tuple2[int, int]{}, err
	}
	for _, element := range elements {
		if element.MustHTML() == "" {
			continue
		}
		href := element.MustAttribute("href")
		if href == nil {
			continue
		}
		u, err := stdUrl.Parse(*href)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing URL")
			continue
		}
		lp := u.Query().Get("CurrentPage")
		lpInt, err := strconv.Atoi(lp)
		if err != nil {
			log.Error().Err(err).Msg("Error converting LP to integer")
			continue
		}
		if lastPage < lpInt {
			lastPage = lpInt
		}
	}

	return lo.Tuple2[int, int]{
		A: lastPage,
		B: totalResult,
	}, nil

}

func isDetailPage(link string) bool {
	checker, err := regexp.Compile("/gd/s")
	if err != nil {
		log.Error().Err(err).Msg("Regex Compile Error")
		return false
	}

	return checker.MatchString(link)
}
