package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"rsshub/internal/config"
	models "rsshub/internal/domain"
	"rsshub/pkg/uuid"
)

type DB struct {
	*sql.DB
}

func NewDB(cfg *config.Config) (*DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PGUser, cfg.PGPassword, cfg.PGHost, cfg.PGPort, cfg.PGDBName,
		cfg.PGSSLmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (d *DB) AddFeed(feed *models.Feed) error {
	if feed.ID == "" {
		id, err := uuid.New()
		if err != nil {
			return err
		}
		feed.ID = id
	}
	_, err := d.Exec(`INSERT INTO feeds (id, name, url) VALUES ($1, $2, $3)`, feed.ID, feed.Name, feed.URL)
	return err
}

func (d *DB) ListFeeds(limit int) ([]models.Feed, error) {
	query := `SELECT id, created_at, updated_at, name, url FROM feeds ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var f models.Feed
		var updated sql.NullTime
		err := rows.Scan(&f.ID, &f.CreatedAt, &updated, &f.Name, &f.URL)
		if err != nil {
			return nil, err
		}
		if updated.Valid {
			f.UpdatedAt = updated.Time
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (d *DB) DeleteFeed(name string) error {
	_, err := d.Exec(`DELETE FROM feeds WHERE name = $1`, name)
	return err
}

func (d *DB) GetArticles(feedName string, limit int) ([]models.Article, error) {
	query := `SELECT a.id, a.created_at, a.updated_at, a.title, a.link, a.published_at, a.description, a.feed_id
      FROM articles a
      JOIN feeds f ON a.feed_id = f.id
      WHERE f.name = $1
      ORDER BY a.published_at DESC
      LIMIT $2`

	rows, err := d.Query(query, feedName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		var updated sql.NullTime
		err := rows.Scan(&a.ID, &a.CreatedAt, &updated, &a.Title, &a.Link, &a.PublishedAt, &a.Description, &a.FeedID)
		if err != nil {
			return nil, err
		}
		if updated.Valid {
			a.UpdatedAt = updated.Time
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (d *DB) GetOutdatedFeeds(limit int) ([]models.Feed, error) {
	query := `SELECT id, created_at, updated_at, name, url FROM feeds ORDER BY updated_at ASC NULLS FIRST LIMIT $1`

	rows, err := d.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var f models.Feed
		var updated sql.NullTime
		err := rows.Scan(&f.ID, &f.CreatedAt, &updated, &f.Name, &f.URL)
		if err != nil {
			return nil, err
		}
		if updated.Valid {
			f.UpdatedAt = updated.Time
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (d *DB) ArticleExists(feedID string, link string) (bool, error) {
	var count int
	err := d.QueryRow(`SELECT COUNT(*) FROM articles WHERE feed_id = $1 AND link = $2`, feedID, link).Scan(&count)
	return count > 0, err
}

func (d *DB) InsertArticle(article *models.Article) error {
	if article.ID == "" {
		id, err := uuid.New()
		if err != nil {
			return err
		}
		article.ID = id
	}
	_, err := d.Exec(`INSERT INTO articles (id, title, link, published_at, description, feed_id)
      VALUES ($1, $2, $3, $4, $5, $6)`, article.ID, article.Title, article.Link, article.PublishedAt, article.Description, article.FeedID)
	return err
}

func (d *DB) UpdateFeedUpdatedAt(id string) error {
	_, err := d.Exec(`UPDATE feeds SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	return err
}
