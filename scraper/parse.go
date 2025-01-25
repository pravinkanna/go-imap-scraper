package scraper

import (
	"strings"

	"github.com/antchfx/htmlquery"
)

func parseGenre(htmlStr string) ([]string, error) {
	var genres []string

	doc, err := htmlquery.Parse(strings.NewReader(htmlStr))
	if err != nil {
		panic(err)
	}

	// Find all news item.
	list, err := htmlquery.QueryAll(doc, xpaths["genre"])
	if err != nil {
		panic(err)
	}

	for _, n := range list {
		a := htmlquery.FindOne(n, "span")
		if a != nil {
			genre := htmlquery.InnerText(a)
			if strings.HasSuffix(genre, "0") {
				genre = genre[:len(genre)-1]
			}
			genres = append(genres, genre)
		}
	}

	return genres, nil
}
