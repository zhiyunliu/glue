package xrpc

type StreamUnmarshaler func([]byte, any) error
type StreamProcessor func(StreamClient) error

type StreamRequest interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
	Send(obj any) (err error)
}

type StreamClient interface {
	StreamRequest
	CloseSend() error
}

type StreamRecvOptions struct {
	Unmarshal StreamUnmarshaler
}

type StreamRevcOption func(*StreamRecvOptions)

func WithStreamUnmarshal(callback StreamUnmarshaler) StreamRevcOption {
	return func(sro *StreamRecvOptions) {
		sro.Unmarshal = callback
	}
}
