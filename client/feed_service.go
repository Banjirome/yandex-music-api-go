package client

import (
	"context"
	"net/http"

	"github.com/Banjirome/yandex-music-go/feed"
	"github.com/Banjirome/yandex-music-go/models"
)

// FeedService реализует получение персональной ленты /feed.
type FeedService struct{ c *Client }

// Get возвращает ленту (аналог YGetFeedBuilder).
func (s *FeedService) Get(ctx context.Context) (*models.Response[feed.Feed], error) {
	return doJSON[feed.Feed](s.c, ctx, http.MethodGet, "feed", nil, nil)
}
