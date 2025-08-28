.
├── cmd/rsshub/
│   └── main.go                          # сборка проводов: конфиг, DI, запуск CLI
├── internal/
│   ├── domain/                          # ЧИСТЫЙ домен
│   │   ├── feed.go                      # Entity Feed
│   │   ├── article.go                   # Entity Article
│   │   ├── errors.go                    # доменные ошибки (var ErrDuplicateLink и т.п.)
│   │   └── policies.go                  # дедупликация, нормализация дат
│   ├── application/                     # USE CASES (все через порты)
│   │   ├── ports/
│   │   │   ├── repo_feeds.go            # FeedRepo (порт)
│   │   │   ├── repo_articles.go         # ArticleRepo (порт)
│   │   │   ├── rss_gateway.go           # RSSGateway (порт для HTTP+XML)
│   │   │   ├── clock.go                 # Clock
│   │   │   ├── logger.go                # Logger
│   │   │   └── control_bus.go           # ControlBus (IPC set-interval / set-workers)
│   │   ├── usecase_add_feed.go
│   │   ├── usecase_list_feeds.go
│   │   ├── usecase_delete_feed.go
│   │   ├── usecase_list_articles.go
│   │   ├── usecase_set_interval.go
│   │   ├── usecase_set_workers.go
│   │   ├── usecase_fetch_loop.go        # агрегатор: тикер + воркеры + graceful shutdown
│   │   └── models.go                    # DTO вход/выход юзкейсов
│   ├── adapters/                        # интерфейсный слой
│   │   ├── cli/                         # контроллеры CLI (без внешних deps)
│   │   │   ├── root.go
│   │   │   ├── add.go
│   │   │   ├── list.go
│   │   │   ├── delete.go
│   │   │   ├── articles.go
│   │   │   ├── fetch.go
│   │   │   ├── set_interval.go
│   │   │   └── set_workers.go
│   │   └── presenter/                   # форматирование вывода (CLI ViewModels)
│   │       ├── feeds_view.go
│   │       └── articles_view.go
│   └── infrastructure/                  # адаптеры портов (детали)
│       ├── config/
│       │   └── config.go                # env → Config (defaults)
│       ├── db/
│       │   ├── pg_conn.go               # подключение pgx
│       │   ├── migrations/
│       │   │   ├── 0001_feeds.up.sql
│       │   │   ├── 0001_feeds.down.sql
│       │   │   ├── 0002_articles.up.sql
│       │   │   ├── 0002_articles.down.sql
│       │   │   └── 0003_indexes.up.sql
│       │   ├── migrate_runner.go        # go:embed + schema_migrations
│       │   ├── repo_feeds_pg.go         # implements application.ports.FeedRepo
│       │   ├── repo_articles_pg.go      # implements application.ports.ArticleRepo
│       │   ├── advisory_lock.go         # singleton fetch через pg_try_advisory_lock
│       │   └── notify_bus.go            # implements ControlBus via LISTEN/NOTIFY
│       ├── rss/
│       │   ├── http_client.go           # http.Client, rate-limit, ETag, Last-Modified
│       │   └── xml_parser.go            # encoding/xml → domain.Article
│       ├── time/
│       │   └── clock_real.go            # implements Clock
│       └── log/
│           └── logger_std.go            # implements Logger (stdout)
├── docker-compose.yml
├── .env.example
└── go.mod
