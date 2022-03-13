package xtypes

import (
	"bytes"
	"encoding/json"
)

type SMap map[string]string

func (m SMap) Scan(obj interface{}) error {
	bytes, _ := json.Marshal(m)
	return json.Unmarshal(bytes, obj)
}

func (m SMap) Read(p []byte) (n int, err error) {
	dataBytes, _ := json.Marshal(m)
	return bytes.NewReader(dataBytes).Read(p)
}

func (m SMap) Get(name string) string {
	if v, ok := m[name]; ok {
		return v
	}
	return ""
}
