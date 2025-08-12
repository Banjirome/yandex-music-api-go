package search_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Banjirome/yandex-music-go/client"
)

func TestTracksQuery(t *testing.T) {
	// Fake server returning minimal JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "track" {
			t.Fatalf("unexpected type param")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"tracks":{"results":[{"id":"1","title":"Song"}]}}}`))
	}))
	defer ts.Close()

	c := client.New(client.WithBaseURL(ts.URL))
	resp, err := c.Search.Tracks(context.Background(), "Song", 0, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result.Tracks == nil || len(resp.Result.Tracks.Results) != 1 {
		t.Fatalf("decode failed %+v", resp.Result.Tracks)
	}
}
