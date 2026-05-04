package main

import (
	"fmt"
	"io"
	"net/http"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

const cacheDir = "cache"

func cacheKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func cachePath(url string) string {
	return filepath.Join(cacheDir, cacheKey(url))
}

func main() {
	
     http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		// get url from the query
		url := r.URL.Query().Get("url")

		if url == "" {
			http.Error(w, "Missing ?url", http.StatusBadRequest)
			return
		}

		os.MkdirAll(cacheDir, 0755)
		filePath := cachePath(url)

		// CACHE HIT
		if _, err := os.Stat(filePath); err == nil {
			w.Header().Set("X-Cache", "HIT")

			file, err := os.Open(filePath)
			if err != nil {
				http.Error(w, "Failed to read cache", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			io.Copy(w, file)
			return
		}

		// CACHE MISS
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Failed to fetch origin", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		cacheFile, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to create cache file", http.StatusInternalServerError)
			return
		}
		defer cacheFile.Close()

		// setup cache header
		w.Header().Set("X-Cache", "MISS")
		// setup content-type
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))

		// sending data to cache and response
		multiWriter := io.MultiWriter(w, cacheFile)
		io.Copy(multiWriter, resp.Body)
	})

	// run server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
