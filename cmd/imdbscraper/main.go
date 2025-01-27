package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pravinkanna/imdb-scraper/logging"
	"github.com/pravinkanna/imdb-scraper/scraper"
	"github.com/pravinkanna/imdb-scraper/sqlite"
)

const (
	concurrency = 1
	maxResults  = 100
)

func main() {
	// Initialize the logger at application startup
	err := logging.InitLogger("./logs")
	if err != nil {
		panic(err)
	}

	// Get logger instance
	logger := logging.GetLogger()
	logger.Info("Application started")

	ctx := context.Background()

	// Initialize DB
	db, err := sqlite.Connect()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize sqlite. Error: %+v", err))
		panic(err)
	}
	defer logger.Info("Successfully close sqlite connection")
	defer db.Close()
	logger.Info("Successfully Connected to the database")

	err = db.CreateTable()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create tables in sqlite. Error: %+v", err))
		panic(err)
	}
	logger.Info("Successfully created movies table in Sqlite")

	// Fetch the list of Genres from IMDB
	genres, err := scraper.GetGenres()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get genres. Error: %+v", err))
		panic(err)
	}

	// Get the filters from user as input (i.e)
	searchConfig, err := scraper.GetInput(genres)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get input. Error: %+v", err))
		panic(err)
	}

	// Fetch the list of movies matching the user input
	moviesUrl, err := scraper.GetMoviesUrl(ctx, searchConfig, maxResults)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get movies. Error: %+v", err))
		panic(err)
	}
	if len(moviesUrl) <= 0 {
		logger.Warn("No movies matched your query. Please try a different query")
		fmt.Println("No Movie matched your query. Try again later")
		os.Exit(0)
	}

	// Concurrently Scrape the movie data and store it to the SQLite DB
	err = scraper.ScrapeMovieAndStore(ctx, db, concurrency, moviesUrl)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to scrape movie and store. Error: %+v", err))
		panic(err)
	}

	logger.Info("Scraping process completed")
	fmt.Println("\nScraping Successfully Completed.")
}
