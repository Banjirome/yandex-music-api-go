package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/Banjirome/yandex-music-go/client"
	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/playlist"
)

// We simulate a create-diff attempt with stale revision followed by success after fetching fresh revision.
func TestPlaylistInsertTracksRevisionConflict(t *testing.T) {
	// We exercise InsertTracks -> applyChanges pipeline to trigger revision conflict detection logic.
	// Flow:
	// 1) Get playlist revision=5
	// 2) Insert -> server signals conflict (body contains "revision" & "conflict")
	// 3) Get playlist revision=6
	// 4) Insert again -> success revision=7

	step := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/users/u1/playlists/10":
			var rev int
			// sequence after change to InsertTracks (which performs a GET itself after apply):
			// 0 -> initial GET (5)
			// 1 -> first POST (conflict, no revision change)
			// 2 -> manual GET (6)
			// 3 -> second POST (returns 7)
			// 4 -> trailing GET triggered inside InsertTracks -> should return 7
			switch step {
			case 0:
				rev = 5
			case 2:
				rev = 6
			case 4: // GET triggered inside second InsertTracks after POST success
				rev = 7
			default:
				if step > 4 {
					rev = 7
				} else {
					rev = 6
				}
			}
			resp := models.Response[playlist.Playlist]{Result: playlist.Playlist{Revision: rev, Kind: "10", Owner: &playlist.Owner{Uid: "u1"}}}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPost && r.URL.Path == "/users/u1/playlists/10/change":
			// first apply attempt sees stale revision (after initial GET -> step==1)
			if step == 1 {
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"error":"playlist revision conflict"}`))
				return
			}
			// second succeeds returning new revision 7
			resp := models.Response[playlist.Playlist]{Result: playlist.Playlist{Revision: 7, Kind: "10", Owner: &playlist.Owner{Uid: "u1"}}}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			t.Fatalf("unexpected %s %s at step %d", r.Method, r.URL.Path, step)
		}
		step++
	}))
	defer srv.Close()

	c := sdk.New(sdk.WithBaseURL(srv.URL))
	ctx := context.Background()

	// 1) fetch
	plResp, err := c.Playlist.Get(ctx, "u1", "10")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	pl := plResp.Result

	// 2) first insert triggers conflict
	_, err = c.Playlist.InsertTracks(ctx, &pl, []playlist.TrackKey{{Id: "t1"}})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "revision conflict") {
		t.Fatalf("expected revision conflict error got %v", err)
	}

	// 3) refetch to update revision
	plResp2, err := c.Playlist.Get(ctx, "u1", "10")
	if err != nil {
		t.Fatalf("refetch error: %v", err)
	}
	pl = plResp2.Result
	if pl.Revision != 6 {
		t.Fatalf("expected revision 6 got %d", pl.Revision)
	}

	// 4) retry insert succeeds
	res, err := c.Playlist.InsertTracks(ctx, &pl, []playlist.TrackKey{{Id: "t1"}})
	if err != nil {
		t.Fatalf("retry insert error: %v", err)
	}
	if res.Result.Revision < 6 {
		t.Fatalf("expected revision >=6 got %d", res.Result.Revision)
	}
}

// TestPlaylistDeleteTracks ensures DeleteTracks builds delete diffs for existing track IDs.
func TestPlaylistDeleteTracks(t *testing.T) {
	step := 0
	// initial playlist has two tracks (t1 with album a1, t2 without album)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/users/u1/playlists/55":
			pl := playlist.Playlist{Revision: 10, Kind: "55", Owner: &playlist.Owner{Uid: "u1"}, Tracks: []playlist.TrackContainer{
				{Track: &playlist.Track{ID: "t1", Albums: []playlist.TrackAlbum{{ID: "a1"}}}},
				{Track: &playlist.Track{ID: "t2"}},
			}}
			_ = json.NewEncoder(w).Encode(models.Response[playlist.Playlist]{Result: pl})
		case r.Method == http.MethodPost && r.URL.Path == "/users/u1/playlists/55/change":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form: %v", err)
			}
			if r.PostForm.Get("revision") != "10" {
				t.Fatalf("expected rev 10 got %s", r.PostForm.Get("revision"))
			}
			diff := r.PostForm.Get("diff")
			// Expect two delete operations at positions 0 and 1 (order not strictly guaranteed but likely sequential)
			if !strings.Contains(diff, "\"operation\":\"delete\"") || !strings.Contains(diff, "t1") || !strings.Contains(diff, "t2") {
				t.Fatalf("unexpected diff %s", diff)
			}
			// respond with new revision 11
			_ = json.NewEncoder(w).Encode(models.Response[playlist.Playlist]{Result: playlist.Playlist{Revision: 11, Kind: "55", Owner: &playlist.Owner{Uid: "u1"}}})
		default:
			t.Fatalf("unexpected %s %s step %d", r.Method, r.URL.Path, step)
		}
		step++
	}))
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	ctx := context.Background()
	plResp, err := c.Playlist.Get(ctx, "u1", "55")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	pl := plResp.Result
	// delete both tracks
	del := []playlist.Track{{ID: "t1", Albums: []playlist.TrackAlbum{{ID: "a1"}}}, {ID: "t2"}}
	res, err := c.Playlist.DeleteTracks(ctx, &pl, del)
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if res.Result.Revision != 11 {
		t.Fatalf("expected revision 11 got %d", res.Result.Revision)
	}
}

// TestPlaylistInsertTracksSuccess verifies a simple successful insert path posts correct diff and refetches.
func TestPlaylistInsertTracksSuccess(t *testing.T) {
	step := 0
	var postedDiff string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/users/u2/playlists/77":
			// first GET revision 1, after POST success second GET revision 2
			rev := 1
			if step > 1 {
				rev = 2
			}
			pl := playlist.Playlist{Revision: rev, Kind: "77", Owner: &playlist.Owner{Uid: "u2"}}
			_ = json.NewEncoder(w).Encode(models.Response[playlist.Playlist]{Result: pl})
		case r.Method == http.MethodPost && r.URL.Path == "/users/u2/playlists/77/change":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form: %v", err)
			}
			if r.PostForm.Get("revision") != "1" {
				t.Fatalf("expected revision 1 got %s", r.PostForm.Get("revision"))
			}
			postedDiff = r.PostForm.Get("diff")
			_ = json.NewEncoder(w).Encode(models.Response[playlist.Playlist]{Result: playlist.Playlist{Revision: 2, Kind: "77", Owner: &playlist.Owner{Uid: "u2"}}})
		default:
			t.Fatalf("unexpected %s %s step %d", r.Method, r.URL.Path, step)
		}
		step++
	}))
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	ctx := context.Background()
	plResp, _ := c.Playlist.Get(ctx, "u2", "77")
	pl := plResp.Result
	res, err := c.Playlist.InsertTracks(ctx, &pl, []playlist.TrackKey{{Id: "t1"}})
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}
	if !strings.Contains(postedDiff, "\"operation\":\"insert\"") {
		t.Fatalf("diff missing insert: %s", postedDiff)
	}
	if res.Result.Revision != 2 {
		t.Fatalf("expected revision 2 got %d", res.Result.Revision)
	}
}

// TestPlaylistGetBatch verifies form key playlist-ids and response decode.
func TestPlaylistGetBatch(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/playlists/list" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.PostForm.Get("playlist-ids") != "u1:10,u2:20" {
			t.Fatalf("playlist-ids %s", r.PostForm.Get("playlist-ids"))
		}
		called = true
		pls := []playlist.Playlist{{Kind: "10", Owner: &playlist.Owner{Uid: "u1"}}, {Kind: "20", Owner: &playlist.Owner{Uid: "u2"}}}
		_ = json.NewEncoder(w).Encode(models.Response[[]playlist.Playlist]{Result: pls})
	}))
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	ctx := context.Background()
	pairs := [][2]string{{"u1", "10"}, {"u2", "20"}}
	res, err := c.Playlist.GetBatch(ctx, pairs)
	if err != nil {
		t.Fatalf("get batch error: %v", err)
	}
	if !called {
		t.Fatalf("handler not called")
	}
	if len(res.Result) != 2 {
		t.Fatalf("expected 2 playlists got %d", len(res.Result))
	}
}

// TestPlaylistDeleteTracksNoInput ensures early return path when no tracks provided only triggers GET.
func TestPlaylistDeleteTracksNoInput(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET got %s", r.Method)
		}
		if r.URL.Path != "/users/u3/playlists/5" {
			t.Fatalf("path %s", r.URL.Path)
		}
		calls++
		pl := playlist.Playlist{Revision: 3, Kind: "5", Owner: &playlist.Owner{Uid: "u3"}}
		_ = json.NewEncoder(w).Encode(models.Response[playlist.Playlist]{Result: pl})
	}))
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	ctx := context.Background()
	plResp, _ := c.Playlist.Get(ctx, "u3", "5")
	pl := plResp.Result
	res, err := c.Playlist.DeleteTracks(ctx, &pl, nil)
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if res.Result.Revision != 3 {
		t.Fatalf("unexpected revision %d", res.Result.Revision)
	}
	if calls != 2 {
		t.Fatalf("expected 2 GET calls (initial + early return) got %d", calls)
	}
}
