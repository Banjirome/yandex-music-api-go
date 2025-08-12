package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// RawResult представляет собой произвольный JSON результат.
type RawResult map[string]any

// GenericService предоставляет упрощённые методы Get/List для не реализованных пока веток.
type GenericService struct {
	c        *Client
	basePath string
}

// Get выполняет GET /{basePath}/{id}
func (s *GenericService) Get(ctx context.Context, id string) (RawResult, error) {
	resp, err := s.c.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/%s", s.basePath, id), nil, nil)
	if err != nil {
		return nil, err
	}
	r, err := s.c.http.Do(resp)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", r.Status)
	}
	var out RawResult
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// List выполняет GET /{basePath}?...
func (s *GenericService) List(ctx context.Context, query url.Values) ([]RawResult, error) {
	resp, err := s.c.newRequest(ctx, http.MethodGet, s.basePath, query, nil)
	if err != nil {
		return nil, err
	}
	r, err := s.c.http.Do(resp)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: %s", r.Status)
	}
	var out []RawResult
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
