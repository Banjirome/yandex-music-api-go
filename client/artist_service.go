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
	"strconv"
	"strings"

	"github.com/Banjirome/yandex-music-go/artist"
	"github.com/Banjirome/yandex-music-go/models"
)

type ArtistService struct{ c *Client }

// Get brief-info for artist (GET /artists/{id}/brief-info)
func (s *ArtistService) Get(ctx context.Context, id string) (*models.Response[artist.ArtistBriefInfo], error) {
	p := path.Join("artists", id, "brief-info")
	return doJSON[artist.ArtistBriefInfo](s.c, ctx, http.MethodGet, p, nil, nil)
}

// GetMany artists (POST /artists form artist-Ids=...)
func (s *ArtistService) GetMany(ctx context.Context, ids ...string) (*models.Response[[]artist.Artist], error) {
	form := url.Values{}
	form.Set("artist-Ids", strings.Join(ids, ","))
	req, err := s.c.newRequest(ctx, http.MethodPost, "artists", nil, nil)
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
		return nil, fmt.Errorf("artists: %s", resp.Status)
	}
	var out models.Response[[]artist.Artist]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTracks paginated (GET /artists/{id}/tracks?page=&pageSize=)
func (s *ArtistService) GetTracks(ctx context.Context, id string, page, pageSize int) (*models.Response[artist.TracksPage], error) {
	q := url.Values{}
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	p := path.Join("artists", id, "tracks")
	return doJSON[artist.TracksPage](s.c, ctx, http.MethodGet, p, q, nil)
}

// GetAllTracks loads brief-info first to get total tracks count and then fetches single page with all tracks.
func (s *ArtistService) GetAllTracks(ctx context.Context, id string) (*models.Response[artist.TracksPage], error) {
	brief, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	total := 0
	if brief.Result.Artist != nil && brief.Result.Artist.Counts != nil {
		total = brief.Result.Artist.Counts.Tracks
	}
	if total == 0 { // fallback to default page fetch
		return s.GetTracks(ctx, id, 0, 20)
	}
	return s.GetTracks(ctx, id, 0, total)
}
