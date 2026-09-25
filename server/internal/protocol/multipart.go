package protocol

import (
	"mime/multipart"
	"net/http"
	"net/textproto"
)

// WriteManifest sends a manifest as a multipart/mixed response.

// body must be the already-encoded manifest JSON: the exact bytes that were signed
// signature is the expo-signature header value, or "" when not signing.
func WriteManifest(w http.ResponseWriter, body []byte, signature string) error {
	mw := multipart.NewWriter(w)
	writeHeaders(w, mw)

	if err := writePart(mw, "manifest", "application/json; charset=utf-8", body, signature); err != nil {
		return err
	}

	extensions := []byte(`{"assetRequestHeaders":{}}`)
	if err := writePart(mw, "extensions", "application/json", extensions, ""); err != nil {
		return err
	}

	return mw.Close()
}

func WriteDirective(w http.ResponseWriter, body []byte, signature string) error {
	mw := multipart.NewWriter(w)
	writeHeaders(w, mw)

	if err := writePart(mw, "directive", "application/json; charset=utf-8", body, signature); err != nil {
		return err
	}

	return mw.Close()
}

func writeHeaders(w http.ResponseWriter, mw *multipart.Writer) {
	h := w.Header()
	h.Set("expo-protocol-version", "1")
	h.Set("expo-sfv-version", "0")
	h.Set("cache-control", "private, max-age=0")
	h.Set("content-type", "multipart/mixed; boundary="+mw.Boundary())
}

func writePart(mw *multipart.Writer, name, contentType string, body []byte, signature string) error {
	h := textproto.MIMEHeader{}
	h.Set("content-disposition", `form-data; name="`+name+`"`)
	h.Set("content-type", contentType)
	if signature != "" {
		h.Set("expo-signature", signature)
	}

	part, err := mw.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = part.Write(body)
	return err
}
