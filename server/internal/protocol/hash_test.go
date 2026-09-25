package protocol

import (
	"strings"
	"testing"
)

// Expected values were computed independently (Python hashlib) so the tests
// check the real algorithms, not just that the code agrees with itself.

func TestAssetHash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU"},
		{"abc", "abc", "ungWv48Bz-pBQUDeXa4iI7ADYaOWF3qctBD_YfIAFa0"},
		{"hello world", "hello world", "uU0nuZNNPgilLlLX2n2r-sSE7-N6U4DukIj3rOLvzek"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetHash([]byte(tt.input)); got != tt.want {
				t.Errorf("AssetHash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// The client rejects standard base64 ("+", "/", "=" padding).
// SHA-256("abc") in standard base64 contains all three, so it proves the encoding.
func TestAssetHashIsBase64URLWithoutPadding(t *testing.T) {
	got := AssetHash([]byte("abc"))

	if strings.ContainsAny(got, "+/=") {
		t.Errorf("AssetHash contains standard base64 characters: %q", got)
	}
	if len(got) != 43 { // 32 bytes -> 43 base64 chars without padding
		t.Errorf("len(AssetHash) = %d, want 43", len(got))
	}
}

func TestAssetHashIsDeterministic(t *testing.T) {
	first := AssetHash([]byte("same bytes"))
	second := AssetHash([]byte("same bytes"))
	if first != second {
		t.Error("AssetHash returned different values for the same input")
	}
	if AssetHash([]byte("a")) == AssetHash([]byte("b")) {
		t.Error("AssetHash returned the same value for different input")
	}
}

func TestAssetKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"abc", "abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"hello world", "hello world", "5eb63bbbe01eeed093cb22bb8f5acdc3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AssetKey([]byte(tt.input)); got != tt.want {
				t.Errorf("AssetKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
