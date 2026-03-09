package main

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/pgomes13/drift-guard-engine/pkg/compare"
)

//go:embed static
var staticFiles embed.FS

// diffCache caches serialised diff results keyed by SHA-256 of the request
// inputs. It clears itself when it reaches maxSize to bound memory usage.
type diffCache struct {
	mu      sync.RWMutex
	entries map[string][]byte
	maxSize int
}

func newDiffCache(maxSize int) *diffCache {
	return &diffCache{entries: make(map[string][]byte, maxSize), maxSize: maxSize}
}

func (c *diffCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.entries[key]
	return v, ok
}

func (c *diffCache) set(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.maxSize {
		c.entries = make(map[string][]byte, c.maxSize)
	}
	c.entries[key] = val
}

func cacheKey(schemaType, base, head string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s", schemaType, base, head)
	return fmt.Sprintf("%x", h.Sum(nil))
}

var cache = newDiffCache(256)

type compareRequest struct {
	SchemaType  string `json:"schema_type"`  // "openapi" | "graphql" | "grpc"
	BaseContent string `json:"base_content"`
	HeadContent string `json:"head_content"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}

	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to sub static FS: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/compare", corsMiddleware(compareHandler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	log.Printf("drift-guard playground API listening on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2<<20) // 2 MB

	var req compareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.BaseContent == "" || req.HeadContent == "" {
		writeError(w, http.StatusBadRequest, "base_content and head_content are required")
		return
	}

	key := cacheKey(req.SchemaType, req.BaseContent, req.HeadContent)
	if cached, ok := cache.get(key); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(cached)
		return
	}

	ext := extForType(req.SchemaType)

	basePath, err := writeTempFile([]byte(req.BaseContent), "base"+ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write base content: "+err.Error())
		return
	}
	defer os.Remove(basePath)

	headPath, err := writeTempFile([]byte(req.HeadContent), "head"+ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write head content: "+err.Error())
		return
	}
	defer os.Remove(headPath)

	result, err := runCompare(req.SchemaType, basePath, headPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode result: "+err.Error())
		return
	}
	cache.set(key, encoded)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(encoded)
}

func runCompare(schemaType, basePath, headPath string) (any, error) {
	switch schemaType {
	case "graphql":
		return compare.GraphQL(basePath, headPath)
	case "grpc":
		return compare.GRPC(basePath, headPath)
	default: // openapi
		return compare.OpenAPI(basePath, headPath)
	}
}

func extForType(schemaType string) string {
	switch schemaType {
	case "graphql":
		return ".graphql"
	case "grpc":
		return ".proto"
	default:
		return ".yaml"
	}
}

func writeTempFile(content []byte, name string) (string, error) {
	ext := filepath.Ext(name)
	f, err := os.CreateTemp("", "drift-guard-*"+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
