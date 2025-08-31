package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"rsshub/internal/adapters/db"
	"rsshub/internal/adapters/handlers"
	"rsshub/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	switch command {
	case "fetch":
		handler.HandleFetch(cfg, database)
	case "add":
		handler.HandleAdd(database)
	case "list":
		handler.HandleList(database)
	case "delete":
		handler.HandleDelete(database)
	case "articles":
		handler.HandleArticles(database)
	case "set-interval":
		handler.HandleSetInterval(cfg)
	case "set-workers":
		handler.HandleSetWorkers(cfg)
	case "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Usage:
  rsshub COMMAND [OPTIONS]

  Common Commands:
     add             add new RSS feed
     set-interval    set RSS fetch interval
     set-workers     set number of workers
     list            list available RSS feeds
     delete          delete RSS feed
     articles        show latest articles
     fetch           starts the background process that periodically fetches and processes RSS feeds using a worker pool`)
}
