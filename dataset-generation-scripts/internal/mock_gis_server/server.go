package mock_gis_server

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Server replays captured GIS server responses over HTTP. It mimics the two
// endpoints used by the dataset generation pipeline and intentionally performs
// no authentication (local use only).
type Server struct {
	cache   *Cache
	verbose bool
}

// NewServer creates a mock GIS server backed by the given cache.
func NewServer(cache *Cache, verbose bool) *Server {
	return &Server{cache: cache, verbose: verbose}
}

// Handler returns the HTTP handler with both GIS routes registered.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/pread_json", func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, "id", s.cache.PreadJSON)
	})
	mux.HandleFunc("/p.co_dvhc_id", func(w http.ResponseWriter, r *http.Request) {
		s.handle(w, r, "malk", s.cache.CoDvhcID)
	})
	return mux
}

func (s *Server) handle(
	w http.ResponseWriter,
	r *http.Request,
	formField string,
	lookup func(string) (string, bool),
) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed; use POST")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form body: "+err.Error())
		return
	}

	key := r.FormValue(formField)
	if key == "" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("missing form field %q", formField))
		return
	}

	// Only keys present in the index are served, so the request value is never
	// joined to a filesystem path (no path traversal).
	path, ok := lookup(key)
	if !ok {
		s.logf("%s %s [%s] -> 404", r.Method, r.URL.Path, key)
		writeError(w, http.StatusNotFound, fmt.Sprintf("no cached response for %q", key))
		return
	}

	body, err := readMaybeGzip(path)
	if err != nil {
		s.logf("%s %s [%s] -> 500 (%v)", r.Method, r.URL.Path, key, err)
		writeError(w, http.StatusInternalServerError, "failed to read cached response: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		s.logf("%s %s [%s] -> write error: %v", r.Method, r.URL.Path, key, err)
		return
	}
	s.logf("%s %s [%s] -> 200 (%d bytes)", r.Method, r.URL.Path, key, len(body))
}

// readMaybeGzip reads a cached response, transparently decompressing `.gz`.
func readMaybeGzip(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		return io.ReadAll(gz)
	}
	return io.ReadAll(f)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, "{\"error\":%q}\n", message)
}

func (s *Server) logf(format string, args ...any) {
	if s.verbose {
		log.Printf(format, args...)
	}
}
