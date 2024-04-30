// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

type protobufBinding struct{}

type xprotobufBinding struct {
	protobufBinding
}

func (xprotobufBinding) Name() string {
	return "x-protobuf"
}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (protobufBinding) Marshal(obj interface{}) ([]byte, error) {
	msg, ok := obj.(proto.Message)
	if !ok {
		return nil, errors.New("obj is not ProtoMessage")
	}
	return proto.Marshal(msg)
}

func (protobufBinding) Unmarshal(body []byte, obj interface{}) error {
	msg, ok := obj.(proto.Message)
	if !ok {
		return errors.New("obj is not ProtoMessage")
	}
	if err := proto.Unmarshal(body, msg); err != nil {
		return err
	}
	// Here it's same to return validate(obj), but util now we can't add
	// `binding:""` to the struct which automatically generate by gen-proto
	return nil
	// return validate(obj)
}
