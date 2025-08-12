package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Banjirome/yandex-music-go/auth"
	"github.com/Banjirome/yandex-music-go/models"
)

// UserService: минимальный набор: валидация токена и получение account/status.
type UserService struct{ c *Client }

// Authorize устанавливает токен и проверяет его валидность через GetUserAuth.
func (s *UserService) Authorize(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}
	s.c.auth.Token = token
	info, err := s.GetUserAuth(ctx)
	if err != nil {
		return err
	}
	if info == nil || info.Result.Account.Uid == "" {
		return fmt.Errorf("unauthorized")
	}
	// persist minimal user info in storage for parity (Uid, Login, IsAuthorized flag)
	s.c.auth.SetUid(info.Result.Account.Uid)
	s.c.auth.SetLogin(info.Result.Account.Login)
	s.c.auth.IsAuthorized = true
	return nil
}

// GetUserAuth запрашивает account/status.
func (s *UserService) GetUserAuth(ctx context.Context) (*models.Response[UserAuthResult], error) {
	return doJSON[UserAuthResult](s.c, ctx, http.MethodGet, "account/status", nil, nil)
}

// UserAuthResult минимальная модель ответа account/status.
type UserAuthResult struct {
	Account struct {
		Uid   string `json:"uid"`
		Login string `json:"login"`
	} `json:"account"`
}

// --- Расширенные методы авторизации (упрощённая адаптация C#) ---

// CreateAuthSession инициирует сессию и получает доступные методы.
func (s *UserService) CreateAuthSession(ctx context.Context, login string) (*models.Response[models.AuthTypes], error) {
	if login == "" {
		return nil, fmt.Errorf("login empty")
	}
	if err := s.ensureCsrf(ctx); err != nil {
		return nil, err
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"login":      login,
	})
	resp, err := s.passportPOST(ctx, "registration-validations/auth/multi_step/start", body)
	if err != nil {
		return nil, err
	}
	var out models.AuthTypes
	if err := decodeJSON(resp, &out); err != nil {
		return nil, err
	}
	if s.c.auth.AuthToken == nil {
		s.c.auth.AuthToken = &auth.AuthToken{}
	}
	s.c.auth.AuthToken.TrackId = out.TrackId
	return &models.Response[models.AuthTypes]{Result: out}, nil
}

// GetAuthQRLink получает ссылку на QR-код.
func (s *UserService) GetAuthQRLink(ctx context.Context) (string, error) {
	if err := s.ensureCsrf(ctx); err != nil {
		return "", err
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"retpath":    "https://passport.yandex.ru/profile",
		"with_code":  "1",
	})
	resp, err := s.passportPOST(ctx, "registration-validations/auth/password/submit", body)
	if err != nil {
		return "", err
	}
	var qr models.AuthQR
	if err := decodeJSON(resp, &qr); err != nil {
		return "", err
	}
	if s.c.auth.AuthToken == nil {
		s.c.auth.AuthToken = &auth.AuthToken{}
	}
	s.c.auth.AuthToken.TrackId = qr.TrackId
	s.c.auth.AuthToken.CsrfToken = qr.CsrfToken
	if strings.ToLower(qr.Status) != "ok" {
		return "", fmt.Errorf("qr status: %s", qr.Status)
	}
	return fmt.Sprintf("https://passport.yandex.ru/auth/magic/code/?track_id=%s", qr.TrackId), nil
}

// AuthorizeByQR опрашивает статус QR до подтверждения или таймаута.
func (s *UserService) AuthorizeByQR(ctx context.Context, pollInterval time.Duration, maxWait time.Duration) (*models.AuthQRStatus, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("qr session not started")
	}
	deadline := time.Now().Add(maxWait)
	for {
		body := urlValues(map[string]string{
			"csrf_token": s.c.auth.AuthToken.CsrfToken,
			"track_id":   s.c.auth.AuthToken.TrackId,
		})
		resp, err := s.passportPOST(ctx, "auth/new/magic/status/", body)
		if err != nil {
			return nil, err
		}
		var st models.AuthQRStatus
		if err := decodeJSON(resp, &st); err != nil {
			return nil, err
		}
		if strings.ToLower(st.Status) == "ok" && st.MagicLinkConfirmed {
			if err := s.loginByCookies(ctx); err != nil {
				return nil, err
			}
			return &st, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("qr timeout: %s", st.Status)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// GetCaptcha получает captcha.
func (s *UserService) GetCaptcha(ctx context.Context) (*models.AuthCaptcha, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("session not started")
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"track_id":   s.c.auth.AuthToken.TrackId,
	})
	resp, err := s.passportPOST(ctx, "registration-validations/textcaptcha", body)
	if err != nil {
		return nil, err
	}
	var cp models.AuthCaptcha
	if err := decodeJSON(resp, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

// AuthorizeByCaptcha отправляет ответ на captcha.
func (s *UserService) AuthorizeByCaptcha(ctx context.Context, answer string) (*models.AuthBase, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("session not started")
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"track_id":   s.c.auth.AuthToken.TrackId,
		"answer":     answer,
	})
	resp, err := s.passportPOST(ctx, "registration-validations/checkHuman", body)
	if err != nil {
		return nil, err
	}
	var base models.AuthBase
	if err := decodeJSON(resp, &base); err != nil {
		return nil, err
	}
	return &base, nil
}

// GetAuthLetter запрашивает отправку письма.
func (s *UserService) GetAuthLetter(ctx context.Context) (*models.AuthLetter, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("session not started")
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"track_id":   s.c.auth.AuthToken.TrackId,
	})
	resp, err := s.passportPOST(ctx, "registration-validations/auth/send_magic_letter", body)
	if err != nil {
		return nil, err
	}
	var lt models.AuthLetter
	if err := decodeJSON(resp, &lt); err != nil {
		return nil, err
	}
	return &lt, nil
}

// AuthorizeByLetter проверяет статус magic link.
func (s *UserService) AuthorizeByLetter(ctx context.Context) (*models.AuthLetterStatus, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("session not started")
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"track_id":   s.c.auth.AuthToken.TrackId,
	})
	resp, err := s.passportPOST(ctx, "auth/letter/status/", body)
	if err != nil {
		return nil, err
	}
	var st models.AuthLetterStatus
	if err := decodeJSON(resp, &st); err != nil {
		return nil, err
	}
	if strings.ToLower(st.Status) == "ok" && st.MagicLinkConfirmed {
		if err := s.loginByCookies(ctx); err != nil {
			return nil, err
		}
		return &st, nil
	}
	return &st, fmt.Errorf("letter status: %s confirmed=%v", st.Status, st.MagicLinkConfirmed)
}

// AuthorizeByAppPassword завершает вход приложенческим паролем.
func (s *UserService) AuthorizeByAppPassword(ctx context.Context, password string) (*models.AuthBase, error) {
	if s.c.auth.AuthToken == nil || s.c.auth.AuthToken.TrackId == "" {
		return nil, errors.New("session not started")
	}
	body := urlValues(map[string]string{
		"csrf_token": s.c.auth.AuthToken.CsrfToken,
		"track_id":   s.c.auth.AuthToken.TrackId,
		"password":   password,
		"retpath":    "https://passport.yandex.ru/am/finish?status=ok&from=Login",
	})
	resp, err := s.passportPOST(ctx, "registration-validations/auth/multi_step/commit_password", body)
	if err != nil {
		return nil, err
	}
	var base models.AuthBase
	if err := decodeJSON(resp, &base); err != nil {
		return nil, err
	}
	if strings.ToLower(base.Status) == "ok" {
		if err := s.loginByCookies(ctx); err != nil {
			return nil, err
		}
	}
	return &base, nil
}

// GetAccessToken обменивает cookies access token на music token (упрощение: предполагаем storage.AccessToken уже установлен внешне).
func (s *UserService) GetAccessToken(ctx context.Context) (*auth.AccessToken, error) {
	if s.c.auth.AccessToken == nil || s.c.auth.AccessToken.AccessToken == "" {
		return nil, errors.New("access token missing")
	}
	form := urlValues(map[string]string{
		"client_id":     s.c.cfg.ClientID,
		"client_secret": s.c.cfg.ClientSecret,
		"grant_type":    "x-token",
		"access_token":  s.c.auth.AccessToken.AccessToken,
	})
	resp, err := s.oauthPOST(ctx, "/1/token", form)
	if err != nil {
		return nil, err
	}
	var acc auth.AccessToken
	if err := decodeJSON(resp, &acc); err != nil {
		return nil, err
	}
	s.c.auth.Token = acc.AccessToken
	return &acc, nil
}

// loginByCookies выполняет обмен cookies -> access token (sessionid) затем устанавливает storage.AccessToken.
func (s *UserService) loginByCookies(ctx context.Context) error {
	// POST mobile proxy endpoint 1/bundle/oauth/token_by_sessionid
	base := s.c.cfg.MobileProxyBaseURL
	if base == "" {
		base = "https://mobileproxy.passport.yandex.net/"
	}
	endpoint := strings.TrimRight(base, "/") + "/1/bundle/oauth/token_by_sessionid"
	form := urlValues(map[string]string{
		"client_id":     s.c.cfg.XClientID,
		"client_secret": s.c.cfg.XClientSecret,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// emulate headers similar to C# builder (Ya-Client-*). We approximate cookie aggregation implicitly via Jar.
	req.Header.Set("Ya-Client-Host", "passport.yandex.ru")
	resp, err := s.c.http.Do(req)
	if err != nil {
		return err
	}
	var acc auth.AccessToken
	if err := decodeJSON(resp, &acc); err != nil {
		return err
	}
	if acc.AccessToken == "" {
		return fmt.Errorf("empty access token")
	}
	s.c.auth.AccessToken = &acc
	s.c.auth.Token = acc.AccessToken
	s.c.auth.IsAuthorized = true
	return nil
}

// GetLoginInfo получает базовую login info.
func (s *UserService) GetLoginInfo(ctx context.Context) (*models.LoginInfo, error) {
	resp, err := s.loginGET(ctx, "info")
	if err != nil {
		return nil, err
	}
	var li models.LoginInfo
	if err := decodeJSON(resp, &li); err != nil {
		return nil, err
	}
	return &li, nil
}

// --- Вспомогательные низкоуровневые ---
var reCsrf = regexp.MustCompile(`"csrf_token" value="([^"]+)"`)

func (s *UserService) ensureCsrf(ctx context.Context) error {
	if s.c.auth.AuthToken != nil && s.c.auth.AuthToken.CsrfToken != "" {
		return nil
	}
	// GET am (auth methods) to extract csrf
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://passport.yandex.ru/am?app_platform=android", nil)
	r, err := s.c.http.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	m := reCsrf.FindSubmatch(b)
	if len(m) < 2 {
		return fmt.Errorf("csrf not found")
	}
	s.c.auth.AuthToken = &auth.AuthToken{CsrfToken: string(m[1])}
	return nil
}

func urlValues(kv map[string]string) string {
	var b strings.Builder
	first := true
	for k, v := range kv {
		if !first {
			b.WriteByte('&')
		} else {
			first = false
		}
		b.WriteString(urlEncode(k))
		b.WriteByte('=')
		b.WriteString(urlEncode(v))
	}
	return b.String()
}

func urlEncode(s string) string { return url.QueryEscape(s) }

func (s *UserService) passportPOST(ctx context.Context, path string, body string) (*http.Response, error) {
	url := "https://passport.yandex.ru/" + path
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.c.http.Do(req)
}

func (s *UserService) oauthPOST(ctx context.Context, path string, body string) (*http.Response, error) {
	url := "https://oauth.yandex.ru" + path
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return s.c.http.Do(req)
}

func (s *UserService) loginGET(ctx context.Context, path string) (*http.Response, error) {
	url := "https://login.yandex.ru/" + path
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	return s.c.http.Do(req)
}

func decodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(v)
}
