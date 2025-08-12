package landing

// Minimal models mirroring C# YLanding / YLandingBlock / YChildrenLanding for parity.
// Entities are kept as raw JSON (map[string]any) to avoid premature over-typing while
// still exposing structured block metadata. This can be expanded later without breaking
// existing code (adding dedicated typed entity structs).

type Landing struct {
	Blocks             []LandingBlock `json:"blocks"`
	ContentID          string         `json:"contentId"`
	HeaderSpecialBlock any            `json:"headerSpecialBlock"`
	Pumpkin            bool           `json:"pumpkin"`
}

type LandingBlock struct {
	ID                 string           `json:"id"`
	BackgroundImageURL string           `json:"backgroundImageUrl"`
	BackgroundVideoURL string           `json:"backgroundVideoUrl"`
	Data               any              `json:"data"`
	Description        string           `json:"description"`
	Entities           []map[string]any `json:"entities"`
	PlayContext        any              `json:"playContext"`
	ViewAllURL         string           `json:"viewAllUrl"`
	ViewAllURLScheme   string           `json:"viewAllUrlScheme"`
	Title              string           `json:"title"`
	Type               string           `json:"type"`
	TypeForFrom        string           `json:"typeForFrom"`
}

type ChildrenLanding struct {
	Title      string         `json:"title"`
	RupEnabled bool           `json:"rupEnabled"`
	Blocks     []LandingBlock `json:"blocks"`
}

// BlockType corresponds to YLandingBlockType enum string values in C# (EnumMember.Value or name).
type BlockType string

const (
	BlockChart              BlockType = "Chart"
	BlockClientWidget       BlockType = "client-widget"
	BlockCategoriesTab      BlockType = "categories-tab"
	BlockEditorialPlaylists BlockType = "editorial-playlists"
	BlockMenu               BlockType = "Menu"
	BlockMixes              BlockType = "Mixes"
	BlockNewReleases        BlockType = "new-releases"
	BlockNewPlaylists       BlockType = "new-playlists"
	BlockPersonalPlaylists  BlockType = "personal-playlists"
	BlockPlayContexts       BlockType = "play-contexts"
	BlockPlaylists          BlockType = "Playlists"
	BlockPodcasts           BlockType = "Podcasts"
	BlockPromotions         BlockType = "Promotions"
	BlockRadio              BlockType = "Radio"
)
