package client

import (
	"net/http"

	"github.com/Banjirome/yandex-music-go/auth"
)

type Option func(*Config)

type Config struct {
	BaseURL            string
	HTTPClient         *http.Client
	UserAgent          string
	AuthStorage        *auth.Storage
	ClientID           string
	ClientSecret       string
	XClientID          string
	XClientSecret      string
	MobileProxyBaseURL string
}

func defaultConfig() Config {
	return Config{
		BaseURL:   "https://api.music.yandex.net/",
		UserAgent: "yandex-music-go/0.1.0",
	}
}

func WithHTTPClient(h *http.Client) Option { return func(c *Config) { c.HTTPClient = h } }
func WithBaseURL(u string) Option          { return func(c *Config) { c.BaseURL = u } }
func WithUserAgent(ua string) Option       { return func(c *Config) { c.UserAgent = ua } }
func WithToken(token string) Option {
	return func(c *Config) {
		if c.AuthStorage == nil {
			c.AuthStorage = auth.New(token)
		} else {
			c.AuthStorage.Token = token
		}
	}
}
func WithAuthStorage(s *auth.Storage) Option { return func(c *Config) { c.AuthStorage = s } }
func WithClientCredentials(id, secret string) Option {
	return func(c *Config) { c.ClientID, c.ClientSecret = id, secret }
}
func WithXClientCredentials(id, secret string) Option {
	return func(c *Config) { c.XClientID, c.XClientSecret = id, secret }
}
func WithMobileProxyBaseURL(u string) Option { return func(c *Config) { c.MobileProxyBaseURL = u } }
