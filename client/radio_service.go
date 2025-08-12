package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/radio"
)

type RadioService struct{ c *Client }

// Dashboard: GET rotor/stations/dashboard
func (s *RadioService) Dashboard(ctx context.Context) (*models.Response[radio.StationsDashboard], error) {
	return doJSON[radio.StationsDashboard](s.c, ctx, http.MethodGet, "rotor/stations/dashboard", nil, nil)
}

// List stations: GET rotor/stations/list
func (s *RadioService) List(ctx context.Context) (*models.Response[[]radio.Station], error) {
	return doJSON[[]radio.Station](s.c, ctx, http.MethodGet, "rotor/stations/list", nil, nil)
}

// Station info: GET rotor/station/{type}:{tag}/info
func (s *RadioService) Station(ctx context.Context, typ, tag string) (*models.Response[[]radio.Station], error) {
	p := path.Join("rotor/station", fmt.Sprintf("%s:%s", typ, tag), "info")
	return doJSON[[]radio.Station](s.c, ctx, http.MethodGet, p, nil, nil)
}

// Tracks sequence: GET rotor/station/{type}:{tag}/tracks?settings2=true[&queue=prevTrackId]
func (s *RadioService) Tracks(ctx context.Context, station radio.Station, prevTrackID string) (*models.Response[radio.StationSequence], error) {
	p := path.Join("rotor/station", fmt.Sprintf("%s:%s", station.Station.ID.Type, station.Station.ID.Tag), "tracks")
	q := url.Values{}
	q.Set("settings2", "true")
	if prevTrackID != "" {
		q.Set("queue", prevTrackID)
	}
	return doJSON[radio.StationSequence](s.c, ctx, http.MethodGet, p, q, nil)
}

// Update settings2: POST rotor/station/{type}:{tag}/settings2 (JSON body)
func (s *RadioService) SetSettings2(ctx context.Context, station radio.Station, settings radio.StationSettings2) (*models.Response[string], error) {
	p := path.Join("rotor/station", fmt.Sprintf("%s:%s", station.Station.ID.Type, station.Station.ID.Tag), "settings2")
	return doJSON[string](s.c, ctx, http.MethodPost, p, nil, settings)
}

// Feedback: POST rotor/station/{type}:{tag}/feedback?batch-id=... JSON body {type, timestamp, from, trackId?, totalPlayedSeconds?}
func (s *RadioService) Feedback(ctx context.Context, station radio.Station, fbType radio.StationFeedbackType, trackID, batchID string, totalPlayedSeconds float64) error {
	p := path.Join("rotor/station", fmt.Sprintf("%s:%s", station.Station.ID.Type, station.Station.ID.Tag), "feedback")
	q := url.Values{}
	if batchID != "" {
		q.Set("batch-id", batchID)
	}
	body := radio.NewFeedbackPayload(fbType, station.Station.IDForFrom, trackID, totalPlayedSeconds, time.Now().Unix())
	req, err := s.c.newRequest(ctx, http.MethodPost, p, q, body)
	if err != nil {
		return err
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("feedback failed: %s %s", resp.Status, string(b))
	}
	return nil
}
