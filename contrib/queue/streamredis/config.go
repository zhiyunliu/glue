package streamredis

type ProductOptions struct {
	StreamMaxLength int64 `json:"stream_max_len" yaml:"stream_max_len"`
}

type ConsumerOptions struct {
	Concurrency     int `json:"concurrency" yaml:"concurrency"`
	BufferSize      int `json:"buffer_size" yaml:"buffer_size"`
	BlockingTimeout int `json:"blocking_timeout" yaml:"blocking_timeout"`
}
