package api

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	ProtocolHTTP1 = "http1"
	ProtocolH2C   = "h2c"
	ProtocolH2    = "h2"
)

var (
	ErrUnsupportedProtocol   = errors.New("api: unsupported protocol")
	ErrInvalidProtocolConfig = errors.New("api: invalid protocol config")
)

func buildProtocolHandler(handler http.Handler, cfg Config) (http.Handler, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.HttpProtocol)) {
	case "", "http", "http1", "h1", ProtocolH2:
		return handler, nil
	case ProtocolH2C:
		return h2c.NewHandler(handler, buildHTTP2Server(cfg.H2C.Http2Config)), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProtocol, cfg.HttpProtocol)
	}
}

func configureProtocolServer(server *http.Server, cfg Config) error {
	switch strings.ToLower(strings.TrimSpace(cfg.HttpProtocol)) {
	case "", "http", "http1", "h1", ProtocolH2C:
		return nil
	case ProtocolH2:
		if cfg.H2.CertFile == "" || cfg.H2.KeyFile == "" {
			return fmt.Errorf("%w: h2 requires cert_file and key_file", ErrInvalidProtocolConfig)
		}
		server.TLSConfig = buildH2TLSConfig(server.TLSConfig)
		return http2.ConfigureServer(server, buildHTTP2Server(cfg.H2.Http2Config))
	default:
		return nil
	}
}

func serveProtocol(server *http.Server, listener net.Listener, cfg Config) error {
	if strings.EqualFold(strings.TrimSpace(cfg.HttpProtocol), ProtocolH2) {
		return server.ServeTLS(listener, cfg.H2.CertFile, cfg.H2.KeyFile)
	}
	return server.Serve(listener)
}

func buildH2TLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		cfg = &tls.Config{}
	} else {
		cfg = cfg.Clone()
	}
	cfg.NextProtos = []string{"h2", "http/1.1"}
	return cfg
}

func buildHTTP2Server(cfg Http2Config) *http2.Server {
	return &http2.Server{
		MaxHandlers:                  0,
		MaxConcurrentStreams:         cfg.MaxConcurrentStreams,
		MaxDecoderHeaderTableSize:    cfg.MaxDecoderHeaderTableSize,
		MaxEncoderHeaderTableSize:    cfg.MaxEncoderHeaderTableSize,
		MaxReadFrameSize:             cfg.MaxReadFrameSize,
		MaxUploadBufferPerConnection: cfg.MaxUploadBufferPerConnection,
		MaxUploadBufferPerStream:     cfg.MaxUploadBufferPerStream,
		IdleTimeout:                  secondsDuration(cfg.IdleTimeout),
		ReadIdleTimeout:              secondsDuration(cfg.ReadIdleTimeout),
		PingTimeout:                  secondsDuration(cfg.PingTimeout),
		WriteByteTimeout:             secondsDuration(cfg.WriteByteTimeout),
	}
}

func secondsDuration(seconds uint) time.Duration {
	const maxSeconds = uint64(1<<63-1) / uint64(time.Second)
	if uint64(seconds) > maxSeconds {
		return time.Duration(1<<63 - 1)
	}
	return time.Duration(seconds) * time.Second
}
