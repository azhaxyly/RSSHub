0) Каркас и базовые гарантии ✅

 go.mod (Go 1.22), базовые пакеты, пустые файлы собираются.

 Запускается go build -o rsshub . и go run -race cmd/rsshub/main.go --help.

 Форматирование gofumpt — подключи pre-commit или просто проверяй перед коммитом.


1) Конфиг и утилиты

 Парсинг ENV: CLI_APP_TIMER_INTERVAL, CLI_APP_WORKERS_COUNT, POSTGRES_*.

 Реальный Clock и Logger (stdlog).

 Единый Config объект с дефолтами (3m, 3 воркера).


2) Миграции и PostgreSQL

 go:embed миграций, раннер миграций (schema_migrations).

 Схема:

 feeds: id UUID PK, name UNIQUE, url, created_at, updated_at, last_fetched_at NULL, etag, last_modified, fetch_fail_count INT DEFAULT 0.

 articles: id UUID PK, feed_id FK, title, link, description, published_at, created_at, updated_at, UNIQUE(feed_id, link).

 Индексы: feeds(last_fetched_at NULLS FIRST, fetch_fail_count DESC), articles(feed_id, published_at DESC).

 Подключение к БД (pgx или database/sql+pgx).

 Репозитории: FeedRepo, ArticleRepo (вставка, выборки, upsert).


3) Порты/интерфейсы приложения (Clean Architecture)

 Определить порты:

 FeedRepo, ArticleRepo

 RSSGateway (Fetch + Parse)

 ControlBus (Listen/Notify)

 Clock, Logger

 DTO/модели use case’ов (вход/выход).


4) RSS слой

 HTTP-клиент с Timeout, User-Agent.

 Conditional requests: отправка If-None-Match/If-Modified-Since, разбор ETag/Last-Modified.

 Rate-limit per host (минимальный sleep, Mutex-карта).

 XML-парсер (encoding/xml) → RSSFeed, RSSItem, нормализация pubDate (RFC1123/variants).


5) Агрегатор (сердце)

 Интерфейс Aggregator (Start/Stop/SetInterval/Resize).

 Тикер (по умолчанию 3m) — безопасная смена интервала (stop+new, без Reset остановленного).

 Пул воркеров:

 Канал jobs (буфер).

 Динамическая смена размера пула (Dispatcher + Mutex).

 Worker: загрузка фида → 304? MarkFetched(time) : 200 → Parse → UpsertMany → MarkFetched(etag,last_modified)/IncFail.

 Выбор «устаревших» фидов: PickOutdated(N) по last_fetched_at NULLS FIRST, last_fetched_at ASC, fetch_fail_count DESC.

 Backoff по fetch_fail_count.

 Graceful shutdown: контекст, wg.Wait(), корректное закрытие каналов.


6) Единичность процесса fetch

 Advisory lock в Postgres: pg_try_advisory_lock(key) / pg_advisory_unlock.

 Если занят — печатаем Background process is already running и выходим.


7) Живые команды (IPC)

 LISTEN/NOTIFY rsshub_control (JSON payload).

 В fetch процессе: Listen и применять SetInterval/Resize при событиях.

 В CLI командах set-interval/set-workers: Notify с нужным payload.

 Логи фраз по ТЗ при применении изменений.


8) CLI (без внешних библиотек)

 Роутинг подкоманд через os.Args + flag.FlagSet.

 Команды:

 fetch — старт агрегатора, блок до Ctrl+C, graceful stop, миграции перед стартом.

 add --name --url — валидация URL, RepoFeeds.Create.

 list [--num N] — последние добавленные (по created_at DESC), форматированный вывод.

 delete --name — удалить по имени.

 articles --feed-name --num — последние статьи.

 set-interval <duration> — Notify в бегущий процесс.

 set-workers <count> — Notify в бегущий процесс.

 --help — usage, краткие описания.


9) Docker Compose

 Сервис postgres (порт 5432), env из .env, volume.

 Сервис rsshub (сборка бинаря, команда по умолчанию ./rsshub --help или fetch).

 .env.example с дефолтами.


10) Валидации/ошибки/устойчивость

 Все ошибки старта → понятное сообщение + non-zero exit.

 Нет горутин-утечек (все слушают ctx.Done()).

 Нет double-close каналов.

 Отсутствие дедлоков (workers постоянно читают jobs).

 -race чисто.

 Не DoSим: таймауты, rate-limit, лимит N фидов за тик.


11) Логи/метрики консолью (на первое время)

 Лог на каждую сетевую попытку (хост, код, длительность).

 Лог по числу вставленных/обновлённых статей.

 Лог по изменению интервала/кол-ва воркеров.


12) Acceptance Criteria (минимально необходимое)

 go build -o rsshub . — без ошибок.

 go run -race cmd/rsshub/main.go — без data race.

 rsshub fetch → сообщение о старте с дефолтами.

 Повторный fetch → «Background process is already running».

 set-interval 2m во втором терминале → сообщение о смене интервала в первом.

 set-workers 5 → сообщение о смене воркеров.

 add/list/delete/articles работают с реальной БД.

 Статьи реально подтягиваются из хотя бы 1–2 тестовых фидов.
 

13) Nice-to-have (после обязательного)

 Web API (GET /feeds, /feeds/{name}/articles).

 Телеграм-бот уведомлений.

 Пер-фидный лимитер и отдельный backoff со «ступенями».

 Кэширование HTTP (ETag/Last-Modified уже есть — расширить до дискового кэша).

 Healthcheck эндпоинт/метрика.