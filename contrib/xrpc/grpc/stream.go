package grpc

import "encoding/json"

func defaultUnmarshaler(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}
