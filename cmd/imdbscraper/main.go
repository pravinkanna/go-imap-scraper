package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pravinkanna/imdb-scraper/scraper"
	"github.com/pravinkanna/imdb-scraper/sqlite"
)

func main() {
	concurrency := 1
	ctx := context.Background()

	// Initialize DB
	db, err := sqlite.Connect()
	if err != nil {
		log.Panicf("Failed to initialize sqlite. Error: %+v", err)
	}
	defer db.Close()

	err = db.CreateTable()
	if err != nil {
		log.Panicf("Failed to create tables in sqlite. Error: %+v", err)
	}

	// Fetch the list of Genres from IMDB
	genres, err := scraper.GetGenres()
	if err != nil {
		log.Panicf("Failed to get genres. Error: %+v", err)
	}

	// Get the filters from user as input (i.e)
	searchConfig, err := scraper.GetInput(genres)
	if err != nil {
		log.Panicf("Failed to get input. Error: %+v", err)
	}

	// Fetch the list of movies matching the user input
	moviesUrl, err := scraper.GetMoviesUrl(ctx, searchConfig)
	if err != nil {
		log.Panicf("Failed to get movies. Error: %+v", err)
	}
	if len(moviesUrl) <= 0 {
		log.Fatalln("No movies matched your query. Please try a different query")
	}

	// Concurrently Scrape the movie data and store it to the SQLite DB
	err = scraper.ScrapeMovieAndStore(ctx, db, concurrency, moviesUrl)
	if err != nil {
		log.Panicf("Failed to scrape movie and store. Error: %+v", err)
	}

	fmt.Println("Scraping process completed. Exiting...")
}
