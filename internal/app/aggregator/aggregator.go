package aggregator

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"rsshub/internal/adapters/db"
	"rsshub/internal/app/rss"
	"rsshub/internal/domain"
)

type Aggregator struct {
	db            *db.DB
	mu            sync.Mutex
	interval      time.Duration
	numWorkers    int
	ticker        *time.Ticker
	jobs          chan domain.Feed
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	workerCancels []context.CancelFunc
	workerDone    chan struct{} // Added to signal worker termination
}

func NewAggregator(db *db.DB, interval time.Duration, numWorkers int) *Aggregator {
	return &Aggregator{
		db:         db,
		interval:   interval,
		numWorkers: numWorkers,
		jobs:       make(chan domain.Feed),
		workerDone: make(chan struct{}), // Initialize the workerDone channel
	}
}

func (a *Aggregator) Start(parentCtx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ticker != nil {
		return fmt.Errorf("already started")
	}
	a.ctx, a.cancel = context.WithCancel(parentCtx)
	a.ticker = time.NewTicker(a.interval)
	a.wg.Add(1)
	go a.fetchLoop()
	a.startWorkers(a.numWorkers)
	return nil
}

func (a *Aggregator) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ticker == nil {
		return fmt.Errorf("not started")
	}
	a.ticker.Stop()
	a.cancel()
	close(a.jobs)
	close(a.workerDone) // Close the workerDone channel
	a.wg.Wait()
	a.ticker = nil
	a.workerCancels = nil
	return nil
}

func (a *Aggregator) Interval() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.interval
}

func (a *Aggregator) Workers() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.numWorkers
}

func (a *Aggregator) SetInterval(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ticker == nil {
		return
	}
	a.interval = d
	a.ticker.Reset(d)
}

func (a *Aggregator) Resize(workers int) error {
	if workers <= 0 {
		return fmt.Errorf("number of workers must be positive")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if workers < a.numWorkers {
		// Reduce workers (stop excess goroutines)
		for i := workers; i < a.numWorkers; i++ {
			select {
			case a.workerDone <- struct{}{}:
			default:
				// If no worker is listening, skip
			}
		}
	} else if workers > a.numWorkers {
		// Add workers
		a.startWorkers(workers - a.numWorkers)
	}
	a.numWorkers = workers
	return nil
}

func (a *Aggregator) fetchLoop() {
	defer a.wg.Done()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-a.ticker.C:
			a.mu.Lock()
			limit := a.numWorkers
			a.mu.Unlock()
			feeds, err := a.db.GetOutdatedFeeds(limit)
			if err != nil {
				log.Printf("[%s] Error fetching outdated feeds: %v\n", time.Now().Format(time.RFC3339), err)
				continue
			}
			for _, feed := range feeds {
				select {
				case a.jobs <- feed:
				case <-a.ctx.Done():
					return
				}
			}
		}
	}
}

func (a *Aggregator) startWorkers(n int) {
	for i := 0; i < n; i++ {
		ctx, cancel := context.WithCancel(a.ctx)
		a.workerCancels = append(a.workerCancels, cancel)
		a.wg.Add(1)
		go a.worker(ctx)
	}
}

func (a *Aggregator) worker(ctx context.Context) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.workerDone: // Check for termination signal
			return
		case feed, ok := <-a.jobs:
			if !ok {
				return
			}
			if err := a.processFeed(feed); err != nil {
				log.Printf("[%s] Error processing feed %s: %v\n", time.Now().Format(time.RFC3339), feed.URL, err)
			}
		}
	}
}

func (a *Aggregator) processFeed(feed domain.Feed) error {
	// First try to parse as RSS using your existing function
	rssFeed, err := rss.FetchAndParse(feed.URL)
	if err == nil && len(rssFeed.Channel.Item) > 0 {
		// Process as RSS feed
		for _, item := range rssFeed.Channel.Item {
			pubDate, err := time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				log.Printf("[%s] Error parsing date %s: %v\n", time.Now().Format(time.RFC3339), item.PubDate, err)
				continue
			}

			article := &domain.Article{
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				PublishedAt: pubDate,
				FeedID:      feed.ID,
			}

			exists, err := a.db.ArticleExists(feed.ID, item.Link)
			if err != nil {
				return fmt.Errorf("error checking article existence: %v", err)
			}
			if exists {
				continue
			}

			if err := a.db.InsertArticle(article); err != nil {
				return fmt.Errorf("error inserting article: %v", err)
			}
		}
	} else {
		// If RSS parsing failed, try Atom format
		atomFeed, err := fetchAndParseAtom(feed.URL)
		if err != nil {
			return fmt.Errorf("error fetching and parsing feed %s as Atom: %v", feed.URL, err)
		}

		for _, entry := range atomFeed.Entries {
			link := getAtomLink(entry.Link)
			description := entry.Summary
			if description == "" {
				description = entry.Content
			}
			pubDate := entry.Published
			if pubDate == "" {
				pubDate = entry.Updated
			}

			pubTime, err := parseAtomDate(pubDate)
			if err != nil {
				log.Printf("[%s] Error parsing date %s: %v\n", time.Now().Format(time.RFC3339), pubDate, err)
				continue
			}

			article := &domain.Article{
				Title:       entry.Title,
				Link:        link,
				Description: description,
				PublishedAt: pubTime,
				FeedID:      feed.ID,
			}

			exists, err := a.db.ArticleExists(feed.ID, link)
			if err != nil {
				return fmt.Errorf("error checking article existence: %v", err)
			}
			if exists {
				continue
			}

			if err := a.db.InsertArticle(article); err != nil {
				return fmt.Errorf("error inserting article: %v", err)
			}
		}
	}

	if err := a.db.UpdateFeedUpdatedAt(feed.ID); err != nil {
		return fmt.Errorf("error updating feed timestamp: %v", err)
	}
	return nil
}

// Add these helper functions at the bottom of aggregator.go file:

// Atom feed structs
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type AtomEntry struct {
	Title     string     `xml:"title"`
	Link      []AtomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
}

func fetchAndParseAtom(url string) (*AtomFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed AtomFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

func getAtomLink(links []AtomLink) string {
	// Try to find the main content link (prefer alternate link)
	for _, link := range links {
		if link.Rel == "alternate" && link.Href != "" {
			return link.Href
		}
	}

	// Fallback to any link with href
	for _, link := range links {
		if link.Href != "" {
			return link.Href
		}
	}

	return ""
}

func parseAtomDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	// Try Atom format (RFC3339) first
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// Fallback to RSS format
	return time.Parse(time.RFC1123, dateStr)
}
