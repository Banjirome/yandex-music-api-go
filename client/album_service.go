package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Banjirome/yandex-music-go/album"
	"github.com/Banjirome/yandex-music-go/models"
)

type AlbumService struct{ c *Client }

// Get returns album with tracks (GET /albums/{id}/with-tracks)
func (s *AlbumService) Get(ctx context.Context, id string) (*models.Response[album.Album], error) {
	p := path.Join("albums", id, "with-tracks")
	return doJSON[album.Album](s.c, ctx, http.MethodGet, p, nil, nil)
}

// GetMany returns multiple albums (POST /albums form album-ids=...)
func (s *AlbumService) GetMany(ctx context.Context, ids ...string) (*models.Response[[]album.Album], error) {
	form := url.Values{}
	form.Set("album-ids", strings.Join(ids, ","))
	req, err := s.c.newRequest(ctx, http.MethodPost, "albums", nil, nil)
	if err != nil {
		return nil, err
	}
	body := form.Encode()
	req.Body = io.NopCloser(bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("albums: %s", resp.Status)
	}
	var out models.Response[[]album.Album]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
