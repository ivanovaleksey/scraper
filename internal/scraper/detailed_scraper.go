package scraper

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/PuerkitoBio/goquery"
)

type DetailedScraper struct {
	logger  *slog.Logger
	dataDir string

	pagesCache ItemsCache
	mediaCache MediaCache
}

type MediaCache interface {
	Set(string)
	Has(string) bool
}

type ItemsCache interface {
	SetMulti(keys ...string)
	GetAll() []string
}

func NewDetailedScraper(logger *slog.Logger) *DetailedScraper {
	return &DetailedScraper{
		logger:     logger,
		pagesCache: NewCache(),
		mediaCache: NewCache(),
	}
}

func (ds *DetailedScraper) AddLinks(urls ...string) {
	ds.pagesCache.SetMulti(urls...)
}

// todo: add concurrency
func (ds *DetailedScraper) Run() error {
	links := ds.pagesCache.GetAll()
	ds.logger.Info("detailed scraper started", slog.Int("links", len(links)))

	// todo: all links
	for _, link := range links[:10] {
		err := ds.parsePage(link)
		if err != nil {
			return fmt.Errorf("failed to parse detailed page %q': %w", link, err)
		}
	}
	return nil
}

func (ds *DetailedScraper) SetDataDir(dir string) {
	ds.dataDir = dir
}

func (ds *DetailedScraper) parsePage(pageURL string) error {
	ds.logger.Debug("parse detailed page", slog.String("url", pageURL))

	page, err := saveObject(ds.dataDir, pageURL)
	if err != nil {
		return fmt.Errorf("failed to save page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page))
	if err != nil {
		return fmt.Errorf("failed to build document: %w", err)
	}

	var images []string
	doc.Find("div#product_gallery img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		images = append(images, combinePath(pageURL, src))
		ds.logger.Debug("product detailed img", slog.String("src", src), slog.String("page", pageURL))
	})

	for _, image := range images {
		if ds.mediaCache.Has(image) {
			ds.logger.Warn("cache hit", slog.String("image", image))
		} else {
			_, err = saveObject(ds.dataDir, image)
			if err != nil {
				return fmt.Errorf("failed to save image: %w", err)
			}
			ds.mediaCache.Set(image)
		}
	}

	return nil
}
