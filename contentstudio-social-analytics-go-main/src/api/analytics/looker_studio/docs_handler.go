package looker_studio

import (
	_ "embed"
	"net/http"
)

//go:embed docs.html
var docsPageHTML []byte

// HandleDocs serves the Data Studio field reference documentation page.
// No authentication required — it is a public reference page.
func (h *Handler) HandleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(docsPageHTML) //nolint:errcheck
}
