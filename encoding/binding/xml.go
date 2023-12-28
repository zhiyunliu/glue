// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"encoding/xml"
	"io"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func decodeXML(r io.Reader, obj any) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

func (xmlBinding) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (xmlBinding) Unmarshal(body []byte, obj interface{}) error {
	return decodeXML(bytes.NewReader(body), obj)
}
