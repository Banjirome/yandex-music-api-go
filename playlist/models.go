package playlist

import "time"

// Playlist минимизированная типизация из оригинальной модели YPlaylist.
type Playlist struct {
	Kind                  string           `json:"kind"`
	Title                 string           `json:"title"`
	Description           string           `json:"description,omitempty"`
	Uid                   string           `json:"uid,omitempty"`
	Owner                 *Owner           `json:"owner,omitempty"`
	TrackCount            int              `json:"trackCount,omitempty"`
	DurationMs            int64            `json:"durationMs,omitempty"`
	Tracks                []TrackContainer `json:"tracks,omitempty"`
	Revision              int              `json:"revision,omitempty"`
	Snapshot              int              `json:"snapshot,omitempty"`
	LikesCount            int              `json:"likesCount,omitempty"`
	Created               *time.Time       `json:"created,omitempty"`
	GeneratedPlaylistType string           `json:"generatedPlaylistType,omitempty"`
}

// Generated playlist types (subset) mirroring C# YGeneratedPlaylistType enum.
const (
	GeneratedPlaylistOfTheDay   = "PlaylistOfTheDay"
	GeneratedPlaylistNeverHeard = "NeverHeard"   // DejaVu
	GeneratedPlaylistRecent     = "RecentTracks" // Premiere
	GeneratedPlaylistMissed     = "MissedLikes"  // Тайник
	GeneratedPlaylistKinopoisk  = "Kinopoisk"    // Кинопоиск
)

// Owner владелец плейлиста.
type Owner struct {
	Uid string `json:"uid"`
}

// TrackContainer связь трека и метаданных.
type TrackContainer struct {
	ID    string `json:"id,omitempty"`
	Track *Track `json:"track,omitempty"`
}

// Track упрощённая модель трека (часть полей).
type Track struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Artists []TrackArtist `json:"artists,omitempty"`
	Albums  []TrackAlbum  `json:"albums,omitempty"`
}

type TrackArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TrackAlbum struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// TrackKey аналог YTrackAlbumPair.
type TrackKey struct {
	Id      string `json:"id"`
	AlbumId string `json:"albumId,omitempty"`
}

// ChangeRequest аналог YPlaylistChange.
type ChangeRequest struct {
	Operation string     `json:"operation"`
	At        int        `json:"at,omitempty"`
	From      int        `json:"from,omitempty"`
	To        int        `json:"to,omitempty"`
	Tracks    []TrackKey `json:"tracks,omitempty"`
}

// ChangeEnvelope оболочка для отправки списка изменений.
type ChangeEnvelope struct {
	Changes []ChangeRequest `json:"changes"`
}
