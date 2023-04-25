package metadata

import (
	"context"
	"fmt"
	"strings"
)

// Metadata is our way of representing request headers internally.
// They're used at the RPC level and translate back and forth
// from Transport headers.
type Metadata map[string]interface{}

// New creates an MD from a given key-values map.
func New(mds ...map[string]interface{}) Metadata {
	md := Metadata{}
	for _, m := range mds {
		for k, v := range m {
			md.Set(k, v)
		}
	}
	return md
}

// Get returns the value associated with the passed key.
func (m Metadata) Get(key string) string {
	k := strings.ToLower(key)
	switch v := m[k].(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

// Set stores the key-value pair.
func (m Metadata) Set(key string, value interface{}) {
	if key == "" || value == "" {
		return
	}
	k := strings.ToLower(key)
	m[k] = value
}

// Range iterate over element in metadata.
func (m Metadata) Range(f func(k string, v interface{}) bool) {
	for k, v := range m {
		ret := f(k, v)
		if !ret {
			break
		}
	}
}

// Clone returns a deep copy of Metadata
func (m Metadata) Clone() Metadata {
	md := Metadata{}
	for k, v := range m {
		md[k] = v
	}
	return md
}

type serverMetadataKey struct{}

// NewServerContext creates a new context with client md attached.
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

// FromServerContext returns the server metadata in ctx if it exists.
func FromServerContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(serverMetadataKey{}).(Metadata)
	return md, ok
}

type clientMetadataKey struct{}

// NewClientContext creates a new context with client md attached.
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, clientMetadataKey{}, md)
}

// FromClientContext returns the client metadata in ctx if it exists.
func FromClientContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(clientMetadataKey{}).(Metadata)
	return md, ok
}

// MergeToClientContext merge new metadata into ctx.
func MergeToClientContext(ctx context.Context, cmd Metadata) context.Context {
	md, _ := FromClientContext(ctx)
	md = md.Clone()
	for k, v := range cmd {
		md[k] = v
	}
	return NewClientContext(ctx, md)
}
