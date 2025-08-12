package library

import "time"

// Section enumerations
// (These map to path segments and are lowercased)
type Section string
type SectionType string

const (
	SectionAlbums    Section = "albums"
	SectionArtists   Section = "artists"
	SectionPlaylists Section = "playlists"
	SectionTracks    Section = "tracks"

	SectionTypeLikes    SectionType = "likes"
	SectionTypeDislikes SectionType = "dislikes"
)

// LibraryTracks mirrors YLibraryTracks -> YLibrary with tracks playlist metadata.
type LibraryTracks struct {
	Library *Library `json:"library"`
}

type Library struct {
	PlaylistUUID string         `json:"playlistUuid"`
	Revision     int            `json:"revision"`
	Tracks       []LibraryTrack `json:"tracks"`
	UID          string         `json:"uid"`
}

type LibraryTrack struct {
	AlbumID   string    `json:"albumId"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type LibraryAlbum struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type LibraryPlaylists struct {
	Playlist  any       `json:"playlist"` // placeholder until playlist model ported fully
	Timestamp time.Time `json:"timestamp"`
}

// Recently listened context models
type RecentlyListenedContext struct {
	Contexts    []RecentlyListened `json:"contexts"`
	OtherTracks []any              `json:"otherTracks"`
}

type RecentlyListened struct {
	Client      string          `json:"client"`
	Context     PlayContextType `json:"context"`
	ContextItem string          `json:"contextItem"`
	Tracks      []ListenedTrack `json:"tracks"`
}

type ListenedTrack struct {
	TrackID   TrackID   `json:"trackId"`
	TimeStamp time.Time `json:"timeStamp"`
}

// TrackID minimal structure referencing a track (AlbumID + ID)
type TrackID struct {
	ID      string `json:"id"`
	AlbumID string `json:"albumId"`
}

// PlayContextType enumerates allowed recently listened context types.
type PlayContextType string

const (
	PlayContextAlbum    PlayContextType = "album"
	PlayContextArtist   PlayContextType = "artist"
	PlayContextPlaylist PlayContextType = "playlist"
)
