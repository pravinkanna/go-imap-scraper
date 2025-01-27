package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
)

// Function to fetch the HTML of the url and return it
func getHtml(url string) (string, error) {
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set User-Agent and other headers to mimic a browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", "https://www.google.com/")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Function to fetch the HTML of list of urls concurrently and returns them in a channel
func getHtmlsConcurrent(urls []string, htmlsCh chan string) (chan string, error) {
	for _, url := range urls {
		go func() {
			time.Sleep(1 * time.Second)
			html, err := getHtml(url)
			if err != nil {
				fmt.Println("Error fetching URL:", url, "\nError:", err)
			}
			htmlsCh <- html
		}()
	}
	return htmlsCh, nil
}

// Function to search for movies using the user input and return the search result page HTML
// This function will handle the pagination
func getMovieSearchResult(ctx context.Context, url string) (string, error) {
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
						if (pagination < 1 && loadMoreMovies()) {
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
		return htmlContent, err
	}
	fmt.Println("Loaded the page")

	return htmlContent, err
}
