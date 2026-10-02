package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testConfig(endpoint string) S3Config {
	return S3Config{
		Endpoint:        endpoint,
		Bucket:          "air",
		AccessKeyID:     "key",
		SecretAccessKey: "secret",
		AssetBaseURL:    "https://cdn.example.com/assets/",
	}
}

func TestNewS3RequiresConfig(t *testing.T) {
	tests := map[string]func(*S3Config){
		"endpoint":       func(c *S3Config) { c.Endpoint = "" },
		"bucket":         func(c *S3Config) { c.Bucket = "" },
		"access key id":  func(c *S3Config) { c.AccessKeyID = "" },
		"secret key":     func(c *S3Config) { c.SecretAccessKey = "" },
		"asset base url": func(c *S3Config) { c.AssetBaseURL = "" },
	}
	for name, unset := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := testConfig("https://example.r2.cloudflarestorage.com")
			unset(&cfg)
			if _, err := NewS3(cfg); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestPublicURL(t *testing.T) {
	s, err := NewS3(testConfig("https://example.r2.cloudflarestorage.com"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := s.PublicURL("abc"), "https://cdn.example.com/assets/abc"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPut(t *testing.T) {
	var got *http.Request
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got, body = r, string(b)
	}))
	defer srv.Close()

	s, err := NewS3(testConfig(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put(context.Background(), "abc", "image/png", 5, strings.NewReader("hello")); err != nil {
		t.Fatal(err)
	}

	if got.Method != http.MethodPut || got.URL.Path != "/air/assets/abc" {
		t.Errorf("got %s %s, want PUT /air/assets/abc", got.Method, got.URL.Path)
	}
	if body != "hello" {
		t.Errorf("body = %q", body)
	}
	if v := got.Header.Get("Content-Type"); v != "image/png" {
		t.Errorf("Content-Type = %q", v)
	}
	if v := got.Header.Get("Cache-Control"); v != cacheControl {
		t.Errorf("Cache-Control = %q", v)
	}
	// R2 rejects the SDK's default checksum headers and trailers.
	if v := got.Header.Get("X-Amz-Sdk-Checksum-Algorithm"); v != "" {
		t.Errorf("unexpected checksum algorithm %q", v)
	}
	if v := got.Header.Get("Content-Encoding"); v != "" {
		t.Errorf("unexpected Content-Encoding %q", v)
	}
}

func TestPutError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	s, err := NewS3(testConfig(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put(context.Background(), "abc", "image/png", 5, strings.NewReader("hello")); err == nil {
		t.Fatal("expected an error")
	}
}
