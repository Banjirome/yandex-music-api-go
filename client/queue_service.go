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

	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/queue"
)

// QueueService mirrors YQueueAPI (list, get, create, update-position).
type QueueService struct{ c *Client }

// List queues: GET queues (optional device header)
func (s *QueueService) List(ctx context.Context, device string) (*models.Response[queue.QueueItemsContainer], error) {
	req, err := s.c.newRequest(ctx, http.MethodGet, "queues", nil, nil)
	if err != nil {
		return nil, err
	}
	if device != "" {
		req.Header.Set("X-Yandex-Music-Device", device)
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("queues list: %s", resp.Status)
	}
	var out models.Response[queue.QueueItemsContainer]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get queue: GET queues/{queueId}
func (s *QueueService) Get(ctx context.Context, queueID string) (*models.Response[queue.Queue], error) {
	p := path.Join("queues", queueID)
	return doJSON[queue.Queue](s.c, ctx, http.MethodGet, p, nil, nil)
}

// Create queue: POST queues (JSON body) optional device header
func (s *QueueService) Create(ctx context.Context, q *queue.Queue, device string) (*models.Response[queue.NewQueue], error) {
	req, err := s.c.newRequest(ctx, http.MethodPost, "queues", nil, q)
	if err != nil {
		return nil, err
	}
	if device != "" {
		req.Header.Set("X-Yandex-Music-Device", device)
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("create queue: %s %s", resp.Status, string(b))
	}
	var out models.Response[queue.NewQueue]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePosition: POST queues/{queueId}/update-position?currentIndex=&isInteractive=
func (s *QueueService) UpdatePosition(ctx context.Context, queueID string, currentIndex int, isInteractive bool, device string) (*models.Response[queue.UpdatedQueue], error) {
	p := path.Join("queues", queueID, "update-position")
	qv := url.Values{}
	qv.Set("currentIndex", fmt.Sprintf("%d", currentIndex))
	qv.Set("isInteractive", strings.ToLower(fmt.Sprintf("%v", isInteractive)))
	req, err := s.c.newRequest(ctx, http.MethodPost, p, qv, nil)
	if err != nil {
		return nil, err
	}
	if device != "" {
		req.Header.Set("X-Yandex-Music-Device", device)
	}
	resp, err := s.c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("update position: %s", resp.Status)
	}
	var out models.Response[queue.UpdatedQueue]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
