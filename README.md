# RSSHub

A powerful CLI application for aggregating RSS feeds with background processing and PostgreSQL storage.

## 🎯 Learning Objectives

- Working with XML and RSS formats
- Concurrency and channels in Go
- Worker Pool implementation
- PostgreSQL integration
- Docker Compose deployment

## 📖 Abstract

RSSHub is a **CLI application — an RSS feed aggregator** that:

- Provides a command-line interface (CLI)
- Fetches and parses RSS feeds from various sources
- Stores articles in PostgreSQL database
- Aggregates RSS feeds using a worker pool in the background

This service collects publications from various sources that provide RSS feeds (news sites, blogs, forums). It helps users stay informed in one place without the need to visit each website manually.

Such a tool is useful for journalists, researchers, analysts, and anyone who wants to stay updated on topics of interest without unnecessary noise. This kind of application makes information more accessible and centralized.

## 🏗️ Project Structure

```
RSSHub/
├── cmd/
│   └── main.go                 # CLI entry point
├── internal/
│   ├── adapters/
│   │   ├── db/                 # Database adapter
│   │   └── handlers/           # CLI command handlers
│   ├── app/
│   │   ├── aggregator/         # RSS feed aggregator
│   │   └── rss/               # RSS parsing logic
│   ├── config/                 # Configuration management
│   └── domain/                 # Domain models
├── migrations/                  # Database migrations
├── pkg/
│   ├── logger/                 # Logging utilities
│   └── uuid/                   # UUID generation
├── docker-compose.yml          # Docker services
├── Dockerfile                  # RSSHub container
└── go.mod                      # Go dependencies
```

## 🚀 Features

- **Background RSS Processing**: Automatically fetches feeds at configurable intervals
- **Worker Pool**: Parallel processing of multiple RSS feeds for improved performance
- **Dynamic Configuration**: Change interval and worker count without restarting
- **PostgreSQL Storage**: Robust database backend for feeds and articles
- **Docker Support**: Easy deployment with Docker Compose
- **Graceful Shutdown**: Proper cleanup of resources on termination

## 🛠️ Prerequisites

- Go 1.23.0 or higher
- Docker and Docker Compose
- PostgreSQL (handled by Docker)

## 📦 Installation

1. **Clone the repository**
   ```bash
   git clone git@git.platform.alem.school:azhaxyly/rsshub.git
   cd RSSHub
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

## 🐳 Quick Start with Docker

1. **Start the services**
   ```bash
   docker-compose up -d
   ```

2. **Verify services are running**
   ```bash
   docker-compose ps
   ```

## 📋 Available Commands

### Start Background Fetching
Starts the background process that periodically fetches and processes RSS feeds.

```bash
./rsshub fetch
```

**Output:**
```
The background process for fetching feeds has started (interval = 3 minutes, workers = 3)
```

### Add New RSS Feed
Add a new RSS feed to the database.

```bash
rsshub add --name "tech-crunch" --url "https://techcrunch.com/feed/"
```

### List Available Feeds
Display RSS feeds stored in the database.

```bash
rsshub list --num 5
```

**Output:**
```
# Available RSS Feeds

1. Name: tech-crunch
   URL: https://techcrunch.com/feed/
   Added: 2025-01-20 15:34

2. Name: hacker-news
   URL: https://news.ycombinator.com/rss
   Added: 2025-01-20 15:37
```

### Show Latest Articles
Display recent articles from a specific feed.

```bash
rsshub articles --feed-name "tech-crunch" --num 5
```

**Output:**
```
Feed: tech-crunch

1. [2025-01-20] Apple announces new M4 chips for MacBook Pro
   https://techcrunch.com/apple-announces-m4/

2. [2025-01-19] OpenAI launches GPT-5 with multimodal capabilities
   https://techcrunch.com/openai-launches-gpt-5/
```

### Dynamic Configuration

#### Change Fetch Interval
```bash
./rsshub set-interval 2m
```
**Output:** `Interval of fetching feeds changed from 3 minutes to 2 minutes`

#### Resize Worker Pool
```bash
./rsshub set-workers 5
```
**Output:** `Number of workers changed from 3 to 5`

### Delete RSS Feed
Remove a feed from the database.

```bash
./rsshub delete --name "tech-crunch"
```

### Show Help
Display usage instructions.

```bash
./rsshub --help
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLI_APP_TIMER_INTERVAL` | RSS fetch interval | `3m` |
| `CLI_APP_WORKERS_COUNT` | Number of worker goroutines | `3` |
| `POSTGRES_HOST` | PostgreSQL host | `postgres` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database username | `postgres` |
| `POSTGRES_PASSWORD` | Database password | `changem` |
| `POSTGRES_DBNAME` | Database name | `rsshub` |

### Default Settings

- **Default interval**: 3 minutes
- **Default workers**: 3
- **Database**: PostgreSQL with automatic migrations

## 🗄️ Database Schema

### Feeds Table
Stores metadata about each RSS feed.

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID (PK) | Unique identifier |
| `created_at` | TIMESTAMP | When feed was added |
| `updated_at` | TIMESTAMP | Last update time |
| `name` | TEXT (unique) | Human-readable name |
| `url` | TEXT | RSS feed URL |

### Articles Table
Stores parsed articles from RSS feeds.

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID (PK) | Unique identifier |
| `created_at` | TIMESTAMP | When stored |
| `updated_at` | TIMESTAMP | Last modified |
| `title` | TEXT | Article title |
| `link` | TEXT | Article URL |
| `published_at` | TIMESTAMP | Original publication time |
| `description` | TEXT | Article summary |
| `feed_id` | UUID (FK) | Reference to feeds.id |

## 🔄 Workflow Example

### Terminal 1: Start Aggregator
```bash
./rsshub fetch
# The background process for fetching feeds has started (interval = 3 minutes, workers = 3)
```

### Terminal 2: Manage Feeds
```bash
# Add a new feed
./rsshub add --name "tech-crunch" --url "https://techcrunch.com/feed/"

# Change settings dynamically
./rsshub set-interval 2m
./rsshub set-workers 5

# View results
./rsshub articles --feed-name "tech-crunch" --num 3
```

### Graceful Shutdown
Press `Ctrl+C` in the aggregator terminal:
```
Graceful shutdown: aggregator stopped
```

## 🚨 Important Notes

### Rate Limiting
- **Do NOT DoS servers** you're fetching feeds from
- Monitor console output for request patterns
- Be ready to stop with `Ctrl+C` if issues arise

### Concurrency Safety
- Uses `sync.Mutex` and `atomic` operations for shared variables
- Proper goroutine management with `context.Context`
- Graceful shutdown prevents resource leaks

## 📚 Sample RSS Feeds

Here are some RSS feeds to get started:

- **TechCrunch**: `https://techcrunch.com/feed/`
- **Hacker News**: `https://news.ycombinator.com/rss`
- **BBC News**: `https://feeds.bbci.co.uk/news/world/rss.xml`
- **Ars Technica**: `http://feeds.arstechnica.com/arstechnica/index`
- **The Verge**: `https://www.theverge.com/rss/index.xml`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you encounter issues:

1. Check the logs: `docker-compose logs rsshub`
2. Verify database connection
3. Ensure all environment variables are set
4. Check that migrations have run successfully


**Happy RSS aggregating! 🚀**
