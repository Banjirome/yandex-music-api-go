package radio

import (
	"time"

	"github.com/Banjirome/yandex-music-go/track"
)

type StationID struct {
	Tag  string `json:"tag"`
	Type string `json:"type"`
}

type StationDescription struct {
	FullImageURL    string     `json:"fullImageUrl"`
	Icon            any        `json:"icon"`
	GeocellIcon     any        `json:"geocellIcon"`
	ID              StationID  `json:"id"`
	IDForFrom       string     `json:"idForFrom"`
	MtsFullImageURL string     `json:"mtsFullImageUrl"`
	MtsIcon         any        `json:"mtsIcon"`
	Name            string     `json:"name"`
	ParentID        *StationID `json:"parentId"`
	Restrictions    any        `json:"restrictions"`
	Restrictions2   any        `json:"restrictions2"`
}

type Station struct {
	AdParams       any                `json:"adParams"`
	CustomName     string             `json:"customName"`
	Data           any                `json:"data"`
	Explanation    string             `json:"explanation"`
	Prerolls       []any              `json:"prerolls"`
	RupTitle       string             `json:"rupTitle"`
	RupDescription string             `json:"rupDescription"`
	Settings       any                `json:"settings"`
	Settings2      *StationSettings2  `json:"settings2"`
	Station        StationDescription `json:"station"`
}

type StationsDashboard struct {
	DashboardID string    `json:"dashboardId"`
	Pumpkin     bool      `json:"pumpkin"`
	Stations    []Station `json:"stations"`
}

type StationSettings2 struct {
	Diversity  string `json:"diversity"`
	Language   string `json:"language"`
	MoodEnergy string `json:"moodEnergy"`
}

type SequenceItem struct {
	Liked           bool        `json:"liked"`
	Track           track.Track `json:"track"`
	TrackParameters any         `json:"trackParameters"`
	Type            string      `json:"type"`
}

type StationSequence struct {
	BatchID        string         `json:"batchId"`
	ID             StationID      `json:"id"`
	Pumpkin        bool           `json:"pumpkin"`
	RadioSessionID string         `json:"radioSessionId"`
	Sequence       []SequenceItem `json:"sequence"`
}

type StationFeedbackType string

const (
	FeedbackRadioStarted  StationFeedbackType = "radioStarted"
	FeedbackTrackStarted  StationFeedbackType = "trackStarted"
	FeedbackTrackFinished StationFeedbackType = "trackFinished"
	FeedbackSkip          StationFeedbackType = "skip"
)

type stationFeedback struct {
	Type               StationFeedbackType `json:"type"`
	Timestamp          int64               `json:"timestamp"`
	From               string              `json:"from"`
	TotalPlayedSeconds float64             `json:"totalPlayedSeconds,omitempty"`
	TrackID            string              `json:"trackId,omitempty"`
}

func newFeedback(t StationFeedbackType, from string, trackID string, total float64, ts int64) stationFeedback {
	return stationFeedback{Type: t, From: from, TrackID: trackID, TotalPlayedSeconds: total, Timestamp: ts}
}

// Utility to generate unix seconds
func nowUnix() int64 { return time.Now().Unix() }

// NewFeedbackPayload constructs payload used by RadioService.Feedback.
func NewFeedbackPayload(t StationFeedbackType, from, trackID string, total float64, ts int64) stationFeedback {
	if ts == 0 {
		ts = nowUnix()
	}
	return newFeedback(t, from, trackID, total, ts)
}
