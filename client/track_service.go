package client

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Banjirome/yandex-music-go/models"
)

// --- Track related minimal models (subset) ---

type Track struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	DurationMs int64    `json:"durationMs"`
	Albums     []Album  `json:"albums,omitempty"`
	Artists    []Artist `json:"artists,omitempty"`
	Position   *int     `json:"position,omitempty"`
}

type Album struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrackDownloadInfo struct {
	BitrateInKbps   int    `json:"bitrateInKbps"`
	Codec           string `json:"codec"`
	Direct          bool   `json:"direct"`
	DownloadInfoURL string `json:"downloadInfoUrl"`
	Gain            bool   `json:"gain"`
	Preview         bool   `json:"preview"`
}

type StorageDownloadFile struct {
	Host string `json:"host"`
	Path string `json:"path"`
	Ts   string `json:"ts"`
	S    string `json:"s"`
}

type TrackSupplement struct { /* placeholder for extended data */
}

type TrackSimilar struct {
	Tracks []Track `json:"tracks"`
}

// TrackService mirrors C# YTrackAPI.
type TrackService struct{ c *Client }

// Get one or many tracks (POST /tracks form: track-ids, with-positions=true)
func (s *TrackService) Get(ctx context.Context, ids ...string) (*models.Response[[]Track], error) {
	form := url.Values{}
	form.Set("track-ids", strings.Join(ids, ","))
	form.Set("with-positions", "true")
	resp, err := s.postForm(ctx, "tracks", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("get tracks: %s", resp.Status)
	}
	var out models.Response[[]Track]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Metadata for download: GET tracks/{trackKey}/download-info?direct=<bool>
func (s *TrackService) DownloadMetadata(ctx context.Context, trackKey string, direct bool) (*models.Response[[]TrackDownloadInfo], error) {
	q := url.Values{}
	q.Set("direct", fmt.Sprintf("%t", direct))
	p := path.Join("tracks", trackKey, "download-info")
	return doJSON[[]TrackDownloadInfo](s.c, ctx, http.MethodGet, p, q, nil)
}

// DownloadMetadataTrack перегрузка по объекту Track.
func (s *TrackService) DownloadMetadataTrack(ctx context.Context, t Track, direct bool) (*models.Response[[]TrackDownloadInfo], error) {
	key := t.ID
	if len(t.Albums) > 0 {
		key = fmt.Sprintf("%s:%s", t.ID, t.Albums[0].ID)
	}
	return s.DownloadMetadata(ctx, key, direct)
}

// Download file info: GET <downloadInfoUrl> (absolute) returns StorageDownloadFile
func (s *TrackService) DownloadFileInfo(ctx context.Context, downloadInfoURL string) (*StorageDownloadFile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadInfoURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download file info: %s", resp.Status)
	}
	var out StorageDownloadFile
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BuildFileLink reproduces C# BuildLinkForDownload algorithm.
func (s *TrackService) BuildFileLink(meta TrackDownloadInfo, file StorageDownloadFile) (string, error) {
	if file.Path == "" || file.Host == "" || file.Ts == "" || file.S == "" {
		return "", errors.New("incomplete storage file info")
	}
	secret := "XGRlBW9FXlekgbPrRHuSiA" + file.Path[1:] + file.S
	md5sum := md5Hash([]byte(secret))
	sign := sha1Hex(md5sum)
	return fmt.Sprintf("https://%s/get-%s/%s/%s%s", file.Host, meta.Codec, sign, file.Ts, file.Path), nil
}

// FileLink shortcut: choose best mp3 bitrate.
func (s *TrackService) FileLink(ctx context.Context, trackKey string) (string, error) {
	metaResp, err := s.DownloadMetadata(ctx, trackKey, false)
	if err != nil {
		return "", err
	}
	list := metaResp.Result
	if len(list) == 0 {
		return "", errors.New("empty metadata list")
	}
	sort.Slice(list, func(i, j int) bool { return list[i].BitrateInKbps > list[j].BitrateInKbps })
	var chosen *TrackDownloadInfo
	for i := range list {
		if list[i].Codec == "mp3" {
			chosen = &list[i]
			break
		}
	}
	if chosen == nil {
		chosen = &list[0]
	}
	fileInfo, err := s.DownloadFileInfo(ctx, chosen.DownloadInfoURL)
	if err != nil {
		return "", err
	}
	return s.BuildFileLink(*chosen, *fileInfo)
}

// FileLinkTrack перегрузка по объекту Track.
func (s *TrackService) FileLinkTrack(ctx context.Context, t Track) (string, error) {
	key := t.ID
	if len(t.Albums) > 0 {
		key = fmt.Sprintf("%s:%s", t.ID, t.Albums[0].ID)
	}
	return s.FileLink(ctx, key)
}

// Supplement: GET tracks/{id}/supplement
func (s *TrackService) Supplement(ctx context.Context, trackID string) (*models.Response[TrackSupplement], error) {
	p := path.Join("tracks", trackID, "supplement")
	return doJSON[TrackSupplement](s.c, ctx, http.MethodGet, p, nil, nil)
}

// SupplementTrack перегрузка по объекту Track.
func (s *TrackService) SupplementTrack(ctx context.Context, t Track) (*models.Response[TrackSupplement], error) {
	return s.Supplement(ctx, t.ID)
}

// Similar: GET tracks/{id}/similar
func (s *TrackService) Similar(ctx context.Context, trackID string) (*models.Response[TrackSimilar], error) {
	p := path.Join("tracks", trackID, "similar")
	return doJSON[TrackSimilar](s.c, ctx, http.MethodGet, p, nil, nil)
}

// SimilarTrack перегрузка по объекту Track.
func (s *TrackService) SimilarTrack(ctx context.Context, t Track) (*models.Response[TrackSimilar], error) {
	return s.Similar(ctx, t.ID)
}

// SendPlayInfo: POST play-audio form encoded.
func (s *TrackService) SendPlayInfo(ctx context.Context, track Track, from, playID, playlistID string, fromCache bool, totalPlayed, endPos float64) error {
	form := url.Values{}
	form.Set("track_id", track.ID)
	form.Set("from-cache", fmt.Sprintf("%t", fromCache))
	if strings.TrimSpace(playID) == "" {
		playID = fmt.Sprintf("%d-%d-%d", rand.Intn(1000), rand.Intn(1000), rand.Intn(1000))
	}
	form.Set("play_id", playID)
	if s.c.auth == nil || s.c.auth.User == nil || s.c.auth.User.Uid == "" {
		return errors.New("send play info: unauthorized (missing user uid)")
	}
	form.Set("uid", s.c.auth.User.Uid)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	form.Set("timestamp", now)
	form.Set("client-now", now)
	if len(track.Albums) > 0 {
		form.Set("album-id", track.Albums[0].ID)
	}
	form.Set("from", from)
	form.Set("playlist-id", playlistID)
	form.Set("track-length-seconds", fmt.Sprintf("%d", track.DurationMs/1000))
	form.Set("total-played-seconds", fmt.Sprintf("%.3f", totalPlayed))
	form.Set("end-position-seconds", fmt.Sprintf("%.3f", endPos))
	resp, err := s.postForm(ctx, "play-audio", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("play-audio failed: %s %s", resp.Status, string(b))
	}
	return nil
}

// Helpers
func (s *TrackService) postForm(ctx context.Context, p string, form url.Values) (*http.Response, error) {
	req, err := s.c.newRequest(ctx, http.MethodPost, p, nil, nil)
	if err != nil {
		return nil, err
	}
	encoded := form.Encode()
	req.Body = io.NopCloser(bytes.NewBufferString(encoded))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.c.http.Do(req)
}

// ExtractData downloads full binary content for provided trackKey using best FileLink selection.
func (s *TrackService) ExtractData(ctx context.Context, trackKey string) ([]byte, error) {
	link, err := s.FileLink(ctx, trackKey)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("extract data: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// ExtractDataTrack перегрузка по объекту Track.
func (s *TrackService) ExtractDataTrack(ctx context.Context, t Track) ([]byte, error) {
	key := t.ID
	if len(t.Albums) > 0 {
		key = fmt.Sprintf("%s:%s", t.ID, t.Albums[0].ID)
	}
	return s.ExtractData(ctx, key)
}

// ExtractStream returns a ReadCloser for streaming track data; caller must Close.
func (s *TrackService) ExtractStream(ctx context.Context, trackKey string) (io.ReadCloser, error) {
	link, err := s.FileLink(ctx, trackKey)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("extract stream: %s", resp.Status)
	}
	return resp.Body, nil
}

// ExtractStreamTrack перегрузка по объекту Track.
func (s *TrackService) ExtractStreamTrack(ctx context.Context, t Track) (io.ReadCloser, error) {
	key := t.ID
	if len(t.Albums) > 0 {
		key = fmt.Sprintf("%s:%s", t.ID, t.Albums[0].ID)
	}
	return s.ExtractStream(ctx, key)
}

// ExtractToFile downloads track into a file path.
func (s *TrackService) ExtractToFile(ctx context.Context, trackKey, filePath string) error {
	data, err := s.ExtractData(ctx, trackKey)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0o644)
}

// ExtractToFileTrack перегрузка по объекту Track.
func (s *TrackService) ExtractToFileTrack(ctx context.Context, t Track, filePath string) error {
	key := t.ID
	if len(t.Albums) > 0 {
		key = fmt.Sprintf("%s:%s", t.ID, t.Albums[0].ID)
	}
	return s.ExtractToFile(ctx, key, filePath)
}

// crypto helpers
func md5Hash(b []byte) []byte { h := md5.New(); h.Write(b); return h.Sum(nil) }
func sha1Hex(b []byte) string { h := sha1.New(); h.Write(b); return fmt.Sprintf("%x", h.Sum(nil)) }

// end of file
