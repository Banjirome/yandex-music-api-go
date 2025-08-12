package search

import (
	"context"
	"net/url"
	"strconv"

	"github.com/Banjirome/yandex-music-go/models"
)

// Service реализует методы поиска.
type Service struct{ c internalClient }

// (legacy clientAdapter removed)

type internalClient interface {
	SearchDo(ctx context.Context, typ Type, text string, page, pageSize int) (*models.Response[Search], error)
	SuggestDo(ctx context.Context, part string) (*models.Response[Suggest], error)
}

// Конструктор вызывается из client пакета.
func NewService(c internalClient) *Service { return &Service{c: c} }

// Search универсальный метод.
func (s *Service) Search(ctx context.Context, text string, typ Type, page, pageSize int) (*models.Response[Search], error) {
	return s.c.SearchDo(ctx, typ, text, page, pageSize)
}

// Helper методы.
func (s *Service) Tracks(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypeTrack, page, pageSize)
}
func (s *Service) Albums(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypeAlbum, page, pageSize)
}
func (s *Service) Artists(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypeArtist, page, pageSize)
}
func (s *Service) Playlists(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypePlaylist, page, pageSize)
}
func (s *Service) PodcastEpisodes(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypePodcastEpisode, page, pageSize)
}
func (s *Service) Videos(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypeVideo, page, pageSize)
}
func (s *Service) Users(ctx context.Context, text string, page, pageSize int) (*models.Response[Search], error) {
	return s.Search(ctx, text, TypeUser, page, pageSize)
}

// Suggest возвращает подсказки (GET search/suggest?part=...)
func (s *Service) Suggest(ctx context.Context, part string) (*models.Response[Suggest], error) {
	return s.c.SuggestDo(ctx, part)
}

// Integrating with client: реализация метода SearchDo должна находиться в пакете client.

// BuildQuery утилита (если понадобится внешнему коду в будущем).
func BuildQuery(text string, typ Type, page, pageSize int) url.Values {
	q := url.Values{}
	q.Set("text", text)
	q.Set("type", string(typ))
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	return q
}
