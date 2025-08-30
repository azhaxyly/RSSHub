package main

import (
	"fmt"
	"log"
	"os"

	"rsshub/internal/adapters/handlers"
	"rsshub/internal/config"
	"rsshub/pkg/db/postgre"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		PrintHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// logger.Init(cfg.LogLevel) Пока не требуется в CLI

	db, err := postgre.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	switch command {
	case "add":
		handlers.AddFeed()
	case "set-interval":
		handlers.SetInterval()
	case "set-workers":
		handlers.SetWorkers()
	case "list":
		handlers.ListFeeds()
	case "delete":
		handlers.DeleteFeed()
	case "articles":
		handlers.ShowArticles()
	case "fetch":
		handlers.StartFetching()
	default:
		fmt.Println("Unknown command:", command)
		PrintHelp()
		os.Exit(1)	
	}
}

func PrintHelp() {
	fmt.Println(`
	./rsshub --help

  Usage:
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