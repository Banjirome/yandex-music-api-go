package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/playlist"
)

// PlaylistService реализует операции над плейлистами с путями users/{uid}/playlists/{kind}.
type PlaylistService struct{ c *Client }

// GetPersonalPlaylists загружает персональные плейлисты через Landing personal-playlists блок.
func (s *PlaylistService) GetPersonalPlaylists(ctx context.Context) ([]*models.Response[playlist.Playlist], error) {
	land, err := s.c.Landing.Get(ctx, "personal-playlists")
	if err != nil {
		return nil, err
	}
	var out []*models.Response[playlist.Playlist]
	for _, b := range land.Result.Blocks {
		if strings.ToLower(b.Type) != "personal-playlists" {
			continue
		}
		for _, ent := range b.Entities {
			// entity expected to contain nested playlist under key "data" -> "data"
			if plMap, ok := ent["data"].(map[string]any); ok {
				if inner, ok2 := plMap["data"].(map[string]any); ok2 {
					// marshal/unmarshal to Playlist
					raw, _ := json.Marshal(inner)
					var pl playlist.Playlist
					if err := json.Unmarshal(raw, &pl); err == nil {
						cpy := pl // avoid pointer alias issues
						out = append(out, &models.Response[playlist.Playlist]{Result: cpy})
					}
				}
			}
		}
	}
	return out, nil
}

func (s *PlaylistService) personalByType(ctx context.Context, typ string) (*models.Response[playlist.Playlist], error) {
	list, err := s.GetPersonalPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range list {
		if r != nil && strings.EqualFold(r.Result.GeneratedPlaylistType, typ) {
			return r, nil
		}
	}
	return nil, errors.New("personal playlist not found")
}

func (s *PlaylistService) OfTheDay(ctx context.Context) (*models.Response[playlist.Playlist], error) {
	return s.personalByType(ctx, playlist.GeneratedPlaylistOfTheDay)
}
func (s *PlaylistService) DejaVu(ctx context.Context) (*models.Response[playlist.Playlist], error) {
	return s.personalByType(ctx, playlist.GeneratedPlaylistNeverHeard)
}
func (s *PlaylistService) Premiere(ctx context.Context) (*models.Response[playlist.Playlist], error) {
	return s.personalByType(ctx, playlist.GeneratedPlaylistRecent)
}
func (s *PlaylistService) Missed(ctx context.Context) (*models.Response[playlist.Playlist], error) {
	return s.personalByType(ctx, playlist.GeneratedPlaylistMissed)
}
func (s *PlaylistService) Kinopoisk(ctx context.Context) (*models.Response[playlist.Playlist], error) {
	return s.personalByType(ctx, playlist.GeneratedPlaylistKinopoisk)
}

// Get получает плейлист по владельцу и kind.
func (s *PlaylistService) Get(ctx context.Context, userUID, kind string) (*models.Response[playlist.Playlist], error) {
	p := path.Join("users", userUID, "playlists", kind)
	return doJSON[playlist.Playlist](s.c, ctx, http.MethodGet, p, nil, nil)
}

// GetByUUID получает плейлист по uuid.
func (s *PlaylistService) GetByUUID(ctx context.Context, uuid string) (*models.Response[playlist.Playlist], error) {
	p := path.Join("playlists", uuid)
	return doJSON[playlist.Playlist](s.c, ctx, http.MethodGet, p, nil, nil)
}

// GetMany получает несколько плейлистов (последовательно).
func (s *PlaylistService) GetMany(ctx context.Context, pairs [][2]string) ([]*models.Response[playlist.Playlist], error) {
	out := make([]*models.Response[playlist.Playlist], 0, len(pairs))
	for _, pr := range pairs {
		r, err := s.Get(ctx, pr[0], pr[1])
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// Create создаёт плейлист.
func (s *PlaylistService) Create(ctx context.Context, userUID, name string) (*models.Response[playlist.Playlist], error) {
	form := url.Values{}
	form.Set("title", name)
	form.Set("visibility", "public")
	p := path.Join("users", userUID, "playlists", "create")
	return s.postFormPlaylist(ctx, p, form)
}

// Rename переименовывает плейлист.
func (s *PlaylistService) Rename(ctx context.Context, userUID, kind, name string) (*models.Response[playlist.Playlist], error) {
	form := url.Values{}
	form.Set("value", name)
	p := path.Join("users", userUID, "playlists", kind, "name")
	return s.postFormPlaylist(ctx, p, form)
}

// Delete удаляет плейлист.
func (s *PlaylistService) Delete(ctx context.Context, userUID, kind string) error {
	// C# удаление: POST users/{uid}/playlists/{kind}/delete
	p := path.Join("users", userUID, "playlists", kind, "delete")
	form := url.Values{}
	resp, err := s.postForm(ctx, p, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("delete playlist: %s", resp.Status)
	}
	return nil
}

// InsertTracks вставляет треки в начало.
func (s *PlaylistService) InsertTracks(ctx context.Context, pl *playlist.Playlist, trackIDs []playlist.TrackKey) (*models.Response[playlist.Playlist], error) {
	change := playlist.ChangeRequest{Operation: "insert", At: 0, Tracks: trackIDs}
	resp, err := s.applyChanges(ctx, pl, []playlist.ChangeRequest{change})
	if err != nil {
		return nil, err
	}
	// обновим локальную ревизию перед повторным чтением (как C# делает повторный Get)
	pl.Revision = resp.Result.Revision
	return s.Get(ctx, pl.Owner.Uid, pl.Kind)
}

// DeleteTracks удаляет указанные треки (определяются по ID, берутся уникально).
func (s *PlaylistService) DeleteTracks(ctx context.Context, pl *playlist.Playlist, tracks []playlist.Track) (*models.Response[playlist.Playlist], error) {
	if pl == nil {
		return nil, fmt.Errorf("playlist nil")
	}
	if len(tracks) == 0 {
		return s.Get(ctx, pl.Owner.Uid, pl.Kind)
	}
	// map id -> struct{} for uniqueness
	uniq := make(map[string]struct{})
	for _, t := range tracks {
		if t.ID != "" {
			uniq[t.ID] = struct{}{}
		}
	}
	changes := make([]playlist.ChangeRequest, 0, len(uniq))
	// Build index map of playlist current tracks
	for idx, cont := range pl.Tracks {
		if cont.Track == nil {
			continue
		}
		if _, ok := uniq[cont.Track.ID]; ok {
			// construct key
			key := playlist.TrackKey{Id: cont.Track.ID}
			if len(cont.Track.Albums) > 0 {
				key.AlbumId = cont.Track.Albums[0].ID
			}
			changes = append(changes, playlist.ChangeRequest{Operation: "delete", From: idx, To: idx + 1, Tracks: []playlist.TrackKey{key}})
		}
	}
	if len(changes) == 0 {
		return s.Get(ctx, pl.Owner.Uid, pl.Kind)
	}
	resp, err := s.applyChanges(ctx, pl, changes)
	if err != nil {
		return nil, err
	}
	pl.Revision = resp.Result.Revision
	return resp, nil
}

// (RemoveTracks / Personal stub удалены для строгого паритета с C# реализацией)

// Favorites соответствует GET users/{uid}/playlists/list
func (s *PlaylistService) Favorites(ctx context.Context, userUID string) (*models.Response[[]playlist.Playlist], error) {
	type wrap = []playlist.Playlist
	p := path.Join("users", userUID, "playlists", "list")
	return doJSON[wrap](s.c, ctx, http.MethodGet, p, nil, nil)
}

// GetBatch соответствует POST playlists/list
func (s *PlaylistService) GetBatch(ctx context.Context, pairs [][2]string) (*models.Response[[]playlist.Playlist], error) {
	ids := make([]string, 0, len(pairs))
	for _, pr := range pairs {
		ids = append(ids, pr[0]+":"+pr[1])
	}
	form := url.Values{}
	form.Set("playlist-ids", strings.Join(ids, ","))
	resp, err := s.postForm(ctx, "playlists/list", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("get playlists: %s", resp.Status)
	}
	data, _ := io.ReadAll(resp.Body)
	var out models.Response[[]playlist.Playlist]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// applyChanges реализует POST users/{uid}/playlists/{kind}/change с diff.
func (s *PlaylistService) applyChanges(ctx context.Context, pl *playlist.Playlist, changes []playlist.ChangeRequest) (*models.Response[playlist.Playlist], error) {
	p := path.Join("users", pl.Owner.Uid, "playlists", pl.Kind, "change")
	diffJSON, _ := json.Marshal(changes)
	form := url.Values{}
	form.Set("kind", pl.Kind)
	form.Set("revision", strconv.Itoa(pl.Revision))
	form.Set("diff", string(diffJSON))
	return s.postFormPlaylist(ctx, p, form)
}

func (s *PlaylistService) postFormPlaylist(ctx context.Context, p string, form url.Values) (*models.Response[playlist.Playlist], error) {
	resp, err := s.postForm(ctx, p, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// Конфликт ревизии: 409 или тело содержит ключевые слова — возвращаем обычную ошибку (без отдельного sentinel).
		lower := strings.ToLower(string(data))
		if resp.StatusCode == http.StatusConflict || (strings.Contains(lower, "revision") && strings.Contains(lower, "conflict")) {
			return nil, fmt.Errorf("playlist revision conflict")
		}
		return nil, fmt.Errorf("playlist op failed: %s", resp.Status)
	}
	var out models.Response[playlist.Playlist]
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PlaylistService) postForm(ctx context.Context, p string, form url.Values) (*http.Response, error) {
	req, err := s.c.newRequest(ctx, http.MethodPost, p, nil, nil)
	if err != nil {
		return nil, err
	}
	body := form.Encode()
	req.Body = io.NopCloser(bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.c.http.Do(req)
}

// join helper (пока не используется)
func joinIDs(user, kind string) string { return strings.Join([]string{user, kind}, ":") }

// SnapshotOptionalQuery пример хелпера
func SnapshotOptionalQuery(snapshot int) string { return strconv.Itoa(snapshot) }

// decodeRaw вспомогательный (пока не используется)
func decodeRaw(data []byte, v any) error { return json.Unmarshal(data, v) }
