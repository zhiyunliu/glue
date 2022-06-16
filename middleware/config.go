package middleware

import (
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name string     `json:"name" yaml:"name"`
	Data RawMessage `json:"data" yaml:"data"`
}

var _ json.Marshaler = (*RawMessage)(nil)
var _ json.Unmarshaler = (*RawMessage)(nil)

var _ yaml.Marshaler = (*RawMessage)(nil)
var _ yaml.Unmarshaler = (*RawMessage)(nil)

type RawMessage struct {
	Data  []byte
	Codec string
}

// MarshalJSON returns m as the JSON encoding of m.
func (m *RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	m.Codec = "json"
	return m.Data, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	m.Codec = "json"
	m.Data = append((m.Data)[0:0], data...)
	return nil
}

func (m *RawMessage) MarshalYAML() (interface{}, error) {
	if m == nil {
		return []byte("null"), nil
	}
	m.Codec = "yaml"
	return m.Data, nil
}

func (m *RawMessage) UnmarshalYAML(value *yaml.Node) error {
	if m == nil {
		return errors.New("yaml.RawMessage: UnmarshalYAML on nil pointer")
	}
	m.Codec = "yaml"
	m.Data = []byte(value.Value)
	return nil
}
