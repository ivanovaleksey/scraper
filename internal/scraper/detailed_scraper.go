package scraper

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

const detailedScraperConcurrency = 5

type DetailedScraper struct {
	logger      *slog.Logger
	dataDir     string
	concurrency int

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
		logger:      logger,
		concurrency: detailedScraperConcurrency,
		pagesCache:  NewCache(),
		mediaCache:  NewCache(),
	}
}

func (ds *DetailedScraper) AddLinks(urls ...string) {
	ds.pagesCache.SetMulti(urls...)
}

func (ds *DetailedScraper) Run() error {
	links := ds.pagesCache.GetAll()
	ds.logger.Info("detailed scraper started", slog.Int("links", len(links)))

	ch := make(chan string, ds.concurrency)

	group, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < ds.concurrency; i++ {
		group.Go(func() error {
			for link := range ch {
				err := ds.parsePage(link)
				if err != nil {
					return fmt.Errorf("failed to parse detailed page %q': %w", link, err)
				}
			}
			return nil
		})
	}
	go func() {
		defer close(ch)
		for _, link := range links {
			select {
			case <-ctx.Done():
				return
			case ch <- link:
			}
		}
	}()

	return group.Wait()
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
