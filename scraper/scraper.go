package scraper

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/network"
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
	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create chromedp context
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var htmlContent string
	var err error

	// Navigate to page and extract HTML
	url := fmt.Sprintf("https://www.imdb.com/search/title/?title_type=%s&genres=%s&keywords=%s", c.Title, c.Genre, c.Keyword)
	fmt.Println("Loading URL:", url)

	port := flag.Int("port", 8544, "port")
	flag.Parse()

	// run server
	go headerServer(fmt.Sprintf(":%d", *port))

	var res string
	err = chromedp.Run(ctx,
		setheaders(
			fmt.Sprintf("http://localhost:%d", *port),
			map[string]interface{}{
				"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language": "en-US,en;q=0.5",
				"Referer":         "https://www.google.com/",
			},
			&res,
		),

		chromedp.Navigate(url),

		// chromedp.WaitVisible(`div.lister-list`),

		// Wait for potential additional content
		chromedp.Sleep(5*time.Second),

		// Capture the entire page HTML
		chromedp.OuterHTML(`html`, &htmlContent),

		// Click "Load More" button repeatedly
		chromedp.Evaluate(`
			function loadMoreMovies() {
					var loadMoreButton = document.querySelector('.load-more-button');
					if (loadMoreButton) {
							loadMoreButton.click();
							return true;
					}
					return false;
			}

			// Attempt to load more movies multiple times
			var attempts = 0;
			function autoLoadMore() {
					if (attempts < 3 && loadMoreMovies()) {
							attempts++;
							// Wait for new content to load
							setTimeout(autoLoadMore, 2000);
					}
			}
			autoLoadMore();
		`, nil),
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

// headerServer is a simple HTTP server that displays the passed headers in the html.
func headerServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		buf, err := json.MarshalIndent(req.Header, "", "  ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(res, indexHTML, string(buf))
	})
	return http.ListenAndServe(addr, mux)
}

// setheaders returns a task list that sets the passed headers.
func setheaders(host string, headers map[string]interface{}, res *string) chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(headers)),
		chromedp.Navigate(host),
		chromedp.Text(`#result`, res, chromedp.ByID, chromedp.NodeVisible),
	}
}

const indexHTML = `<!doctype html>
<html>
<body>
  <div id="result">%s</div>
</body>
</html>`
