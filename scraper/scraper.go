package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/chromedp"
)

// SearchConfig store the information about the config for the search
type SearchConfig struct {
	Title   string
	Genre   string
	Keyword string
}

// Movie stores information about a imdb movie
type Movie struct {
	ImdbID     string
	Released   string
	Creator    string
	Level      string
	URL        string
	Language   string
	Commitment string
	Rating     string
}

var xpaths = map[string]string{
	"genre": `//*[@id="accordion-item-genreAccordion"]/div/section/button`,
}

// Returns the list of Genres from the IMDB website
func GetGenres() ([]string, error) {
	url := "https://www.imdb.com/search/title"
	htmlStr, err := getHtml(url)
	if err != nil {
		return []string{}, err
	}
	genres, err := parseGenre(htmlStr)
	if err != nil {
		return []string{}, err
	}
	return genres, nil
}

func ScrapeMovies(ctx context.Context, c SearchConfig) error {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts,
		chromedp.Flag("headless", false),       // Disable headless mode
		chromedp.Flag("disable-gpu", false),    // Enable GPU (optional)
		chromedp.Flag("start-maximized", true), // Start with a maximized window
		chromedp.Flag("auto-open-devtools-for-tabs", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create chromedp context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Create a new context with timeout
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var htmlContent string
	var err error

	// Navigate to page and extract HTML
	url := fmt.Sprintf("https://www.imdb.com/search/title/?title_type=%s&genres=%s&keywords=%s", c.Title, c.Genre, c.Keyword)
	fmt.Println("Loading URL:", url)

	// var res string
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),

		chromedp.WaitVisible(`.ipc-see-more__button`),

		// Click "Load More" button repeatedly
		chromedp.Evaluate(`
			function loadMoreMovies() {
					console.log("loadMoreMovies function invoked. pagination:", pagination)
					var loadMoreButton = document.querySelector('.ipc-see-more__button');
					if (loadMoreButton) {
							loadMoreButton.click();
							return true;
					}
					return false;
			}
			
			// Attempt to load more movies multiple times
			let pagination = 0;
			function autoLoadMore() {
					console.log("autoLoadMore function invoked. pagination:", pagination)
					if (pagination <= 5 && loadMoreMovies()) {
							pagination++;
							setTimeout(autoLoadMore, 2000); 	// Wait for new content to load
					} else {
					 	console.log("Done loading. Calling doneLoading()")
						document.body.classList.add('done-loading-result');
						console.log("Done calling doneLoading()")
					}
			}
			autoLoadMore();
		`, nil),

		chromedp.WaitVisible(`.done-loading-result`),

		// Capture the entire page HTML
		chromedp.OuterHTML(`html`, &htmlContent),

		// Wait for potential additional content
		chromedp.Sleep(1*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to load page: %v", err)
	}
	fmt.Println("Loaded the page")

	// Parse HTML with htmlquery
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}
	fmt.Println("Parsed the HTML:")

	// XPath to extract movie details
	movieNodes := htmlquery.Find(doc, `//*[@id="__next"]/main/div[2]/div[3]/section/section/div/section/section/div[2]/div/section/div[2]/div[2]/ul/li`)
	fmt.Println("No. of movies len:", len(movieNodes))

	for _, movie := range movieNodes {
		title, year, rating := "", "", ""

		// Extract movie details using XPath
		titleNode, err := htmlquery.Query(movie, `/div/div/div/div[1]/div[2]/div[1]/a/h3`)
		if err != nil {
			fmt.Println("Failed to fetch Title")
		}
		if titleNode != nil {
			title = htmlquery.InnerText(titleNode)
		}
		yearNode, err := htmlquery.Query(movie, `/div/div/div/div[1]/div[2]/div[2]/span[1]`)
		if err != nil {
			fmt.Println("Failed to fetch Year")
		}
		if yearNode != nil {
			year = htmlquery.InnerText(yearNode)
		}
		ratingNode, err := htmlquery.Query(movie, `/div/div/div/div[1]/div[2]/span/div/span/span[1]`)
		if err != nil {
			fmt.Println("Failed to fetch Rating")
		}
		if ratingNode != nil {
			rating = htmlquery.InnerText(ratingNode)
		}

		fmt.Printf("Title: %s, Year: %s, Rating: %s\n", title, year, rating)
	}

	fmt.Println("All Scraped!")

	return nil
}
