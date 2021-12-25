package coordinator

import (
	"mkm-watchdog/cache"
	"mkm-watchdog/scraper"
	"reflect"
	"testing"
)

func TestKeepNewItems(t *testing.T) {
	a1 := scraper.Article{
		ID:        "id-1",
		Name:      "name-1",
		Condition: "condition-1",
		Language:  "language-1",
		Comment:   "comment-1",
		Price:     "price-1",
		Quantity:  "quantity-1",
		URL:       "url-1",
		Seller: scraper.Seller{
			Name:     "name-1",
			Details:  nil,
			Location: nil,
		},
	}

	a2 := scraper.Article{
		ID:        "id-2",
		Name:      "name-2",
		Condition: "condition-2",
		Language:  "language-2",
		Comment:   "comment-2",
		Price:     "price-2",
		Quantity:  "quantity-2",
		URL:       "url-2",
		Seller: scraper.Seller{
			Name:     "name-2",
			Details:  nil,
			Location: nil,
		},
	}

	m := map[string][]scraper.Article{
		"1": {
			a1,
		},
		"2": {
			a2,
		},
	}

	cached := map[string][]cache.CachedArticle{
		"1": {
			cache.CachedArticle{ID: "id-1"},
		},
	}

	got := keepNewItems(m, cached)
	exp := []scraper.Article{a2}

	if !reflect.DeepEqual(exp, got) {
		t.Errorf("expected %v but got %v", exp, got)
	}
}
