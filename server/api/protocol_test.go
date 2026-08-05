package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

func TestBuildProtocolHandlerKeepsHTTP1HandlerByDefault(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})

	got, err := buildProtocolHandler(handler, Config{})
	if err != nil {
		t.Fatalf("buildProtocolHandler() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	got.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("handler was not called")
	}
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusAccepted)
	}
}

func TestBuildProtocolHandlerSupportsH2C(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Fatalf("request proto = %s, want HTTP/2", r.Proto)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	wrapped, err := buildProtocolHandler(handler, Config{HttpProtocol: ProtocolH2C})
	if err != nil {
		t.Fatalf("buildProtocolHandler() error = %v", err)
	}

	server := httptest.NewServer(wrapped)
	t.Cleanup(server.Close)

	client := &http.Client{Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, network, addr)
		},
	}}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if resp.ProtoMajor != 2 {
		t.Fatalf("response proto = %s, want HTTP/2", resp.Proto)
	}
}

func TestBuildProtocolHandlerKeepsH2Handler(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	got, err := buildProtocolHandler(handler, Config{HttpProtocol: ProtocolH2})
	if err != nil {
		t.Fatalf("buildProtocolHandler() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	got.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("handler was not called")
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestConfigureProtocolServerSupportsH2TLS(t *testing.T) {
	server := &http.Server{Handler: http.NotFoundHandler()}

	err := configureProtocolServer(server, Config{
		HttpProtocol: ProtocolH2,
		H2: H2Config{
			CertFile: "server.crt",
			KeyFile:  "server.key",
			Http2Config: Http2Config{
				MaxConcurrentStreams: 128,
			},
		},
	})
	if err != nil {
		t.Fatalf("configureProtocolServer() error = %v", err)
	}
	if server.TLSConfig == nil {
		t.Fatal("TLSConfig is nil")
	}
	if got, want := server.TLSConfig.NextProtos, []string{"h2", "http/1.1"}; !equalStrings(got, want) {
		t.Fatalf("NextProtos = %v, want %v", got, want)
	}
}

func TestConfigureProtocolServerRequiresH2CertificateFiles(t *testing.T) {
	err := configureProtocolServer(&http.Server{}, Config{HttpProtocol: ProtocolH2})
	if err == nil {
		t.Fatal("configureProtocolServer() error is nil")
	}
	if !errors.Is(err, ErrInvalidProtocolConfig) {
		t.Fatalf("configureProtocolServer() error = %v, want ErrInvalidProtocolConfig", err)
	}
}

func TestServeProtocolSupportsTLSH2(t *testing.T) {
	certFile, keyFile := writeTestCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Fatalf("request proto = %s, want HTTP/2", r.Proto)
		}
		w.WriteHeader(http.StatusNoContent)
	})}
	cfg := Config{HttpProtocol: ProtocolH2, H2: H2Config{CertFile: certFile, KeyFile: keyFile}}
	if err := configureProtocolServer(server, cfg); err != nil {
		t.Fatalf("configureProtocolServer() error = %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- serveProtocol(server, listener, cfg)
	}()
	t.Cleanup(func() {
		_ = server.Close()
		if err := <-errCh; err != nil && err != http.ErrServerClosed {
			t.Fatalf("serveProtocol() error = %v", err)
		}
	})

	client := &http.Client{Transport: &http2.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get("https://" + listener.Addr().String())
	if err != nil {
		t.Fatalf("client.Get() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if resp.ProtoMajor != 2 {
		t.Fatalf("response proto = %s, want HTTP/2", resp.Proto)
	}
}

func TestBuildProtocolHandlerRejectsUnknownProtocol(t *testing.T) {
	_, err := buildProtocolHandler(http.NotFoundHandler(), Config{HttpProtocol: "ftp"})
	if err == nil {
		t.Fatal("buildProtocolHandler() error is nil")
	}
	if !errors.Is(err, ErrUnsupportedProtocol) {
		t.Fatalf("buildProtocolHandler() error = %v, want ErrUnsupportedProtocol", err)
	}
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func writeTestCertificate(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	certFile := writePEMFile(t, "server-*.crt", "CERTIFICATE", certDER)
	keyFile := writePEMFile(t, "server-*.key", "RSA PRIVATE KEY", keyDER)
	return certFile, keyFile
}

func writePEMFile(t *testing.T, pattern string, blockType string, bytes []byte) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatalf("os.CreateTemp() error = %v", err)
	}
	if err := pem.Encode(file, &pem.Block{Type: blockType, Bytes: bytes}); err != nil {
		_ = file.Close()
		t.Fatalf("pem.Encode() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("file.Close() error = %v", err)
	}
	return file.Name()
}
