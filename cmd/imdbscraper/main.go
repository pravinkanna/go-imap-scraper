package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/pravinkanna/imdb-scraper/scraper"
)

func main() {
	ctx := context.Background()

	// Fetch the Genres from IMDB
	genres, err := scraper.GetGenres()
	if err != nil {
		fmt.Println("Failed to fetch page")
	}

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
	err = scraper.ScrapeMovies(ctx, searchConfig)
	if err != nil {
		log.Fatalln("Error scraping movies:", err)
	}
}
