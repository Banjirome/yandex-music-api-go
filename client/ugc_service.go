package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/playlist"
	"github.com/Banjirome/yandex-music-go/ugc"
)

// UgcService реализует операции загрузки пользовательских треков.
type UgcService struct{ c *Client }

// GetUploadLink: GET handlers/ugc-upload.jsx
func (s *UgcService) GetUploadLink(ctx context.Context, pl *playlist.Playlist, fileName string) (*ugc.Upload, error) {
	if pl == nil {
		return nil, fmt.Errorf("playlist is nil")
	}
	q := url.Values{}
	q.Set("filename", fileName)
	q.Set("kind", pl.Kind)
	q.Set("visibility", "private")
	q.Set("external-domain", "music.yandex.ru")
	q.Set("ncrnd", strconv.FormatFloat(rand.Float64(), 'f', 16, 64))
	req, err := s.c.newRequest(ctx, http.MethodGet, path.Join("handlers", "ugc-upload.jsx"), q, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("get upload link: %s", resp.Status)
	}
	var out ugc.Upload
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadBytes: POST <post-target> multipart/form-data with field "file".
func (s *UgcService) UploadBytes(ctx context.Context, upload *ugc.Upload, data []byte) (*models.Response[string], error) {
	if upload == nil {
		return nil, fmt.Errorf("upload link nil")
	}
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("file", "upload.bin")
	if err != nil {
		return nil, err
	}
	if _, err = part.Write(data); err != nil {
		return nil, err
	}
	_ = mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, upload.PostTarget, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if s.c.auth.Token != "" {
		req.Header.Set("Authorization", "OAuth "+s.c.auth.Token)
	}
	if s.c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", s.c.cfg.UserAgent)
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("upload ugc: %s %s", resp.Status, string(b))
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Используем math/rand.Float64 для паритета с Random.NextDouble.
