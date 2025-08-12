package queue

// Context mirrors YContext (type/id/login/description)
type Context struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Login       string `json:"login"`
	Description string `json:"description"`
}

// TrackID minimal representation referencing a track inside queue (albumId, id)
type TrackID struct {
	AlbumID string `json:"albumId"`
	ID      string `json:"id"`
}

// Queue mirrors YQueue
type Queue struct {
	ID            string    `json:"id"`
	Context       *Context  `json:"context"`
	Tracks        []TrackID `json:"tracks"`
	CurrentIndex  *int      `json:"currentIndex"`
	Modified      string    `json:"modified"`
	From          string    `json:"from"`
	IsInteractive bool      `json:"isInteractive"`
}

// NewQueue mirrors YNewQueue (id + modified)
type NewQueue struct {
	ID       string `json:"id"`
	Modified string `json:"modified"`
}

// UpdatedQueue mirrors YUpdatedQueue
type UpdatedQueue struct {
	Status          string `json:"status"`
	MostRecentQueue bool   `json:"mostRecentQueue"`
}

// QueueItem mirrors YQueueItem
type QueueItem struct {
	ID             string   `json:"id"`
	Context        *Context `json:"context"`
	InitialContext *Context `json:"initialContext"`
	Modified       string   `json:"modified"`
}

// QueueItemsContainer mirrors YQueueItemsContainer
type QueueItemsContainer struct {
	Queues []QueueItem `json:"queues"`
}
