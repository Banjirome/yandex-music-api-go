package artist

import (
	"time"

	"github.com/Banjirome/yandex-music-go/album"
	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/track"
)

type Artist struct {
	ActionButton         any            `json:"actionButton"`
	Available            bool           `json:"available"`
	Composer             bool           `json:"composer"`
	Countries            []string       `json:"countries"`
	Counts               *ArtistCounts  `json:"counts"`
	Cover                any            `json:"cover"`
	DbAliases            []string       `json:"dbAliases"`
	Decomposed           []any          `json:"decomposed"`
	DerivedColors        any            `json:"derivedColors"`
	Description          any            `json:"description"`
	Deprecation          any            `json:"deprecation"`
	Disclaimers          []string       `json:"disclaimers"`
	EndDate              string         `json:"endDate"`
	EnWikipediaLink      string         `json:"enWikipediaLink"`
	ExtraActions         []any          `json:"extraActions"`
	Genres               []string       `json:"genres"`
	ID                   string         `json:"id"`
	InitDate             string         `json:"initDate"`
	LikesCount           int            `json:"likesCount"`
	Links                []any          `json:"links"`
	Name                 string         `json:"name"`
	NoPicturesFromSearch bool           `json:"noPicturesFromSearch"`
	OgImage              string         `json:"ogImage"`
	Ratings              *ArtistRatings `json:"ratings"`
	TicketsAvailable     bool           `json:"ticketsAvailable"`
	Timestamp            time.Time      `json:"timestamp"`
	Various              bool           `json:"various"`
	YaMoneyID            string         `json:"yaMoneyId"`
	HasTrailer           bool           `json:"hasTrailer"`
	Trailer              any            `json:"trailer"`
}

type ArtistRatings struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Week  int `json:"week"`
}

type ArtistCounts struct {
	AlsoAlbums   int `json:"alsoAlbums"`
	AlsoTracks   int `json:"alsoTracks"`
	DirectAlbums int `json:"directAlbums"`
	Tracks       int `json:"tracks"`
}

type ArtistBriefInfo struct {
	ActionButton        any           `json:"actionButton"`
	Albums              []album.Album `json:"albums"`
	AllCovers           []any         `json:"allCovers"`
	AlsoAlbums          []album.Album `json:"alsoAlbums"`
	Artist              *Artist       `json:"artist"`
	BackgroundVideoURL  string        `json:"backgroundVideoUrl"`
	BandlinkScannerLink any           `json:"bandlinkScannerLink"`
	Clips               []any         `json:"clips"`
	Concerts            []any         `json:"concerts"`
	CustomWave          any           `json:"customWave"`
	ExtraActions        []any         `json:"extraActions"`
	HasPromotions       bool          `json:"hasPromotions"`
	HasTrailer          bool          `json:"hasTrailer"`
	LastReleaseIDs      []string      `json:"lastReleaseIds"`
	LastReleases        []album.Album `json:"lastReleases"`
	PlaylistIDs         []any         `json:"playlistIds"`
	Playlists           []any         `json:"playlists"`
	PopularTracks       []track.Track `json:"popularTracks"`
	SimilarArtists      []Artist      `json:"similarArtists"`
	Stats               any           `json:"stats"`
	Videos              []any         `json:"videos"`
	Vinyls              []any         `json:"vinyls"`
	Links               []any         `json:"links"`
}

type TracksPage struct {
	Pager  *models.Pager `json:"pager"`
	Tracks []track.Track `json:"tracks"`
}
