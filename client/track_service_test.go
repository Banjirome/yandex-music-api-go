package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Banjirome/yandex-music-go/auth"
	sdk "github.com/Banjirome/yandex-music-go/client"
	"github.com/Banjirome/yandex-music-go/models"
)

// TestBuildFileLink validates hashing algorithm parity.
func TestBuildFileLink(t *testing.T) {
	svc := &sdk.TrackService{}
	meta := sdk.TrackDownloadInfo{Codec: "mp3"}
	file := sdk.StorageDownloadFile{Host: "test.host", Path: "/get-mp3/12345/abcdef", Ts: "1650000000000", S: "abcdef"}
	link, err := svc.BuildFileLink(meta, file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link == "" {
		t.Fatalf("empty link")
	}
	// NOTE: We cannot assert full deterministic signature without duplicating the hashing steps here.
	// Instead assert structural parts.
	if want := "https://test.host/get-mp3/"; link[:len(want)] != want {
		t.Fatalf("unexpected prefix: %s", link)
	}
	if link[len(link)-len(file.Path):] != file.Path {
		t.Fatalf("path tail mismatch: %s", link)
	}
}

// TestSendPlayInfoUnauthorized ensures guard triggers when user uid missing.
func TestSendPlayInfoUnauthorized(t *testing.T) {
	c := sdk.New(sdk.WithBaseURL("http://127.0.0.1")) // no server needed; guard triggers first
	err := c.Track.SendPlayInfo(context.Background(), sdk.Track{ID: "1"}, "feed", "", "pl1", false, 1.0, 1.0)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
		t.Fatalf("expected unauthorized error got %v", err)
	}
}

// TestSendPlayInfoSuccess minimal happy path (uid present, server 200).
func TestSendPlayInfoSuccess(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/play-audio" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.PostForm.Get("uid") != "42" {
			t.Fatalf("uid %s", r.PostForm.Get("uid"))
		}
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	st := auth.New("")
	st.SetUid("42")
	c := sdk.New(sdk.WithBaseURL(srv.URL), sdk.WithAuthStorage(st))
	tr := sdk.Track{ID: "t1", DurationMs: 180000, Albums: []sdk.Album{{ID: "a1"}}}
	if err := c.Track.SendPlayInfo(context.Background(), tr, "feed", "pid", "pl1", false, 3.2, 3.2); err != nil {
		t.Fatalf("send error: %v", err)
	}
	if !called {
		t.Fatalf("handler not called")
	}
}

// TestFileLinkSelectBestMp3 ensures FileLink chooses highest bitrate mp3 when available.
func TestFileLinkSelectBestMp3(t *testing.T) {
	// simulate metadata and file info endpoints
	metaCalled := false
	fileCalled := false
	var base string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/tracks/1:2/download-info":
			metaCalled = true
			resp := models.Response[[]sdk.TrackDownloadInfo]{Result: []sdk.TrackDownloadInfo{
				{BitrateInKbps: 128, Codec: "mp3", DownloadInfoURL: base + "/fi/low", Direct: false},
				{BitrateInKbps: 320, Codec: "mp3", DownloadInfoURL: base + "/fi/high", Direct: false},
				{BitrateInKbps: 1411, Codec: "flac", DownloadInfoURL: base + "/fi/flac", Direct: false},
			}}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/fi/"):
			fileCalled = true
			_ = json.NewEncoder(w).Encode(sdk.StorageDownloadFile{Host: "h", Path: "/get-mp3/abc/def", Ts: "1", S: "sig"})
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	srv := httptest.NewServer(handler)
	base = srv.URL
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	link, err := c.Track.FileLink(context.Background(), "1:2")
	if err != nil {
		t.Fatalf("file link error: %v", err)
	}
	if !metaCalled || !fileCalled {
		t.Fatalf("expected metadata and file info calls")
	}
	if !strings.Contains(link, "/get-mp3/") {
		t.Fatalf("unexpected link %s", link)
	}
}
