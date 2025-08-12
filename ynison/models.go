package ynison

// Snake_case JSON is used by Ynison service. Tags reflect that.

type Redirect struct {
	Host           string           `json:"host"`
	RedirectTicket string           `json:"redirect_ticket"`
	SessionID      string           `json:"session_id"`
	KeepAlive      *KeepAliveParams `json:"keep_alive_params"`
}

type KeepAliveParams struct {
	KeepAliveTimeSeconds    int `json:"keep_alive_time_seconds"`
	KeepAliveTimeoutSeconds int `json:"keep_alive_timeout_seconds"`
}

type Version struct {
	DeviceID    string `json:"device_id"`
	Version     string `json:"version"`
	TimestampMs int64  `json:"timestamp_ms"`
}

// Activity interception types (string as in server protocol)
const (
	ActivityDoNotInterceptByDefault = "DO_NOT_INTERCEPT_BY_DEFAULT"
)

// Common entity types (enum mapped to upper snake case in C#; we use raw strings)
const (
	EntityTypeRadio = "RADIO"
)

type PlayerState struct {
	PlayerQueue *PlayerQueue       `json:"player_queue"`
	Status      *PlayerStateStatus `json:"status"`
}

type PlayerQueue struct {
	CurrentPlayableIndex int            `json:"current_playable_index"`
	EntityID             string         `json:"entity_id"`
	EntityType           string         `json:"entity_type"`
	EntityContext        string         `json:"entity_context"`
	Options              *QueueOptions  `json:"options"`
	PlayableList         []PlayableItem `json:"playable_list"`
	Queue                *Queue         `json:"queue"`
	FromOptional         string         `json:"from_optional"`
	Version              *Version       `json:"version"`
}

type QueueOptions struct{}

type Queue struct {
	WaveQueue *WaveQueue `json:"wave_queue"`
}

type WaveQueue struct {
	RecommendedPlayableList []PlayableItem `json:"recommended_playable_list"`
	LivePlayableIndex       int            `json:"live_playable_index"`
	EntityOptions           *EntityOptions `json:"entity_options"`
}

type EntityOptions struct {
	TrackSources       []TrackSource       `json:"track_sources"`
	WaveEntityOptional *WaveEntityOptional `json:"wave_entity_optional"`
}

type TrackSource struct {
	Key             int64            `json:"key"`
	WaveSource      *struct{}        `json:"wave_source"`
	PhonotekaSource *PhonotekaSource `json:"phonoteka_source"`
}

type PhonotekaSource struct {
	EntityContext string `json:"entity_context"`
	AlbumID       *ID    `json:"album_id"`
	PlaylistID    *ID    `json:"playlist_id"`
}

type WaveEntityOptional struct {
	SessionID string `json:"session_id"`
}

type ID struct {
	ID string `json:"id"`
}

type PlayableItem struct {
	AlbumIDOptional  string     `json:"album_id_optional"`
	CoverURLOptional string     `json:"cover_url_optional"`
	From             string     `json:"from"`
	PlayableID       string     `json:"playable_id"`
	PlayableType     string     `json:"playable_type"`
	Title            string     `json:"title"`
	TrackInfo        *TrackInfo `json:"track_info"`
}

type TrackInfo struct {
	// placeholder - extend as needed
}

type PlayerStateStatus struct {
	DurationMs    int64    `json:"duration_ms"`
	Paused        bool     `json:"paused"`
	PlaybackSpeed float64  `json:"playback_speed"`
	ProgressMs    int64    `json:"progress_ms"`
	Version       *Version `json:"version"`
}

type State struct {
	Devices     []DeviceFull `json:"devices"`
	PlayerState *PlayerState `json:"player_state"`
	TimestampMs int64        `json:"timestamp_ms"`
}

// UpdatePlayerStateMessage mirrors C# YYnisonUpdatePlayerStateMessage
type UpdatePlayerStateMessage struct {
	UpdatePlayerState        *PlayerState `json:"update_player_state"`
	ActivityInterceptionType string       `json:"activity_interception_type"`
	PlayerActionTimestampMs  int64        `json:"player_action_timestamp_ms"`
}

// UpdateFullStateMessage used for initial bootstrap (not fully typed here)
type UpdateFullStateMessage struct {
	UpdateFullState          any    `json:"update_full_state"`
	ActivityInterceptionType string `json:"activity_interception_type,omitempty"`
	PlayerActionTimestampMs  int64  `json:"player_action_timestamp_ms,omitempty"`
}

type DeviceFull struct {
	Device
	Session   *Session `json:"session"`
	Volume    float64  `json:"volume"`
	IsOffline bool     `json:"is_offline"`
}

type Device struct {
	Info         *DeviceInfo         `json:"info"`
	Capabilities *DeviceCapabilities `json:"capabilities"`
	VolumeInfo   *DeviceVolumeInfo   `json:"volume_info"`
	IsShadow     bool                `json:"is_shadow"`
}

type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
}

type DeviceCapabilities struct {
	CanBePlayer           bool    `json:"can_be_player"`
	CanBeRemoteController bool    `json:"can_be_remote_controller"`
	VolumeGranularity     float64 `json:"volume_granularity"`
}

type DeviceVolumeInfo struct {
	Volume  float64  `json:"volume"`
	Version *Version `json:"version"`
}

type Session struct {
	ID string `json:"id"`
}

type ErrorMessage struct {
	Error *Error `json:"error"`
}

type Error struct {
	Details    *ErrorDetails `json:"details"`
	GrpcCode   int           `json:"grpc_code"`
	HttpCode   int           `json:"http_code"`
	HttpStatus string        `json:"http_status"`
	Message    string        `json:"message"`
}

type ErrorDetails struct {
	YnisonErrorCode     string `json:"ynison_error_code"`
	YnisonBackoffMillis string `json:"ynison_backoff_millis"`
}
