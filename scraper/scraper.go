package scraper

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"mkm-watchdog/web"
	"time"
)

type Scraper struct {
	URLs []string
}

func NewScraper(URLs []string) *Scraper {
	return &Scraper{
		URLs: URLs,
	}
}

type Article struct {
	ID        string
	Name      string
	Condition string
	Language  string
	Comment   string
	Price     string
	Quantity  string
	URL       string
	Seller    Seller
}

type Seller struct {
	Name     string
	Details  *string
	Location *string
}

func (s *Scraper) Scrape() map[string][]Article {
	log.Println("Scraping new listings")

	res := make(map[string][]Article)

	for _, URL := range s.URLs {
		log.Printf("Handling url %s\n", URL)
		doc, err := web.Get(URL)
		if err != nil {
			log.Printf("could not make request to URL %s: %s\n", URL, err)
			continue
		}

		res[URL] = parseDoc(doc, URL)

		// We space each queries just in case, to prevent getting throttled
		time.Sleep(2 * time.Second)
	}

	return res
}

func parseDoc(doc *goquery.Document, URL string) []Article {
	articles := make([]Article, 0)

	title := doc.Find("div.page-title-container").Find("h1").Text()
	doc.Find("div.article-row").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		a, err := parseArticle(sel, title, URL)
		if err != nil {
			log.Printf("error while parsing article: %s\n", err)

			return false
		}

		articles = append(articles, a)

		return true
	})

	return articles
}

func parseArticle(
	sel *goquery.Selection,
	title string,
	URL string,
) (Article, error) {
	ID, exists := sel.Attr("id")
	if !exists {
		return Article{}, fmt.Errorf("could not find ID")
	}

	condition, exists := sel.Find("div.product-attributes").Find("a").Attr("title")
	if !exists {
		return Article{}, fmt.Errorf("could not find condition")
	}

	language, exists := sel.Find("div.product-attributes").Find("span.icon").Attr("data-original-title")
	if !exists {
		return Article{}, fmt.Errorf("could not find language")
	}

	comment, exists := sel.Find("div.product-comments").Find("span").Last().Attr("data-original-title")
	if !exists {
		log.Printf("No comment available for listing with ID %s for URL %s", ID, URL)
	}

	price := sel.Find("div.price-container").Text()

	quantity := sel.Find("span.item-count").Text()

	sellerName := sel.Find("span.seller-name").Find("a").Text()

	s := Seller{
		Name: sellerName,
	}

	a := Article{
		ID:        ID,
		Name:      title,
		Condition: condition,
		Language:  language,
		Comment:   comment,
		Price:     price,
		Quantity:  quantity,
		URL:       URL,
		Seller:    s,
	}

	sellerDetails, exists := sel.Find("span.seller-info").Find("span").Find(".badge").Attr("title")
	if !exists {
		log.Println("Could not find seller details, skipping...")
		return a, nil
	}

	sellerLocation, exists := sel.Find("span.seller-info").Find("span").Find(".icon").Attr("title")
	if !exists {
		log.Println("Could not find seller location, skipping...")
		return a, nil
	}

	a.Seller.Details = &sellerDetails
	a.Seller.Location = &sellerLocation

	return a, nil
}

func (a Article) String() string {
	return fmt.Sprintf("%s %s %s %s", a.Name, a.URL, a.Price, a.Condition)
}
