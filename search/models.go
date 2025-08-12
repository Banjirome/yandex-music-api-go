package search

// Search основной объект результата поиска.
type Search struct {
	Albums            *SearchResult[SearchAlbum]    `json:"albums,omitempty"`
	Artists           *SearchResult[SearchArtist]   `json:"artists,omitempty"`
	Best              *SearchBest                   `json:"best,omitempty"`
	MisspellCorrected bool                          `json:"misspellCorrected"`
	MisspellOriginal  string                        `json:"misspellOriginal"`
	MisspellResult    string                        `json:"misspellResult"`
	NoCorrect         bool                          `json:"noCorrect"`
	Page              int                           `json:"page"`
	PerPage           int                           `json:"perPage"`
	Playlists         *SearchResult[SearchPlaylist] `json:"playlists,omitempty"`
	PodcastEpisode    *SearchResult[SearchTrack]    `json:"podcast_episodes,omitempty"`
	SearchRequestID   string                        `json:"searchRequestId"`
	Text              string                        `json:"text"`
	Tracks            *SearchResult[SearchTrack]    `json:"tracks,omitempty"`
	Type              Type                          `json:"type"`
	Users             *SearchResult[SearchUser]     `json:"users,omitempty"`
	Videos            *SearchResult[SearchVideo]    `json:"videos,omitempty"`
}

// Suggest модели (упрощённо)
type Suggest struct {
	Suggestions []SuggestItem `json:"suggestions"`
}

type SuggestItem struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type SearchResult[T any] struct {
	Order   int `json:"order"`
	PerPage int `json:"perPage"`
	Results []T `json:"results"`
	Total   int `json:"total"`
}

type SearchAlbum struct { /* минимально */
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SearchArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// PopularTracks omitted for brevity
}

type SearchPlaylist struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SearchTrack struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Artists  []TrackArtist   `json:"artists,omitempty"`
	Albums   []TrackAlbumRef `json:"albums,omitempty"`
	Explicit bool            `json:"explicit"`
}

type TrackArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrackAlbumRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SearchUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SearchVideo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SearchBest struct {
	Type   Type        `json:"type"`
	Result interface{} `json:"result"`
}

type Type string

const (
	TypeTrack          Type = "track"
	TypeAlbum          Type = "album"
	TypeArtist         Type = "artist"
	TypePlaylist       Type = "playlist"
	TypePodcastEpisode Type = "podcastEpisode"
	TypeVideo          Type = "video"
	TypeUser           Type = "user"
)
