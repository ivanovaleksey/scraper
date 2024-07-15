package scraper

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

const listScraperConcurrency = 5

type ListScraper struct {
	logger      *slog.Logger
	dataDir     string
	concurrency int
}

func NewListScraper(logger *slog.Logger) *ListScraper {
	return &ListScraper{
		logger:      logger,
		concurrency: listScraperConcurrency,
	}
}

func (ls *ListScraper) SetDataDir(dir string) {
	ls.dataDir = dir
}

// Run parses given page and paginates further.
// Given callback is called with detailed links from each page.
func (ls *ListScraper) Run(page string, detailedCb, assetsCb func([]string)) error {
	for {
		links, assets, nextPage, err := ls.parsePage(page)
		if err != nil {
			return err
		}
		detailedCb(links)
		assetsCb(assets)

		if nextPage == "" {
			break
		}
		page = path.Join(path.Dir(page), nextPage)
	}
	return nil
}

// RunMultiple parses given pages calling Run method concurrently.
func (ls *ListScraper) RunMultiple(pages []string, detailedCb, assetsCb func([]string)) error {
	ch := make(chan string, ls.concurrency)

	group, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < ls.concurrency; i++ {
		group.Go(func() error {
			for page := range ch {
				err := ls.Run(page, detailedCb, assetsCb)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	go func() {
		defer close(ch)
		for _, p := range pages {
			select {
			case <-ctx.Done():
				return
			case ch <- p:
			}
		}
	}()

	return group.Wait()
}

func (ls *ListScraper) parsePage(pageURL string) ([]string, []string, string, error) {
	ls.logger.Debug("parse list page", slog.String("url", pageURL))

	resp, err := http.Get(absURL(pageURL))
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil, "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	page, err := saveObject(ls.dataDir, pageURL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to save page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to build document: %w", err)
	}

	var (
		links  []string
		images []string
	)
	doc.Find("article.product_pod div.image_container a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		img, _ := s.Find("img").Attr("src")
		links = append(links, combinePath(pageURL, href))
		images = append(images, combinePath(pageURL, img))
		ls.logger.Debug("product href", slog.String("href", href), slog.String("img", img))
	})

	nextPage, _ := doc.Find("ul.pager li.next a").First().Attr("href")
	ls.logger.Debug("parsed href with page", slog.String("next_page", nextPage))
	return links, images, nextPage, nil
}
