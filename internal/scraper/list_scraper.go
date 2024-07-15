package scraper

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/PuerkitoBio/goquery"
)

type ListScraper struct {
	logger  *slog.Logger
	dataDir string
}

func NewListScraper(logger *slog.Logger) *ListScraper {
	return &ListScraper{
		logger: logger,
	}
}

func (ls *ListScraper) SetDataDir(dir string) {
	ls.dataDir = dir
}

// Run parses given page and paginates further.
// Given callback is called with detailed links from each page.
func (ls *ListScraper) Run(pageURL string, cb func([]string)) error {
	for {
		detailed, nextPage, err := ls.parsePage(pageURL)
		if err != nil {
			return err
		}
		cb(detailed)

		if nextPage == "" {
			break
		}
		pageURL = path.Join(path.Dir(pageURL), nextPage)
	}
	return nil
}

func (ls *ListScraper) parsePage(pageURL string) ([]string, string, error) {
	ls.logger.Debug("parse list page", slog.String("url", pageURL))

	resp, err := http.Get(absURL(pageURL))
	if err != nil {
		return nil, "", fmt.Errorf("failed to get page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	page, err := saveObject(ls.dataDir, pageURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to save page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return nil, "", fmt.Errorf("failed to build document: %w", err)
	}

	var links []string
	doc.Find("article.product_pod div.image_container a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		links = append(links, combinePath(pageURL, href))
		ls.logger.Debug("product href", slog.String("href", href))
	})

	// todo: remove
	if pageURL == "catalogue/category/books_1/page-10.html" {
		return nil, "", nil
	}

	nextPage, _ := doc.Find("ul.pager li.next a").First().Attr("href")
	ls.logger.Debug("parsed href with page", slog.String("next_page", nextPage))
	return links, nextPage, nil
}
