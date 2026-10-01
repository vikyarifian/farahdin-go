// Package web embeds the static assets and builds cache-busted asset URLs.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"sort"
	"strings"
	"sync"
)

//go:embed all:static
var embedded embed.FS

// Static is the static asset tree, rooted at web/static.
var Static = mustSub(embedded, "static")

var (
	hashOnce sync.Once
	hashes   map[string]string
	version  string
)

func computeHashes() {
	hashes = map[string]string{}
	all := sha256.New()
	var paths []string
	fs.WalkDir(Static, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	for _, p := range paths {
		b, err := fs.ReadFile(Static, p)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(sum[:])[:10]
		all.Write(sum[:])
	}
	version = hex.EncodeToString(all.Sum(nil))[:12]
}

// Asset returns the URL of a static file with a content hash, e.g.
// Asset("css/app.css") == "/static/css/app.css?v=1a2b3c4d5e".
func Asset(path string) string {
	hashOnce.Do(computeHashes)
	path = strings.TrimPrefix(path, "/")
	if h, ok := hashes[path]; ok {
		return "/static/" + path + "?v=" + h
	}
	return "/static/" + path
}

// Version identifies the current static asset set (used by the service worker cache name).
func Version() string {
	hashOnce.Do(computeHashes)
	return version
}

func mustSub(f fs.FS, dir string) fs.FS {
	s, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return s
}
