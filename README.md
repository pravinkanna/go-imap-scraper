# IMDb Movie Scraper

A web scraper built in Go that extracts movie information from IMDb, handling pagination and storing data in SQLite Database.

## Features

- **Search movies** by genre or keyword.
- **Extract detailed movie information** including:
  - Title
  - Year
  - Rating
  - Directors
  - Cast
  - Summary
- **Handles pagination** for comprehensive data collection.
- **Stores data** in a SQL format for easy access and manipulation.

## Prerequisites

- **Go** version 1.23 or higher (tested on this version).
- **Chrome/Chromium browser** (required for Chromedp).

## Installation

1. Clone the repository:

```bash
git clone https://github.com/pravinkanna/go-imdb-scraper
cd imdb-scraper
```

2. Install dependencies:

```bash
go mod tidy
```

## Usage

1. Run the scraper:

```bash
make run
```

2. Set scrape config in `cmd/imdbscraper/main.go`:

```go
const (
	concurrency = 1
	maxResults  = 170
)
```

## Bonus Features

- **Concurrency with Goroutines**: Enhances performance by scraping multiple pages simultaneously.
- **Rate Limiting Awareness**: Includes sleep intervals to reduce the risk of being rate-limited or IP-blocked (note: My IP got blocked in the process. Had to use VPN after that).
- **Robust Error Handling**: Accounts for missing or incomplete parameters (e.g., movies without ratings).
- **User Configurable Parameters**: Allows dynamic input for search criteria.
