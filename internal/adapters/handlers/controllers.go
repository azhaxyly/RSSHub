package handler

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"rsshub/internal/adapters/db"
	"rsshub/internal/app/aggregator"
	"rsshub/internal/config"
	models "rsshub/internal/domain"
)

const sockAddr = "./rsshub.sock"

func HandleFetch(cfg *config.Config, database *db.DB) {
	agg := aggregator.NewAggregator(database, cfg.TimerInterval, cfg.WorkersCount)
	listener, err := net.Listen("unix", "/tmp/rsshub.sock")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	// Start the aggregator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = agg.Start(ctx)
	if err != nil {
		log.Printf("Failed to start aggregator: %v", err)
		return
	}

	fmt.Printf("The background process for fetching feeds has started (interval = %s, workers = %d)\n",
		cfg.TimerInterval, cfg.WorkersCount)

	// Handle control commands via socket
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				return
			}
			go handleControl(conn, agg, cfg)
		}
	}()

	// Wait for interrupt signal to stop gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	err = agg.Stop()
	if err != nil {
		log.Printf("Failed to stop aggregator: %v", err)
	}
	fmt.Println("Graceful shutdown: aggregator stopped")
}

func handleControl(conn net.Conn, agg *aggregator.Aggregator, cfg *config.Config) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	cmd, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	cmd = strings.TrimSpace(cmd)
	parts := strings.Fields(cmd)
	if len(parts) != 2 {
		fmt.Fprint(conn, "Invalid command\n")
		return
	}
	switch parts[0] {
	case "set-interval":
		d, err := time.ParseDuration(parts[1])
		if err != nil {
			fmt.Fprint(conn, "Invalid duration\n")
			return
		}
		// Add limits for interval
		minInterval := 20 * time.Second
		maxInterval := 60 * time.Minute
		if d < minInterval {
			fmt.Fprint(conn, "Interval too short (minimum 1 minute)\n")
			return
		}
		if d > maxInterval {
			fmt.Fprint(conn, "Interval too long (maximum 60 minutes)\n")
			return
		}
		old := cfg.TimerInterval // Adjust based on where interval is stored
		agg.SetInterval(d)
		cfg.TimerInterval = d
		msg := fmt.Sprintf("[%s] Interval of fetching feeds changed from %s to %s\n",
			time.Now().Format(time.RFC3339), old.String(), d.String())
		fmt.Print(msg)        // Log to server stdout
		fmt.Fprint(conn, msg) // Send to client
	case "set-workers":
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 {
			fmt.Fprint(conn, "Invalid number\n")
			return
		}
		// Add limits for workers
		maxWorkers := 5
		if n > maxWorkers {
			fmt.Fprint(conn, "Too many workers (maximum 5)\n")
			return
		}
		oldN := cfg.WorkersCount // Adjust based on where workers count is stored
		err = agg.Resize(n)      // Capture the error
		if err != nil {
			fmt.Fprintf(conn, "Error resizing: %v\n", err)
			return
		}
		cfg.WorkersCount = n
		msg := fmt.Sprintf("[%s] Number of workers changed from %d to %d\n",
			time.Now().Format(time.RFC3339), oldN, n)
		fmt.Print(msg)        // Log to server stdout
		fmt.Fprint(conn, msg) // Send to client
	default:
		fmt.Fprint(conn, "Unknown command\n")
	}
}

func HandleSetInterval(cfg *config.Config) {
	if len(os.Args) < 3 {
		fmt.Printf("[%s] Usage: rsshub set-interval <duration>\n",
			time.Now().Format(time.RFC3339))
		return
	}
	durStr := os.Args[2]
	_, err := time.ParseDuration(durStr) // Validate locally first.
	if err != nil {
		fmt.Printf("[%s] Invalid duration: %v\n",
			time.Now().Format(time.RFC3339), err)
		return
	}
	conn, err := net.Dial("unix", "/tmp/rsshub.sock")
	if err != nil {
		fmt.Printf("[%s] Background process is not running or failed to connect: %v\n",
			time.Now().Format(time.RFC3339), err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "set-interval %s\n", durStr)
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("[%s] Error reading response: %v\n",
			time.Now().Format(time.RFC3339), err)
	} else {
		fmt.Print(response) // Print the server's confirmation message.
	}
}

func HandleSetWorkers(cfg *config.Config) {
	if len(os.Args) < 3 {
		fmt.Printf("[%s] Usage: rsshub set-workers <count>\n",
			time.Now().Format(time.RFC3339))
		return
	}
	countStr := os.Args[2]
	newWorkers, err := strconv.Atoi(countStr)
	if err != nil || newWorkers <= 0 {
		fmt.Printf("[%s] Invalid number of workers\n",
			time.Now().Format(time.RFC3339))
		return
	}
	conn, err := net.Dial("unix", "/tmp/rsshub.sock")
	if err != nil {
		fmt.Printf("[%s] Background process is not running or failed to connect: %v\n",
			time.Now().Format(time.RFC3339), err)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "set-workers %d\n", newWorkers)
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("[%s] Error reading response: %v\n",
			time.Now().Format(time.RFC3339), err)
	} else {
		fmt.Print(response)
	}
}

// Implement other handlers as per your existing code
func HandleAdd(database *db.DB) {
	addSet := flag.NewFlagSet("add", flag.ExitOnError)
	name := addSet.String("name", "", "feed name")
	url := addSet.String("url", "", "feed url")
	addSet.Parse(os.Args[2:])

	if *name == "" || *url == "" {
		fmt.Printf("[%s] Missing name or url\n", time.Now().Format(time.RFC3339))
		return
	}

	feed := &models.Feed{Name: *name, URL: *url}
	err := database.AddFeed(feed)
	if err != nil {
		fmt.Printf("[%s] Error adding feed: %v\n", time.Now().Format(time.RFC3339), err)
	} else {
		fmt.Printf("[%s] Feed added successfully\n", time.Now().Format(time.RFC3339))
	}
}

func HandleList(database *db.DB) {
	listSet := flag.NewFlagSet("list", flag.ExitOnError)
	num := listSet.Int("num", 0, "number of feeds")
	listSet.Parse(os.Args[2:])
	if *num < 0 {
		fmt.Printf("[%s] This number %v cannot be negative \n", time.Now().Format(time.RFC3339), *num)
		os.Exit(1)
	}
	feeds, err := database.ListFeeds(*num)
	if err != nil {
		fmt.Printf("[%s] Error listing feeds: %v\n", time.Now().Format(time.RFC3339), err)
		return
	}
	fmt.Printf("[%s] # Available RSS Feeds\n", time.Now().Format(time.RFC3339))
	for i, f := range feeds {
		fmt.Printf("%d. Name: %s\n   URL: %s\n   Added: %s\n", i+1, f.Name, f.URL, f.CreatedAt.Format("2006-01-02 15:04"))
	}
}

func HandleDelete(database *db.DB) {
	delSet := flag.NewFlagSet("delete", flag.ExitOnError)
	name := delSet.String("name", "", "feed name")
	delSet.Parse(os.Args[2:])

	if *name == "" {
		fmt.Printf("[%s] Missing name\n", time.Now().Format(time.RFC3339))
		return
	}

	err := database.DeleteFeed(*name)
	if err != nil {
		fmt.Printf("[%s] Error deleting feed: %v\n", time.Now().Format(time.RFC3339), err)
	} else {
		fmt.Printf("[%s] Feed deleted successfully\n", time.Now().Format(time.RFC3339))
	}
}

func HandleArticles(database *db.DB) {
	artSet := flag.NewFlagSet("articles", flag.ExitOnError)
	feedName := artSet.String("feed-name", "", "feed name")
	num := artSet.Int("num", 3, "number of articles")
	artSet.Parse(os.Args[2:])

	if *feedName == "" {
		fmt.Printf("[%s] Missing feed-name\n", time.Now().Format(time.RFC3339))
		return
	}

	articles, err := database.GetArticles(*feedName, *num)
	if err != nil {
		fmt.Printf("[%s] Error getting articles: %v\n", time.Now().Format(time.RFC3339), err)
		return
	}
	fmt.Printf("[%s] Feed: %s\n", time.Now().Format(time.RFC3339), *feedName)
	for i, a := range articles {
		fmt.Printf("%d. [%s] %s\n   %s\n", i+1, a.PublishedAt.Format("2006-01-02"), a.Title, a.Link)
	}
}
