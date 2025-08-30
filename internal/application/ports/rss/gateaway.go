package rss

import (
	"context"

	"rsshub/internal/domain"
)

type RSSGateway interface {
	Fetch(ctx context.Context, url string, etag, lastModified *string) (status int, payload []byte, newETag, newLastModified *string, err error)
	Parse(ctx context.Context, xml []byte) (feedTitle string, items []domain.Article, err error)
}

type ControlBus interface {
    Listen(ctx context.Context, handle func(evt ControlEvent)) error
    Notify(ctx context.Context, evt ControlEvent) error
}
type ControlEvent struct {
    Type  string      // "set_interval" | "set_workers"
    Value interface{} // "2m" | 5
}
