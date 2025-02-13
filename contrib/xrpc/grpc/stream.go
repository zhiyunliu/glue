package grpc

import "encoding/json"

func unmarshaler(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}
