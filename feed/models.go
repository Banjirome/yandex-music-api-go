package feed

// NOTE: This is a minimal mirror of YFeed subset (parity-first; can expand later).

type Feed struct {
	Days []FeedDay `json:"days"`
}

type FeedDay struct {
	Events              []FeedEventTitled  `json:"events"`
	TracksToPlayWithAds []FeedTrackWithAds `json:"tracksToPlayWithAds"`
}

type FeedTrackWithAds struct {
	Type string `json:"type"`
}

type FeedEvent struct {
	EventType string `json:"type"`
}

type FeedEventTitled struct {
	FeedEvent
	Title       []FeedEventTitle `json:"title"`
	TypeForFrom string           `json:"typeForFrom"`
}

type FeedEventTitle struct {
	Type string `json:"type"`
}
