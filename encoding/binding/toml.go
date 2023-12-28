// Copyright 2022 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"io"

	"github.com/pelletier/go-toml/v2"
)

type tomlBinding struct{}

func (tomlBinding) Name() string {
	return "toml"
}

func decodeToml(r io.Reader, obj any) error {
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return decoder.Decode(obj)
}

func (tomlBinding) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

func (tomlBinding) Unmarshal(body []byte, obj interface{}) error {
	return decodeToml(bytes.NewReader(body), obj)
}
