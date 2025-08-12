package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Banjirome/yandex-music-go/artist"
	sdk "github.com/Banjirome/yandex-music-go/client"
	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/track"
)

func TestArtistGetAllTracks(t *testing.T) {
	// Sequence:
	// 1) brief-info
	// 2) tracks page with pageSize = counts.tracks
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/artists/42/brief-info":
			resp := models.Response[artist.ArtistBriefInfo]{Result: artist.ArtistBriefInfo{Artist: &artist.Artist{Counts: &artist.ArtistCounts{Tracks: 37}}}}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodGet && r.URL.Path == "/artists/42/tracks":
			q := r.URL.Query()
			if q.Get("page") != "0" {
				t.Fatalf("expected page 0 got %s", q.Get("page"))
			}
			if q.Get("pageSize") != strconv.Itoa(37) {
				t.Fatalf("expected pageSize 37 got %s", q.Get("pageSize"))
			}
			resp := models.Response[artist.TracksPage]{Result: artist.TracksPage{Tracks: []track.Track{}}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			t.Fatalf("unexpected call %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := sdk.New(sdk.WithBaseURL(srv.URL))
	_, err := c.Artist.GetAllTracks(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetAllTracks error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}
