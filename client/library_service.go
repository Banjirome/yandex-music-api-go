package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/Banjirome/yandex-music-go/library"
	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/playlist"
)

// LibraryService implements YLibraryAPI parity (likes/dislikes & recently listened).
type LibraryService struct{ c *Client }

// internal helper building path: users/{uid}/{type}/{section}
func (s *LibraryService) sectionPath(section library.Section, typ library.SectionType) (string, error) {
	if s.c.auth == nil || s.c.auth.User == nil || s.c.auth.User.Uid == "" {
		return "", fmt.Errorf("user uid not set; call Account.Status first")
	}
	return path.Join("users", s.c.auth.User.Uid, string(typ), string(section)), nil
}

// generic GET for sections
func getSection[T any](s *LibraryService, ctx context.Context, section library.Section, typ library.SectionType) (*models.Response[T], error) {
	p, err := s.sectionPath(section, typ)
	if err != nil {
		return nil, err
	}
	return doJSON[T](s.c, ctx, http.MethodGet, p, nil, nil)
}

// Liked entities
func (s *LibraryService) LikedTracks(ctx context.Context) (*models.Response[library.LibraryTracks], error) {
	return getSection[library.LibraryTracks](s, ctx, library.SectionTracks, library.SectionTypeLikes)
}
func (s *LibraryService) LikedAlbums(ctx context.Context) (*models.Response[[]library.LibraryAlbum], error) {
	return getSection[[]library.LibraryAlbum](s, ctx, library.SectionAlbums, library.SectionTypeLikes)
}
func (s *LibraryService) LikedArtists(ctx context.Context) (*models.Response[[]any], error) { // TODO: implement full artist model ref if needed
	return getSection[[]any](s, ctx, library.SectionArtists, library.SectionTypeLikes)
}
func (s *LibraryService) LikedPlaylists(ctx context.Context) (*models.Response[[]library.LibraryPlaylists], error) {
	return getSection[[]library.LibraryPlaylists](s, ctx, library.SectionPlaylists, library.SectionTypeLikes)
}

// Disliked entities
func (s *LibraryService) DislikedTracks(ctx context.Context) (*models.Response[library.LibraryTracks], error) {
	return getSection[library.LibraryTracks](s, ctx, library.SectionTracks, library.SectionTypeDislikes)
}
func (s *LibraryService) DislikedArtists(ctx context.Context) (*models.Response[[]any], error) {
	return getSection[[]any](s, ctx, library.SectionArtists, library.SectionTypeDislikes)
}

// modify helper: POST users/{uid}/{type}/{section}/add-multiple or /remove with form field <singular>-ids
func (s *LibraryService) modifyForm(ctx context.Context, section library.Section, typ library.SectionType, add bool, id string) (*http.Response, error) {
	p, err := s.sectionPath(section, typ)
	if err != nil {
		return nil, err
	}
	action := "remove"
	if add {
		action = "add-multiple"
	}
	p = path.Join(p, action)
	form := url.Values{}
	singular := strings.TrimSuffix(string(section), "s")
	form.Set(singular+"-ids", id)
	req, err := s.c.newRequest(ctx, http.MethodPost, p, nil, nil)
	if err != nil {
		return nil, err
	}
	encoded := form.Encode()
	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.c.http.Do(req)
}

// Likes/dislikes operations
// AddTrackLike returns resulting playlist (C# returns YPlaylist)
func (s *LibraryService) AddTrackLike(ctx context.Context, trackKey string) (*models.Response[playlist.Playlist], error) {
	resp, err := s.modifyForm(ctx, library.SectionTracks, library.SectionTypeLikes, true, trackKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("add track like failed: %s", resp.Status)
	}
	var out models.Response[playlist.Playlist]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) RemoveTrackLike(ctx context.Context, trackKey string) (*models.Response[models.Revision], error) {
	resp, err := s.modifyForm(ctx, library.SectionTracks, library.SectionTypeLikes, false, trackKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remove track like failed: %s", resp.Status)
	}
	var out models.Response[models.Revision]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) AddTrackDislike(ctx context.Context, trackKey string) (*models.Response[models.Revision], error) {
	resp, err := s.modifyForm(ctx, library.SectionTracks, library.SectionTypeDislikes, true, trackKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("add track dislike failed: %s", resp.Status)
	}
	var out models.Response[models.Revision]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) RemoveTrackDislike(ctx context.Context, trackKey string) (*models.Response[models.Revision], error) {
	resp, err := s.modifyForm(ctx, library.SectionTracks, library.SectionTypeDislikes, false, trackKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remove track dislike failed: %s", resp.Status)
	}
	var out models.Response[models.Revision]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) AddAlbumLike(ctx context.Context, albumID string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionAlbums, library.SectionTypeLikes, true, albumID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("add album like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) RemoveAlbumLike(ctx context.Context, albumID string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionAlbums, library.SectionTypeLikes, false, albumID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remove album like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) AddArtistLike(ctx context.Context, artistID string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionArtists, library.SectionTypeLikes, true, artistID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("add artist like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) RemoveArtistLike(ctx context.Context, artistID string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionArtists, library.SectionTypeLikes, false, artistID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remove artist like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) AddPlaylistLike(ctx context.Context, playlistKey string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionPlaylists, library.SectionTypeLikes, true, playlistKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("add playlist like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
func (s *LibraryService) RemovePlaylistLike(ctx context.Context, playlistKey string) (*models.Response[string], error) {
	resp, err := s.modifyForm(ctx, library.SectionPlaylists, library.SectionTypeLikes, false, playlistKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("remove playlist like failed: %s", resp.Status)
	}
	var out models.Response[string]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RecentlyListened mirrors GetRecentlyListenedAsync
func (s *LibraryService) RecentlyListened(ctx context.Context, types []library.PlayContextType, trackCount, contextCount int) (*models.Response[library.RecentlyListenedContext], error) {
	if s.c.auth == nil || s.c.auth.User == nil || s.c.auth.User.Uid == "" {
		return nil, fmt.Errorf("user uid not set")
	}
	p := path.Join("users", s.c.auth.User.Uid, "contexts")
	q := url.Values{}
	q.Set("trackCount", fmt.Sprintf("%d", trackCount))
	q.Set("contextCount", fmt.Sprintf("%d", contextCount))
	parts := make([]string, len(types))
	for i, t := range types {
		parts[i] = string(t)
	}
	q.Set("types", strings.Join(parts, ","))
	return doJSON[library.RecentlyListenedContext](s.c, ctx, http.MethodGet, p, q, nil)
}

// doForm is like doJSON but sends form data.
func doForm[T any](c *Client, ctx context.Context, p string, form url.Values) (*models.Response[T], error) {
	req, err := c.newRequest(ctx, http.MethodPost, p, nil, nil)
	if err != nil {
		return nil, err
	}
	encoded := form.Encode()
	req.Body = io.NopCloser(strings.NewReader(encoded))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("form post failed: %s", resp.Status)
	}
	var out models.Response[T]
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
