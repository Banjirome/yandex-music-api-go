package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Banjirome/yandex-music-go/auth"
	"github.com/Banjirome/yandex-music-go/ynison"
	"nhooyr.io/websocket"
)

// YnisonService provides realtime player state via dual websocket (redirector -> state).
type YnisonService struct {
	c *Client
}

// Player represents a connected Ynison player instance.
type Player struct {
	cli *Client
	st  *auth.Storage

	redirectConn *websocket.Conn
	stateConn    *websocket.Conn

	stateMu sync.RWMutex
	state   *ynison.State

	OnReceive func(p *Player, s *ynison.State)
	OnClose   func(p *Player, err error)

	keepAliveCancel context.CancelFunc
}

// Connect establishes websocket connections and starts read loop.
func (s *YnisonService) Connect(ctx context.Context) (*Player, error) {
	if s.c.auth.Token == "" {
		return nil, errors.New("token required")
	}
	p := &Player{cli: s.c, st: s.c.auth}
	if err := p.connect(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Player) connect(ctx context.Context) error {
	// 1. redirector
	rc, _, err := websocket.Dial(ctx, "wss://ynison.music.yandex.ru/redirector.YnisonRedirectService/GetRedirectToYnison", &websocket.DialOptions{
		HTTPHeader: p.wsHeaders(""),
	})
	if err != nil {
		return fmt.Errorf("redirect dial: %w", err)
	}
	p.redirectConn = rc
	// read single redirect message
	_, data, err := rc.Read(ctx)
	if err != nil {
		return fmt.Errorf("redirect read: %w", err)
	}
	var red ynison.Redirect
	if err := json.Unmarshal(data, &red); err != nil {
		return fmt.Errorf("redirect decode: %w", err)
	}
	// 2. state websocket
	stateURL := fmt.Sprintf("wss://%s/ynison_state.YnisonStateService/PutYnisonState", red.Host)
	sc, _, err := websocket.Dial(ctx, stateURL, &websocket.DialOptions{HTTPHeader: p.wsHeaders(red.RedirectTicket)})
	if err != nil {
		return fmt.Errorf("state dial: %w", err)
	}
	p.stateConn = sc
	// send default state bootstrap
	if err := p.sendDefaultState(ctx); err != nil {
		return err
	}
	// start read loop + keepalive if params present
	go p.readLoop()
	if red.KeepAlive != nil && red.KeepAlive.KeepAliveTimeSeconds > 0 {
		ctxKA, cancel := context.WithCancel(context.Background())
		p.keepAliveCancel = cancel
		interval := time.Duration(red.KeepAlive.KeepAliveTimeSeconds) * time.Second
		if interval > 0 {
			go p.keepAliveLoop(ctxKA, interval)
		}
	}
	return nil
}

func (p *Player) wsHeaders(ticket string) http.Header {
	h := http.Header{}
	h.Set("Origin", "https://music.yandex.ru")
	h.Set("Authorization", "OAuth "+p.st.Token)
	protocolMeta := fmt.Sprintf("{\"Ynison-Device-Id\":\"%s\",\"Ynison-Device-Info\":{\"app_name\":\"Chrome\",\"type\":1}}", p.st.DeviceID)
	if ticket == "" {
		h.Set("Sec-WebSocket-Protocol", "Bearer, v2, "+protocolMeta)
	} else {
		protocolMetaWithTicket := fmt.Sprintf("{\"Ynison-Device-Id\":\"%s\",\"Ynison-Device-Info\":{\"app_name\":\"Chrome\",\"type\":1},\"Ynison-Redirect-Ticket\":\"%s\"}", p.st.DeviceID, ticket)
		h.Set("Sec-WebSocket-Protocol", "Bearer, v2, "+protocolMetaWithTicket)
	}
	return h
}

func (p *Player) sendDefaultState(ctx context.Context) error {
	ver := &ynison.Version{DeviceID: p.st.DeviceID, Version: "0", TimestampMs: time.Now().UnixMilli()}
	bootstrap := map[string]any{
		"update_full_state": map[string]any{
			"player_state": map[string]any{
				"player_queue": map[string]any{"version": ver},
				"status":       map[string]any{"version": ver},
			},
			"device": map[string]any{
				"capabilities": map[string]any{"can_be_player": true},
				"info": map[string]any{
					"device_id":   ver.DeviceID,
					"app_name":    "Yandex Music API Go",
					"app_version": "0.0.1",
					"type":        "WEB",
					"title":       "yandex-music-go",
				},
				"is_shadow": true,
			},
		},
		"activity_interception_type": ynison.ActivityDoNotInterceptByDefault,
		"player_action_timestamp_ms": time.Now().UnixMilli(),
	}
	b, _ := json.Marshal(bootstrap)
	return p.stateConn.Write(ctx, websocket.MessageText, b)
}

func (p *Player) readLoop() {
	ctx := context.Background()
	for {
		_, data, err := p.stateConn.Read(ctx)
		if err != nil {
			if p.OnClose != nil {
				p.OnClose(p, err)
			}
			return
		}
		// try decode error
		var em ynison.ErrorMessage
		if err := json.Unmarshal(data, &em); err == nil && em.Error != nil {
			if p.OnClose != nil {
				p.OnClose(p, errors.New(em.Error.Message))
			}
			return
		}
		var st ynison.State
		if err := json.Unmarshal(data, &st); err == nil && st.PlayerState != nil {
			p.stateMu.Lock()
			p.state = &st
			p.stateMu.Unlock()
			if p.OnReceive != nil {
				p.OnReceive(p, &st)
			}
		}
	}
}

// keepAliveLoop sends periodic websocket pings to keep the state connection alive.
func (p *Player) keepAliveLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if p.stateConn == nil {
				return
			}
			_ = p.stateConn.Ping(ctx)
		}
	}
}

// State returns last cached state copy.
func (p *Player) State() *ynison.State {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()
	if p.state == nil {
		return nil
	}
	cp := *p.state
	return &cp
}

// Current returns the current playing track (lazy fetch via Track API) or nil if unavailable.
func (p *Player) Current(ctx context.Context) (*Track, error) {
	st := p.State()
	if st == nil || st.PlayerState == nil || st.PlayerState.PlayerQueue == nil {
		return nil, nil
	}
	idx := st.PlayerState.PlayerQueue.CurrentPlayableIndex
	if idx < 0 || idx >= len(st.PlayerState.PlayerQueue.PlayableList) {
		return nil, nil
	}
	playable := st.PlayerState.PlayerQueue.PlayableList[idx]
	if playable.PlayableID == "" {
		return nil, nil
	}
	resp, err := p.cli.Track.Get(ctx, playable.PlayableID)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Result) == 0 {
		return nil, nil
	}
	return &resp.Result[0], nil
}

// Управляющие методы Next/Previous/Play/Pause/SendRawPlayerState удалены для паритета (в оригинале закомментированы).

// Close terminates connections.
func (p *Player) Close(ctx context.Context) error {
	if p.keepAliveCancel != nil {
		p.keepAliveCancel()
	}
	if p.stateConn != nil {
		_ = p.stateConn.Close(websocket.StatusNormalClosure, "")
	}
	if p.redirectConn != nil {
		_ = p.redirectConn.Close(websocket.StatusNormalClosure, "")
	}
	return nil
}
