package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pravinkanna/imdb-scraper/scraper"
)

func main() {
	ctx := context.Background()

	scrapeID := createResultDirectory()

	// Fetch the list of Genres from IMDB
	genres, err := scraper.GetGenres()
	if err != nil {
		fmt.Println("Failed to fetch page")
	}

	searchConfig := getUserInput(genres)

	// Function to Scrape the movie details and store it to the SQLite DB
	err = scraper.ScrapeMovies(ctx, searchConfig, scrapeID)
	if err != nil {
		log.Fatalln("Error scraping movies:", err)
	}
}

// Function to Generate the scrapeID and create result directory
func createResultDirectory() string {
	// Create a result directory with a unique id
	t := time.Now()
	scrapeID := t.Format("20060102150405")
	resultPath := fmt.Sprintf("./results/%s/", scrapeID)
	if _, err := os.Stat(resultPath); os.IsNotExist(err) {
		err := os.Mkdir(resultPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create result directory: %+v", err)
		}
	}

	return scrapeID
}

// Function to get the filters from user (i.e) Genre and keyword
func getUserInput(genres []string) scraper.SearchConfig {

	// Get the Genre from user and validate against the fetched Genres
	var genre string
	fmt.Printf("Please select one of the Genre (case sensitive)\n\n")
	fmt.Printf("%s\n\n", strings.Join(genres, ", "))
	fmt.Printf("Your Selection (Default: search all genre): ")
	fmt.Scanln(&genre)
	isValid := slices.Contains(genres, genre)
	if genre != "" && !isValid {
		log.Fatalf("Not a valid genre. please try again restarting the script")
	}
	fmt.Println("Successfully selected genre:", genre)

	// Get the keyword to search from user
	var keyword string
	fmt.Printf("Enter a keyword to search (Default: None): ")
	fmt.Scanln(&keyword)
	fmt.Println("Given keyword: ", keyword)

	// Title is set as "feature" to filter only the movies
	searchConfig := scraper.SearchConfig{Title: "feature", Genre: genre, Keyword: keyword}
	return searchConfig
}
