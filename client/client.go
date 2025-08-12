package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Banjirome/yandex-music-go/auth"
	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/search"
)

// Client корневой объект SDK.
type Client struct {
	cfg  Config
	http *http.Client
	auth *auth.Storage

	Search   *search.Service
	Album    *AlbumService
	Artist   *ArtistService
	Track    *TrackService
	Account  *AccountService
	Playlist *PlaylistService
	User     *UserService
	Queue    *QueueService
	Radio    *RadioService
	Library  *LibraryService
	Landing  *LandingService
	Feed     *FeedService
	Label    *GenericService
	Ugc      *UgcService
	Ynison   *YnisonService

	Downloader *Downloader
}

// New создаёт новый клиент.
func New(opts ...Option) *Client {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.HTTPClient.Jar == nil {
		if jar, err := cookiejar.New(nil); err == nil {
			cfg.HTTPClient.Jar = jar
		}
	}
	if cfg.AuthStorage == nil {
		cfg.AuthStorage = auth.New("")
	}

	c := &Client{cfg: cfg, http: cfg.HTTPClient, auth: cfg.AuthStorage}

	c.Search = search.NewService(c)
	c.Album = &AlbumService{c: c}
	c.Artist = &ArtistService{c: c}
	c.Track = &TrackService{c: c}
	c.Account = &AccountService{c: c}
	c.Playlist = &PlaylistService{c: c}
	c.User = &UserService{c: c}
	c.Queue = &QueueService{c: c}
	c.Radio = &RadioService{c: c}
	c.Library = &LibraryService{c: c}
	c.Landing = &LandingService{c: c}
	c.Feed = &FeedService{c: c}
	c.Label = &GenericService{c: c, basePath: "labels"}
	c.Ugc = &UgcService{c: c}
	c.Ynison = &YnisonService{c: c}

	c.Downloader = &Downloader{c: c}
	return c
}

// SearchDo реализует internal интерфейс search.Service.
func (c *Client) SearchDo(ctx context.Context, typ search.Type, text string, page, pageSize int) (*models.Response[search.Search], error) {
	q := search.BuildQuery(text, typ, page, pageSize)
	return doJSON[search.Search](c, ctx, http.MethodGet, "search", q, nil)
}

// SuggestDo implements search suggest.
func (c *Client) SuggestDo(ctx context.Context, part string) (*models.Response[search.Suggest], error) {
	q := url.Values{}
	q.Set("part", part)
	return doJSON[search.Suggest](c, ctx, http.MethodGet, "search/suggest", q, nil)
}

// newRequest формирует http.Request.
func (c *Client) newRequest(ctx context.Context, method, p string, q url.Values, body any) (*http.Request, error) {
	base := strings.TrimRight(c.cfg.BaseURL, "/")
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, p)
	if q != nil {
		u.RawQuery = q.Encode()
	}

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = io.NopCloser(strings.NewReader(string(b)))
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), r)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.auth.Token != "" {
		req.Header.Set("Authorization", "OAuth "+c.auth.Token)
	}
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	return req, nil
}

// doJSON выполняет запрос и декодирует JSON в Response[T].
// doJSON выполняет запрос и декодирует JSON.
func doJSON[T any](c *Client, ctx context.Context, method, p string, q url.Values, body any) (*models.Response[T], error) {
	req, err := c.newRequest(ctx, method, p, q, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		apiErr := &models.APIError{StatusCode: resp.StatusCode}
		_ = json.Unmarshal(data, apiErr) // best-effort (Body field)
		return nil, apiErr
	}

	var out models.Response[T]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
