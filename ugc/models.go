package ugc

// Upload mirrors YUgcUpload (poll-result, post-target, ugc-track-id).
type Upload struct {
	PollResult string `json:"poll-result"`
	PostTarget string `json:"post-target"`
	UgcTrackID string `json:"ugc-track-id"`
}
