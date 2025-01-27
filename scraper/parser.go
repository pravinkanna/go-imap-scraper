package scraper

import (
	"strings"

	"github.com/antchfx/htmlquery"
)

var xpaths = map[string]string{
	"genre":         `//*[@id="accordion-item-genreAccordion"]/div/section/button`,
	"name":          `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[2]/div[1]/h1/span`,
	"rating":        `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[3]/div[2]/div[2]/div[1]/div/div[1]/a/span/div/div[2]/div[1]/span[1]`,
	"released_year": `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[2]/div[1]/ul/li[1]/a`,
	"summary":       `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[3]/div[2]/div[1]/section/p/span[1]`,
	"directors":     `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[3]/div[2]/div[2]/div[2]/div/ul/li[1]/div/ul/li`,
	"cast":          `//*[@id="__next"]/main/div/section[1]/section/div[3]/section/section/div[3]/div[2]/div[1]/section/div[2]/div/ul/li[3]/div/ul/li`,
}

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

func parseMovie(htmlStr string) (Movie, error) {
	name, rating, releasedYear, summary := "", "", "", ""
	directors, cast := []string{}, []string{}

	doc, err := htmlquery.Parse(strings.NewReader(htmlStr))
	if err != nil {
		panic(err)
	}

	// Parse the Movie name
	nameNode, err := htmlquery.QueryAll(doc, xpaths["name"])
	if err != nil {
		panic(err)
	}
	if len(nameNode) > 0 {
		name = htmlquery.InnerText(nameNode[0])
	}

	// Parse the Movie name
	ratingNode, err := htmlquery.QueryAll(doc, xpaths["rating"])
	if err != nil {
		panic(err)
	}
	if len(ratingNode) > 0 {
		rating = htmlquery.InnerText(ratingNode[0])
	}

	// Parse the Movie name
	releasedNode, err := htmlquery.QueryAll(doc, xpaths["released_year"])
	if err != nil {
		panic(err)
	}
	if len(releasedNode) > 0 {
		releasedYear = htmlquery.InnerText(releasedNode[0])
	}

	// Parse the Movie name
	summaryNode, err := htmlquery.QueryAll(doc, xpaths["summary"])
	if err != nil {
		panic(err)
	}
	if len(summaryNode) > 0 {
		summary = htmlquery.InnerText(summaryNode[0])
	}

	// Iterate the Directors node and get all the directors
	directosNodeList, err := htmlquery.QueryAll(doc, xpaths["directors"])
	if err != nil {
		panic(err)
	}
	if len(directosNodeList) > 0 {
		for _, n := range directosNodeList {
			a := htmlquery.FindOne(n, "a")
			if a != nil {
				director := htmlquery.InnerText(a)
				directors = append(directors, director)
			}
		}
	}

	// Iterate the Cast node and get all the cast members
	castNodeList, err := htmlquery.QueryAll(doc, xpaths["cast"])
	if err != nil {
		panic(err)
	}
	if len(castNodeList) > 0 {
		for _, n := range castNodeList {
			a := htmlquery.FindOne(n, "a")
			if a != nil {
				star := htmlquery.InnerText(a)
				cast = append(cast, star)
			}
		}
	}

	movie := Movie{
		Name:         name,
		Rating:       rating,
		ReleasedYear: releasedYear,
		Summary:      summary,
		Directors:    directors,
		Cast:         cast,
	}
	return movie, nil
}
