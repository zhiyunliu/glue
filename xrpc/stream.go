package xrpc

import (
	"fmt"
	"reflect"

	"golang.org/x/sync/errgroup"
)

// StreamUnmarshaler is a callback function that can be used to unmarshal a received message.
type StreamUnmarshaler func([]byte, any) error

// ClientStreamRequest is an interface that represents a client stream request.
type ClientStreamRequest interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
}

// ServerStreamRequest is an interface that represents a server stream request.
type ServerStreamRequest interface {
	Send(obj any) (err error)
}

// StreamRequest is an interface that represents a stream request.
type BidirectionalStreamRequest interface {
	ClientStreamRequest
	ServerStreamRequest
}

// BidirectionalStreamClient is an interface that represents a bidirectional stream client.
type BidirectionalStreamClient interface {
	ClientStreamRequest
	ServerStreamRequest
	CloseSend() error
}

// ClientStreamClient is an interface that represents a stream client.
type ClientStreamClient interface {
	Send(obj any) (err error)
}

// ServerStreamClient is an interface that represents a stream server.
type ServerStreamClient interface {
	Recv(obj any, opts ...StreamRevcOption) (closed bool, err error)
}

// BidirectionalStreamProcessor is a callback function that can be used to process a bidirectional stream.
type BidirectionalStreamProcessor func(BidirectionalStreamClient) error

// ClientStreamProcessor is a callback function that can be used to process a stream client.
type ClientStreamProcessor func(ClientStreamClient) (err error)

// ServerStreamProcessor is a callback function that can be used to process a stream server.
type ServerStreamProcessor func(ServerStreamClient) (err error)

// DefaultProcessor is a default implementation of StreamProcessor.
type DefaultProcessor struct{}

// ClientStreamObjects 客户端流对象接口
type ClientStreamObjects interface {
	GetObjects() []any
}

// ClientStreamChan 客户端流通道接口
type ClientStreamChan interface {
	GetObject() <-chan any
}

// BuildDefaultClientStreamProcess 构建默认的客户端流处理器
func BuildDefaultClientStreamProcess(input any) (processor ClientStreamProcessor, err error) {
	refval := reflect.ValueOf(input)
	if refval.Kind() == reflect.Ptr && refval.IsNil() {
		return nil, fmt.Errorf("input is nil")
	}

	if input, ok := input.(ClientStreamObjects); ok {
		return buildClientSliceProcess(reflect.ValueOf(input.GetObjects()))
	}

	if input, ok := input.(ClientStreamChan); ok {
		return buildClientChanProcess(input)
	}

	refType := refval.Type()
	if refType.Kind() != reflect.Array &&
		refType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input is not array or slice")
	}

	//[]byte数组
	if refType.Kind() == reflect.Slice &&
		refType.Elem().Kind() == reflect.Uint8 {
		return nil, fmt.Errorf("input is []byte array")
	}

	return buildClientSliceProcess(refval)
}

func buildClientChanProcess(channel ClientStreamChan) (processor ClientStreamProcessor, err error) {
	return func(client ClientStreamClient) error {
		errGroup := errgroup.Group{}
		//调用grpc服务
		errGroup.Go(func() error {
			dataChan := channel.GetObject()
			for item := range dataChan {
				if err := client.Send(item); err != nil {
					return err
				}
			}
			return nil
		})
		return errGroup.Wait()
	}, nil
}

func buildClientSliceProcess(refval reflect.Value) (processor ClientStreamProcessor, err error) {

	return func(client ClientStreamClient) error {
		errGroup := errgroup.Group{}
		//调用grpc服务
		errGroup.Go(func() error {
			for i, valLen := 0, refval.Len(); i < valLen; i++ {
				if err := client.Send(refval.Index(i).Interface()); err != nil {
					return err
				}
			}
			return nil
		})
		return errGroup.Wait()
	}, nil

}
