package client

import (
	"context"
	"net/http"

	"github.com/Banjirome/yandex-music-go/models"
)

type AccountService struct{ c *Client }

type accountStatusResult struct {
	Account struct {
		Uid string `json:"uid"`
	} `json:"account"`
}

// Status вызывает account/status и обновляет auth.User.Uid.
func (s *AccountService) Status(ctx context.Context) (*models.Response[accountStatusResult], error) {
	resp, err := doJSON[accountStatusResult](s.c, ctx, http.MethodGet, "account/status", nil, nil)
	if err == nil && resp != nil && resp.Result.Account.Uid != "" {
		s.c.auth.SetUid(resp.Result.Account.Uid)
	}
	return resp, err
}
