package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Banjirome/yandex-music-go/album"
	sdk "github.com/Banjirome/yandex-music-go/client"
	"github.com/Banjirome/yandex-music-go/models"
)

func TestAlbumGetManyForm(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/albums" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.PostForm.Get("album-ids") != "1,2,3" {
			t.Fatalf("album-ids form incorrect: %v", r.PostForm)
		}
		called = true
		resp := models.Response[[]album.Album]{Result: []album.Album{{ID: "1", Title: "A"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := sdk.New(sdk.WithBaseURL(srv.URL))
	res, err := c.Album.GetMany(context.Background(), "1", "2", "3")
	if err != nil {
		t.Fatalf("GetMany error: %v", err)
	}
	if !called {
		t.Fatal("handler not called")
	}
	if len(res.Result) != 1 || res.Result[0].ID != "1" {
		t.Fatalf("unexpected result %+v", res.Result)
	}
}

func TestAlbumGetPath(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method %s", r.Method)
		}
		if r.URL.Path != "/albums/123/with-tracks" {
			t.Fatalf("path %s", r.URL.Path)
		}
		called = true
		resp := models.Response[album.Album]{Result: album.Album{ID: "123", Title: "T"}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	c := sdk.New(sdk.WithBaseURL(srv.URL))
	res, err := c.Album.Get(context.Background(), "123")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !called {
		t.Fatal("handler not called")
	}
	if res.Result.ID != "123" {
		t.Fatalf("unexpected id %s", res.Result.ID)
	}
}
