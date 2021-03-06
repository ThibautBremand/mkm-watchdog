package scraper

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"mkm-watchdog/web"
	"strings"
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
	ExtraData string
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
			log.Printf("error: could not make request to URL %s: %s, skipping...", URL, err)
			continue
		}

		if doc == nil {
			log.Printf("error: received an empty document for URL %s, skipping...", URL)
			continue
		}

		res[URL] = parseDoc(doc, URL)

		// We space each queries just in case, to prevent getting throttled
		time.Sleep(4 * time.Second)
	}

	return res
}

func parseDoc(doc *goquery.Document, URL string) []Article {
	articles := make([]Article, 0)

	titleContainer := doc.Find("div.page-title-container")
	if titleContainer == nil {
		log.Printf("error: could not find titlecontainer for URL %s, skipping...", URL)
		return articles
	}

	titleHeader := titleContainer.Find("h1")
	if titleHeader == nil {
		log.Printf("error: could not find title header for URL %s, skipping...", URL)
		return articles
	}

	title := titleHeader.Text()
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

	// No need to check existence since condition is optional. For example there is no condition for sealed items.
	condition, _ := sel.Find("div.product-attributes").Find("a").Attr("title")

	language, exists := sel.Find("div.product-attributes").Find("span.icon").Attr("data-original-title")
	if !exists {
		return Article{}, fmt.Errorf("could not find language")
	}

	// No need to check existence since comment is optional. Some articles have an empty comment.
	comment, _ := sel.Find("div.product-comments").Find("span").Last().Attr("data-original-title")

	price := sel.Find("div.price-container").Text()

	quantity := sel.Find("span.item-count").Text()

	sellerName := sel.Find("span.seller-name").Find("a").Text()

	extraData := parseExtraData(sel)

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
		ExtraData: extraData,
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

func parseExtraData(sel *goquery.Selection) string {
	extraData := make([]string, 0)

	sel.Find("span.st_SpecialIcon").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		extraDataAttr, ok := sel.Attr("data-original-title")
		if !ok {
			// Do not break out of the loop since we need to check the other extra data available
			return true
		}

		extraData = append(extraData, extraDataAttr)
		return true
	})

	if len(extraData) == 0 {
		return ""
	}

	return strings.Join(extraData, ", ")
}

func (a Article) String() string {
	return fmt.Sprintf("%s %s %s %s", a.Name, a.URL, a.Price, a.Condition)
}
