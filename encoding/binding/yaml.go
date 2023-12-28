// Copyright 2018 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

type yamlBinding struct{}

type xyamlBinding struct {
	yamlBinding
}

func (xyamlBinding) Name() string {
	return "x-yaml"
}

func (yamlBinding) Name() string {
	return "yaml"
}

func decodeYAML(r io.Reader, obj any) error {
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

func (yamlBinding) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (yamlBinding) Unmarshal(body []byte, obj interface{}) error {
	return decodeYAML(bytes.NewReader(body), obj)
}
