package repo

import (
	"context"
	"time"

	"rsshub/internal/domain"
)

type FeedRepo interface {
	Create(ctx context.Context, name, url string) error
	DeleteByName(ctx context.Context, name string) error
	List(ctx context.Context, limit *int) ([]domain.Feed, error)
	PickOutdated(ctx context.Context, limit int) ([]domain.Feed, error)
	MarkFetched(ctx context.Context, id string, etag, lastModified *string, at time.Time) error
	IncFail(ctx context.Context, id string) error
	ResetFail(ctx context.Context, id string) error
	WithAdvisoryLock(ctx context.Context, key int64, fn func(context.Context) error) error
}

type ArticleRepo interface {
    UpsertMany(ctx context.Context, feedID string, items []domain.Article) (int, error)
    ListByFeedName(ctx context.Context, feedName string, limit int) ([]domain.Article, error)
}

