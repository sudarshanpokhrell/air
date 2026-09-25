package protocol

import "time"

// Asset is a file in manifest
type Asset struct {
	Hash          string `json:"hash"` // base64url(sha256), verified by the client
	Key           string `json:"key"`  // md5 hex, client cache key
	ContentType   string `json:"contentType"`
	FileExtension string `json:"fileExtension,omitempty"`
	URL           string `json:"url"`
}

// Manifest describes the update
type Manifest struct {
	ID             string         `json:"id"`
	CreatedAt      string         `json:"createdAt"`
	RuntimeVersion string         `json:"runtimeVersion"`
	LaunchAsset    Asset          `json:"launchAsset"`
	Assets         []Asset        `json:"assets"`
	Metadata       map[string]any `json:"metadata"`
	Extra          map[string]any `json:"extra"`
}

// Directive is sent instead of a manifest to instruct the client.
type Directive struct {
	Type       string         `json:"type"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

func NoUpdateAvailable() Directive {
	return Directive{Type: "noUpdateAvailable"}
}

func RollBackToEmbedded(commitTime time.Time) Directive {
	return Directive{
		Type:       "rollBackToEmbedded",
		Parameters: map[string]any{"commitTime": FormatTime(commitTime)},
	}
}

func FormatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
