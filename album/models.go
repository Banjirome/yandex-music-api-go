package album

import (
	"time"

	"github.com/Banjirome/yandex-music-go/models"
	"github.com/Banjirome/yandex-music-go/track"
)

// Album mirrors C# YAlbum fields.
type Album struct {
	ActionButton             any             `json:"actionButton"`
	Artists                  []ArtistRef     `json:"artists"`
	Available                bool            `json:"available"`
	AvailableForMobile       bool            `json:"availableForMobile"`
	AvailableForOptions      []string        `json:"availableForOptions"`
	AvailableForPremiumUsers bool            `json:"availableForPremiumUsers"`
	AvailablePartially       bool            `json:"availablePartially"`
	BackgroundImageURL       string          `json:"backgroundImageUrl"`
	BackgroundVideoURL       string          `json:"backgroundVideoUrl"`
	Bests                    []string        `json:"bests"`
	Buy                      []string        `json:"buy"`
	ChildContent             bool            `json:"childContent"`
	ContentWarning           string          `json:"contentWarning"`
	CoverURI                 string          `json:"coverUri"`
	Cover                    any             `json:"cover"`
	CustomWave               any             `json:"customWave"`
	DerivedColors            any             `json:"derivedColors"`
	Description              string          `json:"description"`
	Disclaimers              []string        `json:"disclaimers"`
	Duplicates               []Album         `json:"duplicates"`
	HasTrailer               bool            `json:"hasTrailer"`
	Genre                    string          `json:"genre"`
	ID                       string          `json:"id"`
	Labels                   any             `json:"labels"`
	LikesCount               int             `json:"likesCount"`
	ListeningFinished        bool            `json:"listeningFinished"`
	MetaTagID                string          `json:"metaTagId"`
	MetaType                 string          `json:"metaType"`
	OgImage                  string          `json:"ogImage"`
	Pager                    *models.Pager   `json:"pager"`
	Prerolls                 []any           `json:"prerolls"`
	Recent                   bool            `json:"recent"`
	ReleaseDate              time.Time       `json:"releaseDate"`
	ShortDescription         string          `json:"shortDescription"`
	SortOrder                string          `json:"sortOrder"`
	StorageDir               string          `json:"storageDir"`
	Title                    string          `json:"title"`
	TrackCount               int             `json:"trackCount"`
	TrackPosition            any             `json:"trackPosition"`
	Trailer                  any             `json:"trailer"`
	Type                     string          `json:"type"`
	Version                  string          `json:"version"`
	VeryImportant            bool            `json:"veryImportant"`
	Volumes                  [][]track.Track `json:"volumes"`
	Year                     int             `json:"year"`
}

// ArtistRef is a lightweight reference to avoid import cycle (full Artist lives in artist package).
type ArtistRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
