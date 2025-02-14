package xrpc

// StreamUnmarshaler is a callback function that can be used to unmarshal a received message.
type StreamUnmarshaler func([]byte, any) error

//StreamRequest is an interface that represents a stream request.
type StreamRequest interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
	Send(obj any) (err error)
}

// StreamClient is an interface that represents a stream client.
type StreamClient interface {
	StreamRequest
	CloseSend() error
}

// StreamProcessor is a callback function that can be used to process a stream client.
type StreamProcessor func(StreamClient) error

// StreamRecvOptions is a struct that contains options for receiving messages.
type StreamRecvOptions struct {
	Unmarshal StreamUnmarshaler
}

// StreamRevcOption is a function that sets an option for receiving messages.
type StreamRevcOption func(*StreamRecvOptions)

// WithStreamUnmarshal sets the callback function to unmarshal a received message.
func WithStreamUnmarshal(callback StreamUnmarshaler) StreamRevcOption {
	return func(sro *StreamRecvOptions) {
		sro.Unmarshal = callback
	}
}
