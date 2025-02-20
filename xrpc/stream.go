package xrpc

// StreamUnmarshaler is a callback function that can be used to unmarshal a received message.
type StreamUnmarshaler func([]byte, any) error

type ClientStreamRequest interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
}
type ServerStreamRequest interface {
	Send(obj any) (err error)
}

//StreamRequest is an interface that represents a stream request.
type BidirectionalStreamRequest interface {
	ClientStreamRequest
	ServerStreamRequest
}

// StreamClient is an interface that represents a stream client.
type BidirectionalStreamClient interface {
	ClientStreamRequest
	ServerStreamRequest
	CloseSend() error
}

type ClientStreamClient interface {
	Send(obj any) (err error)
}

type ServerStreamClient interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
}

// StreamProcessor is a callback function that can be used to process a stream client.
type BidirectionalStreamProcessor func(BidirectionalStreamClient) error
type ClientStreamProcessor func(ClientStreamClient) (err error)
type ServerStreamProcessor func(ServerStreamClient) (err error)
