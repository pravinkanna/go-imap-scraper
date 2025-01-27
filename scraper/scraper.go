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
	var genre string
	fmt.Printf("Please select one of the Genre (case sensitive)\n\n")
	fmt.Printf("%s\n\n", strings.Join(genres, ", "))
	fmt.Printf("Your Selection (Default: search all genre): ")
	fmt.Scanln(&genre)
	isValid := slices.Contains(genres, genre)
	if genre != "" && !isValid {
		return SearchConfig{}, errors.New("Not a valid genre. please try again restarting the script")
	}
	fmt.Println("Successfully selected genre:", genre)

	// Get the keyword to search from user
	var keyword string
	fmt.Printf("Enter a keyword to search (Default: None): ")
	fmt.Scanln(&keyword)
	fmt.Println("Given keyword: ", keyword)

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
	var moviesUrl []string

	// Get teh HTML of the movie result page
	url := fmt.Sprintf("%s/search/title/?title_type=%s&genres=%s&keywords=%s", BASE_URL, c.Title, c.Genre, c.Keyword)
	searchResultHtml, err := getMovieSearchResult(ctx, url, maxResults)

	// Parse HTML with htmlquery
	doc, err := htmlquery.Parse(strings.NewReader(searchResultHtml))
	if err != nil {
		return moviesUrl, err
	}
	fmt.Println("Parsed the HTML:")

	// XPath to extract movie details
	movieNodes := htmlquery.Find(doc, `//*[@id="__next"]/main/div[2]/div[3]/section/section/div/section/section/div[2]/div/section/div[2]/div[2]/ul/li`)
	fmt.Println("No. of movies len:", len(movieNodes))

	counter := maxResults
	for _, movie := range movieNodes {
		href := ""

		// Extract movie details using XPath
		aNode, err := htmlquery.Query(movie, `/div/div/div/div[1]/div[2]/div[1]/a`)
		if err != nil {
			fmt.Println("Failed to get href of the movie")
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

	fmt.Println("All movie URLs Scraped!. Length:", len(moviesUrl))
	return moviesUrl, err
}

// Concurrently scrape movies data and store them
func ScrapeMovieAndStore(ctx context.Context, db sqlite.Service, concurrency int, moviesUrl []string) error {
	fmt.Println("Inside ScrapeMovieAndStore")

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
				fmt.Printf("Failed parsing movie: %+v", movie)
				continue
			}

			fmt.Printf("Parsed movie: %+v\n", movie)
			err = db.InsertMovie(movie.Name, movie.ReleasedYear, movie.Rating, movie.Summary, movie.Directors, movie.Cast)
			if err != nil {
				log.Panicf("Error inserting to SQLite: %+v", err)
			}

			if counter == 0 {
				close(htmlsCh)
			}
		}

		fmt.Println("Sent all data for parsing")
	}()

	// Concurrently scrape the movie data from URLs
	htmlsCh, err := getHtmlsConcurrent(moviesUrl, htmlsCh)
	if err != nil {
		return err
	}

	wg.Wait()
	fmt.Println("Done all data for parsing. Returning fn...")
	return nil
}
