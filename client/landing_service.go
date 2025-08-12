package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Banjirome/yandex-music-go/landing"
	"github.com/Banjirome/yandex-music-go/models"
)

// LandingService provides access to landing3, feed and children-landing endpoints.
type LandingService struct {
	c *Client
}

// Get fetches personalized landing blocks. Blocks slice must be non-empty.
// Mirrors GET /landing3?blocks=b1,b2,... with block identifiers matching C# enum values.
func (s *LandingService) Get(ctx context.Context, blocks ...landing.BlockType) (*models.Response[landing.Landing], error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("at least one block required")
	}
	bvals := make([]string, len(blocks))
	for i, b := range blocks {
		bvals[i] = string(b)
	}
	q := url.Values{}
	q.Set("blocks", strings.Join(bvals, ","))
	return doJSON[landing.Landing](s.c, ctx, http.MethodGet, "landing3", q, nil)
}

// Feed fetches user feed. GET /feed
func (s *LandingService) Feed(ctx context.Context) (*models.Response[map[string]any], error) {
	return doJSON[map[string]any](s.c, ctx, http.MethodGet, "feed", nil, nil)
}

// Children fetches children landing. GET /children-landing/catalogue
func (s *LandingService) Children(ctx context.Context) (*models.Response[landing.ChildrenLanding], error) {
	return doJSON[landing.ChildrenLanding](s.c, ctx, http.MethodGet, "children-landing/catalogue", nil, nil)
}
