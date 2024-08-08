## Scraper

### Assignment

Create a console program in your language of choice that:
- Traverses all pages on https://books.toscrape.com/
- Downloads and saves all files (pages, images...) to disk while keeping the file structure
- Shows some kind of progress information in the console

#### Definition of done
When your application has completed execution it should be possible to view the original
page locally on your computer.

#### Constraints
The focus of the challenge is the actual scraping. It’s OK to use external libraries for link
extraction and DOM parsing. Or you can roll your own if you’re feeling productive!

#### Good to know
On top of the basics, we do appreciate it if your program displays a good use of
asynchronicity, parallelism and threading.

---

### Build
```shell
docker build -f Dockerfile -t scraper .
```

### Run
```shell
docker run -it -v $(pwd)/data:/app/data -p 8000:8000 scraper
```
