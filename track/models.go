package track

import (
	"time"
)

// Track mirrors essential YTrack fields for album/artist responses.
type Track struct {
	Albums                         []any     `json:"albums"`
	Artists                        []any     `json:"artists"`
	Available                      bool      `json:"available"`
	AvailableForPremiumUsers       bool      `json:"availableForPremiumUsers"`
	AvailableFullWithoutPermission bool      `json:"availableFullWithoutPermission"`
	AvailableForOptions            []string  `json:"availableForOptions"`
	BackgroundVideoURI             string    `json:"backgroundVideoUri"`
	Best                           bool      `json:"best"`
	Chart                          any       `json:"chart"`
	ContentWarning                 string    `json:"contentWarning"`
	CoverURI                       string    `json:"coverUri"`
	ClipIDs                        []string  `json:"clipIds"`
	DerivedColors                  any       `json:"derivedColors"`
	Disclaimers                    []string  `json:"disclaimers"`
	DurationMs                     int64     `json:"durationMs"`
	Error                          string    `json:"error"`
	Fade                           any       `json:"fade"`
	FileSize                       int64     `json:"fileSize"`
	ID                             string    `json:"id"`
	IsSuitableForChildren          bool      `json:"isSuitableForChildren"`
	Major                          any       `json:"major"`
	Normalization                  any       `json:"normalization"`
	R128                           any       `json:"r128"`
	OgImage                        string    `json:"ogImage"`
	LyricsAvailable                bool      `json:"lyricsAvailable"`
	LyricsInfo                     any       `json:"lyricsInfo"`
	PlayerID                       string    `json:"playerId"`
	PreviewDurationMs              int64     `json:"previewDurationMs"`
	PodcastEpisodeType             string    `json:"podcastEpisodeType"`
	PubDate                        time.Time `json:"pubDate"`
	RealID                         string    `json:"realId"`
	RememberPosition               bool      `json:"rememberPosition"`
	ShortDescription               string    `json:"shortDescription"`
	SpecialAudioResources          []string  `json:"specialAudioResources"`
	StorageDir                     string    `json:"storageDir"`
	Substituted                    *Track    `json:"substituted"`
	Title                          string    `json:"title"`
	TrackSharingFlag               any       `json:"trackSharingFlag"`
	TrackSource                    any       `json:"trackSource"`
	Type                           string    `json:"type"`
	Version                        string    `json:"version"`
}
