package coordinator

import (
	"bytes"
	"fmt"
	"log"
	"mkm-watchdog/cache"
	"mkm-watchdog/config"
	"mkm-watchdog/scraper"
	"mkm-watchdog/web"
	"os"
	"strings"
	"text/template"
	"time"
)

type Coordinator struct {
	Scraper     *scraper.Scraper
	SleepPeriod time.Duration
	Tpl         *template.Template
}

func NewCoordinator(
	searches []config.Search,
	sleepPeriod time.Duration,
	tpl *template.Template,
) *Coordinator {
	URLs := make([]string, len(searches))
	for i, search := range searches {
		URLs[i] = search.URL
	}

	s := scraper.NewScraper(URLs)
	return &Coordinator{
		Scraper:     s,
		SleepPeriod: sleepPeriod,
		Tpl:         tpl,
	}
}

func (c *Coordinator) Start() {
	for {
		cached, err := cache.LoadCache()
		if err != nil {
			log.Fatalf("Could not load cache: %v", err)
		}

		m := c.Scraper.Scrape()

		newItems := keepNewItems(m, cached)

		log.Printf("Found %d new articles\n", len(newItems))

		if len(newItems) > 0 {
			sendToTelegram(newItems, c.Tpl)
		}

		refreshCache(m)

		time.Sleep(c.SleepPeriod)
	}
}

func keepNewItems(
	m map[string][]scraper.Article,
	cached map[string][]cache.CachedArticle,
) []scraper.Article {
	newArticles := make([]scraper.Article, 0)

	for URL, articles := range m {

		// Fill a temp map to easily detect articles that have already been scrapped
		cachedArticlesMap := make(map[string]int)
		for _, cachedArticle := range cached[URL] {
			cachedArticlesMap[cachedArticle.ID] = 1
		}

		for _, a := range articles {
			if _, ok := cachedArticlesMap[a.ID]; ok {
				// Ignore the article since it has already been scrapped
				continue
			}

			newArticles = append(newArticles, a)
		}
	}

	return newArticles
}

func refreshCache(m map[string][]scraper.Article) {
	toBeCached := buildCache(m)
	err := cache.UpdateCache(toBeCached)
	if err != nil {
		log.Println("error while writing cache, skipping", err)
	}
}

// buildCache takes a map[string][]scraper.Article and returns a ready to use map for the cache in the
// map[string][]cache.CachedArticle format.
func buildCache(m map[string][]scraper.Article) map[string][]cache.CachedArticle {
	res := make(map[string][]cache.CachedArticle)

	for key, articles := range m {
		res[key] = make([]cache.CachedArticle, len(articles))
		for i, a := range articles {
			res[key][i] = cache.CachedArticle{ID: a.ID}
		}
	}

	return res
}

func sendToTelegram(articles []scraper.Article, tpl *template.Template) {
	for _, article := range articles {
		buf := &bytes.Buffer{}
		err := tpl.Execute(buf, article)
		var msg string
		if err != nil {
			log.Println("could not execute template", err)
			msg = fmt.Sprintf("%s", article)
		} else {
			msg = buf.String()
		}

		// Double quotes are not correctly parsed by Telegram
		msg = strings.ReplaceAll(msg, `"`, "")

		err = web.SendTelegramMessage(
			os.Getenv("TELEGRAM_TOKEN"),
			os.Getenv("TELEGRAM_CHAT_ID"),
			msg,
		)
		if err != nil {
			log.Println("could not send Telegram message", err)
		}
	}
}