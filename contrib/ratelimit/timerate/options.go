package timerate

type options struct {
	CountLimit int   `json:"count_limit,omitempty"`
	Every      int64 `json:"every,omitempty"`
}
