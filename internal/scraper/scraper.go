package scraper

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const host = "https://books.toscrape.com"

type Service struct {
	logger  *slog.Logger
	dataDir string

	listScraper     *ListScraper
	detailedScraper *DetailedScraper
}

func NewService(logger *slog.Logger) *Service {
	return &Service{
		logger:          logger,
		listScraper:     NewListScraper(logger),
		detailedScraper: NewDetailedScraper(logger),
	}
}

// todo: save main page
// todo: indicate progress (might have multiple: categories and detailed pages)
func (srv *Service) Run() error {
	var err error
	srv.dataDir, err = os.MkdirTemp("data", "run_*")
	if err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	srv.listScraper.SetDataDir(srv.dataDir)
	srv.detailedScraper.SetDataDir(srv.dataDir)

	resp, err := http.Get(host)
	if err != nil {
		return fmt.Errorf("failed to get page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to build document: %w", err)
	}

	err = srv.saveStatics(doc)
	if err != nil {
		return fmt.Errorf("failed to save statics: %w", err)
	}

	{
		detailedCallback := func(links []string) {
			srv.detailedScraper.AddLinks(links...)
		}
		assetsCallback := func(links []string) {
			srv.detailedScraper.AddAssets(links...)
		}
		err = srv.listScraper.Run("index.html", detailedCallback, assetsCallback)
		if err != nil {
			return fmt.Errorf("failed to paginate main page: %w", err)
		}

		err = srv.parseCategories(doc, detailedCallback, assetsCallback)
		if err != nil {
			return err
		}
	}

	err = srv.detailedScraper.Run()
	if err != nil {
		return fmt.Errorf("failed to parse detailed pages: %w", err)
	}

	return nil
}

func (srv *Service) parseCategories(doc *goquery.Document, cb1, cb2 func([]string)) error {
	var categories []string
	doc.Find("div.side_categories li a").Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("href")
		categories = append(categories, value)
	})
	srv.logger.Debug("got categories", slog.Int("num", len(categories)))

	err := srv.listScraper.RunMultiple(categories, cb1, cb2)
	if err != nil {
		return fmt.Errorf("failed to parse categories: %w", err)
	}
	return nil
}

func (srv *Service) saveStatics(doc *goquery.Document) error {
	var links []string
	doc.Find("link[type='text/css']").Each(func(i int, selection *goquery.Selection) {
		href, _ := selection.Attr("href")
		links = append(links, href)
		srv.logger.Debug("static href", slog.String("href", href))
	})

	for _, link := range links {
		_, err := saveObject(srv.dataDir, link)
		if err != nil {
			return fmt.Errorf("failed to save static file: %w", err)
		}
	}

	return nil
}

func saveObject(dataDir, url string) ([]byte, error) {
	resp, err := http.Get(absURL(url))
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	dir, base := path.Split(url)
	subDir := path.Join(dataDir, dir)
	err = os.MkdirAll(subDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create subdir: %w", err)
	}
	err = os.WriteFile(path.Join(subDir, base), body, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return body, nil
}

func absURL(relURL string) string {
	return host + "/" + relURL
}

func combinePath(basePath, relPath string) string {
	parts := strings.Split(relPath, "/")
	numBacks := 0
	var clearedParts []string
	for _, part := range parts {
		if part == ".." {
			numBacks++
			continue
		}
		clearedParts = append(clearedParts, part)
	}
	if numBacks == 0 {
		return path.Join(path.Dir(basePath), relPath)
	}
	dirParts := strings.Split(path.Dir(basePath), "/")
	pathParts := dirParts[:len(dirParts)-numBacks]
	pathParts = append(pathParts, clearedParts...)
	return path.Join(pathParts...)
}
