package scraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/pravinkanna/imdb-scraper/logging"
	"github.com/pravinkanna/imdb-scraper/sqlite"
)

// SearchConfig store the information about the config for the search
type SearchConfig struct {
	Title   string
	Genre   string
	Keyword string
}

// Movie stores information about a imdb movie
type Movie struct {
	ID           string
	Name         string
	ReleasedYear string
	Rating       string
	Directors    []string
	Cast         []string
	Summary      string
}

var BASE_URL = "https://www.imdb.com"

// Function to get the filters from user (i.e) Genre and keyword
func GetInput(genres []string) (SearchConfig, error) {
	// Get the Genre from user and validate against the fetched Genres
	genre := ""
	fmt.Printf("Please select one of the Genre:\n")
	fmt.Printf("%s\n\n", strings.Join(genres, ", "))
	fmt.Printf("Genre (Optional, Case-sensitive): ")
	fmt.Scanln(&genre)
	isValid := slices.Contains(genres, genre)
	if genre != "" && !isValid {
		return SearchConfig{}, errors.New("Not a valid genre. please try again restarting the script")
	}

	// Get the keyword to search from user
	keyword := ""
	fmt.Printf("Keyword to search (Optional): ")
	fmt.Scanln(&keyword)

	// Title is set as "feature" to filter only the movies
	return SearchConfig{Title: "feature", Genre: genre, Keyword: keyword}, nil
}

// Returns the list of Genres from the IMDB website
func GetGenres() ([]string, error) {
	searchPageURL := fmt.Sprintf("%s/search/title", BASE_URL)
	htmlStr, err := getHtml(searchPageURL)
	if err != nil {
		return []string{}, err
	}

	genres, err := parseGenre(htmlStr)
	if err != nil {
		return []string{}, err
	}
	return genres, nil
}

// Returns the individual movies url matching the search query
func GetMoviesUrl(ctx context.Context, c SearchConfig, maxResults int) ([]string, error) {
	logger := logging.GetLogger()
	var moviesUrl []string

	// Get teh HTML of the movie result page
	url := fmt.Sprintf("%s/search/title/?title_type=%s&genres=%s&keywords=%s", BASE_URL, c.Title, c.Genre, c.Keyword)
	searchResultHtml, err := getMovieSearchResult(ctx, url, maxResults)

	// Parse HTML with htmlquery
	doc, err := htmlquery.Parse(strings.NewReader(searchResultHtml))
	if err != nil {
		return moviesUrl, err
	}

	// XPath to extract movie details
	movieNodes := htmlquery.Find(doc, `//*[@id="__next"]/main/div[2]/div[3]/section/section/div/section/section/div[2]/div/section/div[2]/div[2]/ul/li`)
	logger.Info(fmt.Sprintf("No. of movies: %d", len(movieNodes)))

	counter := maxResults
	for _, movie := range movieNodes {
		href := ""

		// Extract movie details using XPath
		aNode, err := htmlquery.Query(movie, `/div/div/div/div[1]/div[2]/div[1]/a`)
		if err != nil {
			logger.Warn("Failed to get href of the movie")
			continue
		}
		if aNode != nil {
			href = htmlquery.SelectAttr(aNode, "href")
			movieLink := fmt.Sprintf("%s/%s", BASE_URL, href)
			moviesUrl = append(moviesUrl, movieLink)
			counter -= 1
		}
		if counter == 0 {
			break
		}
	}

	logger.Info(fmt.Sprintf("All movie URLs Scraped!. Length: %d", len(moviesUrl)))
	return moviesUrl, err
}

// Concurrently scrape movies data and store them
func ScrapeMovieAndStore(ctx context.Context, db sqlite.Service, concurrency int, moviesUrl []string) error {
	logger := logging.GetLogger()
	htmlsCh := make(chan string, concurrency)
	var wg sync.WaitGroup
	wg.Add(len(moviesUrl))

	// Goroutine to parse and store the Movie data
	go func() {
		counter := len(moviesUrl)
		for htmlStr := range htmlsCh {
			defer wg.Done()
			counter -= 1

			// Parse the movie data from HTML
			movie, err := parseMovie(htmlStr)
			if err != nil {
				logger.Warn(fmt.Sprintf("Failed parsing movie: %+v", movie))
				continue
			}

			logger.Info(fmt.Sprintf("Successfully Parsed movie: %s", movie.Name))
			err = db.InsertMovie(movie.Name, movie.ReleasedYear, movie.Rating, movie.Summary, movie.Directors, movie.Cast)
			if err != nil {
				log.Panicf("Error inserting to SQLite: %+v", err)
			}
			logger.Info(fmt.Sprintf("Successfully inserted movie: %s", movie.Name))

			if counter == 0 {
				close(htmlsCh)
			}
		}
	}()

	// Concurrently scrape the movie data from URLs
	htmlsCh, err := getHtmlsConcurrent(moviesUrl, htmlsCh)
	if err != nil {
		return err
	}

	wg.Wait()
	logger.Info("Successfully parsed and stored all the movies data")
	return nil
}
