package services

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/yourusername/car-listing-service/config"
	"github.com/yourusername/car-listing-service/models"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	cookieFile = "facebook_cookies.json"
	targetURL  = "https://www.facebook.com/marketplace/manila/cars?minPrice=350000&exact=false"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ScrollState tracks the state of scrolling and end detection
type ScrollState struct {
	currentScroll           int
	consecutiveNoNewItems   int
	consecutiveUnchangedDOM int
	consecutiveScrollNoMove int
	previousDOMCount        int
	previousScrollY         int
	currentDelay            time.Duration
	seenURLs                map[string]bool
	totalItemsFound         int
	totalNewItems           int
	totalDuplicates         int
	startTime               time.Time
}

func ScrapeCars(resultsChan chan<- []models.Car) error {
	// Load configuration
	cfg := config.LoadConfig()
	scraperConfig := cfg.Scraper

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := loadCookies(ctx, cookieFile); err != nil {
		if err := login(ctx); err != nil {
			return err
		}
		if err := saveCookies(ctx, cookieFile); err != nil {
			log.Printf("Warning: Failed to save cookies: %v", err)
		}
	}

	// Initialize scroll state
	state := &ScrollState{
		currentScroll: 0,
		currentDelay:  scraperConfig.InitialDelay,
		seenURLs:      make(map[string]bool, 10000),
		startTime:     time.Now(),
	}

	return chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(5*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for !shouldStopScraping(state, scraperConfig) {
				if err := performScrollCycle(ctx, state, scraperConfig, resultsChan); err != nil {
					return err
				}
			}
			logSummary(state)
			return nil
		}),
	)
}

func performScrollCycle(ctx context.Context, state *ScrollState, config config.ScraperConfig, resultsChan chan<- []models.Car) error {
	prevDOMCount, prevScrollY := captureDOMState(ctx)

	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`window.scrollBy(0, window.innerHeight)`, nil),
		chromedp.Sleep(state.currentDelay),
	); err != nil {
		return err
	}

	state.currentScroll++
	currDOMCount, currScrollY := captureDOMState(ctx)
	updateScrollSignals(state, prevDOMCount, currDOMCount, prevScrollY, currScrollY)

	allListings := extractListings(ctx)
	newListings := filterDuplicates(allListings, state)

	if len(newListings) == 0 {
		state.consecutiveNoNewItems++
	} else {
		state.consecutiveNoNewItems = 0
		state.totalNewItems += len(newListings)
		resultsChan <- newListings
	}

	state.currentDelay = calculateAdaptiveDelay(
		state.currentDelay,
		len(newListings),
		state.totalDuplicates,
		len(allListings),
		config,
	)

	if state.currentScroll%config.ExtractionInterval == 0 {
		logProgress(state, config)
	}

	return nil
}

func captureDOMState(ctx context.Context) (domCount, scrollY int) {
	chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelectorAll("a[href*='/marketplace/item/']").length`, &domCount),
		chromedp.Evaluate(`window.scrollY`, &scrollY),
	)
	return
}

func extractListings(ctx context.Context) []models.Car {
	var listings []models.Car
	chromedp.Run(ctx, chromedp.Evaluate(`
		Array.from(document.querySelectorAll("a[href*='/marketplace/item/']")).map(a => {
			const text = a.innerText.split('\n');
			let price = "";
			let title = "";
			let location = "";
			let mileage = "";

			text.forEach(line => {
				if (line.includes("₱") || line.includes("PHP") || line.includes("$")) {
					price = line;
				} else if (line.toLowerCase().includes("km")) {
					mileage = line;
				} else if (title === "" && line.length > 5) {
					title = line;
				} else if (location === "" && title !== "" && line !== price && line !== mileage) {
					location = line;
				}
			});

			const url = new URL(a.href);
			const cleanLink = url.origin + url.pathname;

			return {
				title: title,
				price: price,
				location: location,
				mileage: mileage,
				link: cleanLink
			};
		})
	`, &listings))

	return listings
}

func shouldStopScraping(state *ScrollState, config config.ScraperConfig) bool {
	if state.currentScroll >= config.MaxScrolls {
		return true
	}

	if time.Since(state.startTime) >= config.MaxDuration {
		return true
	}

	triggeredCount := 0

	if state.consecutiveUnchangedDOM >= config.MaxConsecutiveUnchanged {
		triggeredCount++
	}

	if state.consecutiveNoNewItems >= config.MaxConsecutiveNoNew {
		triggeredCount++
	}

	if state.consecutiveScrollNoMove >= 10 {
		triggeredCount++
	}

	return triggeredCount >= 2
}

func updateScrollSignals(state *ScrollState, prevDOMCount, currDOMCount, prevScrollY, currScrollY int) {
	if currDOMCount == prevDOMCount && prevDOMCount > 0 {
		state.consecutiveUnchangedDOM++
	} else {
		state.consecutiveUnchangedDOM = 0
	}
	state.previousDOMCount = currDOMCount

	if currScrollY == prevScrollY && prevScrollY > 0 {
		state.consecutiveScrollNoMove++
	} else {
		state.consecutiveScrollNoMove = 0
	}
	state.previousScrollY = currScrollY
}

func filterDuplicates(allListings []models.Car, state *ScrollState) []models.Car {
	var newListings []models.Car
	duplicateCount := 0

	for _, listing := range allListings {
		if listing.Link == "" {
			continue
		}

		if state.seenURLs[listing.Link] {
			duplicateCount++
		} else {
			state.seenURLs[listing.Link] = true
			newListings = append(newListings, listing)
			state.totalItemsFound++
		}
	}

	state.totalDuplicates = duplicateCount
	return newListings
}

func calculateAdaptiveDelay(
	currentDelay time.Duration,
	newItemsCount int,
	duplicateCount int,
	totalInBatch int,
	config config.ScraperConfig,
) time.Duration {
	if newItemsCount > 20 {
		newDelay := currentDelay - 100*time.Millisecond
		if newDelay < config.MinDelay {
			return config.MinDelay
		}
		return newDelay
	}

	if newItemsCount == 0 {
		newDelay := currentDelay + 500*time.Millisecond
		if newDelay > config.MaxDelay {
			return config.MaxDelay
		}
		return newDelay
	}

	if totalInBatch > 0 && float64(duplicateCount)/float64(totalInBatch) > 0.5 {
		newDelay := currentDelay + 200*time.Millisecond
		if newDelay > config.MaxDelay {
			return config.MaxDelay
		}
		return newDelay
	}

	return currentDelay
}

func logProgress(state *ScrollState, config config.ScraperConfig) {
	elapsed := time.Since(state.startTime)
	log.Printf("Progress: Scroll %d/%d | Total items: %d | New in batch: %d | Duplicates: %d | Delay: %v | Elapsed: %v",
		state.currentScroll,
		config.MaxScrolls,
		state.totalItemsFound,
		state.totalNewItems,
		state.totalDuplicates,
		state.currentDelay,
		elapsed.Round(time.Second),
	)
	state.totalNewItems = 0
}

func logSummary(state *ScrollState) {
	elapsed := time.Since(state.startTime)
	log.Printf("Scraping Complete!")
	log.Printf("Total Scrolls: %d", state.currentScroll)
	log.Printf("Total Items Found: %d", state.totalItemsFound)
	log.Printf("Total Duration: %v", elapsed.Round(time.Second))
	if elapsed.Minutes() > 0 {
		log.Printf("Average Items/Minute: %.1f", float64(state.totalItemsFound)/(elapsed.Minutes()))
	}
}

func login(ctx context.Context) error {
	email := getEnv("FACEBOOK_EMAIL", "")
	password := getEnv("FACEBOOK_PASSWORD", "")

	if email == "" || password == "" {
		log.Fatal("FACEBOOK_EMAIL and FACEBOOK_PASSWORD must be set in environment variables")
	}

	return chromedp.Run(ctx,
		chromedp.Navigate("https://www.facebook.com/login"),
		chromedp.WaitVisible("input#email", chromedp.ByQuery),
		chromedp.SendKeys("input#email", email, chromedp.ByQuery),
		chromedp.SendKeys("input#pass", password, chromedp.ByQuery),
		chromedp.Click("button#loginbutton", chromedp.ByQuery),
		chromedp.WaitVisible("div[role='banner']", chromedp.ByQuery),
	)
}

func saveCookies(ctx context.Context, filename string) error {
	var cookies []*network.Cookie
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func loadCookies(ctx context.Context, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var cookies []*network.CookieParam
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	return chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetCookies(cookies).Do(ctx)
		}),
	)
}
