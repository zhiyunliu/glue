// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"fmt"
	"net/url"

	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

type formBinding struct{}
type formPostBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Marshal(v interface{}) ([]byte, error) {
	mapVal, err := xtypes.AnyToMap(v)
	if err != nil {
		return nil, err
	}
	vals := url.Values{}

	for k, v := range mapVal {
		vals.Set(k, fmt.Sprint(v))
	}
	return bytesconv.StringToBytes(vals.Encode()), nil
}

func (formBinding) Unmarshal(body []byte, obj interface{}) (err error) {
	vs, err := url.ParseQuery(string(body))
	if err != nil {
		return
	}
	if err := mapForm(obj, vs); err != nil {
		return err
	}
	return nil
}

func (formPostBinding) Name() string {
	return "x-www-form-urlencoded"
}

func (formPostBinding) Marshal(v interface{}) ([]byte, error) {
	mapVal, err := xtypes.AnyToMap(v)
	if err != nil {
		return nil, err
	}
	vals := url.Values{}

	for k, v := range mapVal {
		vals.Set(k, fmt.Sprint(v))
	}
	return bytesconv.StringToBytes(vals.Encode()), nil
}

func (formPostBinding) Unmarshal(body []byte, obj interface{}) (err error) {
	vs, err := url.ParseQuery(string(body))
	if err != nil {
		return
	}
	if err := mapForm(obj, vs); err != nil {
		return err
	}
	return nil
}

// func (formMultipartBinding) Name() string {
// 	return "multipart/form-data"
// }

// func (formMultipartBinding) Marshal(v interface{}) ([]byte, error) {
// 	mapVal, err := xtypes.AnyToMap(v)
// 	if err != nil {
// 		return nil, err
// 	}

// 	byteBuffer := &bytes.Buffer{}
// 	writer := multipart.NewWriter(byteBuffer)

// 	for k, v := range mapVal {
// 		writer.WriteField(k, fmt.Sprint(v))
// 	}
// 	writer.Close()
// 	return byteBuffer.Bytes(), nil
// }

// func (r formMultipartBinding) Unmarshal(body []byte, obj interface{}) (err error) {
// 	const (
// 		defaultMaxMemory = 32 << 20 // 32 MB
// 	)
// 	boundary, ok := params["boundary"]
// 	if !ok {
// 		return http.ErrMissingBoundary
// 	}

// 	reader := multipart.NewReader(bytes.NewReader(body), boundary)

// 	multiForm, err := reader.ReadForm(defaultMaxMemory)
// 	if err != nil {
// 		return
// 	}

// 	if err := mapForm(obj, multiForm.Value); err != nil {
// 		return err
// 	}
// 	return nil
// }
