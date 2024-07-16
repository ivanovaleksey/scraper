## Scraper

Web scraper for https://books.toscrape.com

---

### Build
```shell
docker build -f Dockerfile -t scraper .
```

### Run
```shell
docker run -it -v $(pwd)/data:/app/data -p 8000:8000 scraper
```
