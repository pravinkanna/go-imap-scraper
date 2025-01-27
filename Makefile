# Simple Makefile for a Go project

# Run the application
run:
	@go run cmd/imdbscraper/main.go

# Build the application
build:
	@env GOOS=darwin GOARCH=arm64 go build -o imdb-scraper-darwin-arm64 cmd/imdbscraper/main.go
	@env GOOS=linux GOARCH=amd64 go build -o imdb-scraper-linux-amd64 cmd/imdbscraper/main.go
	@echo "Build Success"

# Clean the binaries
clean:
	@rm -f imdb-scraper-linux-amd64
	@rm -f imdb-scraper-darwin-arm64
	@echo "All Binaries cleared!"