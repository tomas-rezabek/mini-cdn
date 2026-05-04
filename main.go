package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)


const cacheDir = "cache"
const cacheTTL = 1 * time.Minute // 60 seconds cache expiration
const originTimeout = 30 * time.Second

var httpClient = &http.Client{
	Timeout: originTimeout,
}

func cacheKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func cachePath(url string) string {
	return filepath.Join(cacheDir, cacheKey(url))
}

func main() {
	os.MkdirAll(cacheDir, 0755)
	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		// get url from the query
		url := r.URL.Query().Get("url")

		if url == "" {
			http.Error(w, "Missing ?url", http.StatusBadRequest)
			return
		}

		filePath := cachePath(url)

		// CACHE HIT
		if fileInfo, err := os.Stat(filePath); err == nil {

			// debug
			cacheAge := time.Since(fileInfo.ModTime())
			fmt.Println("Cache age:", cacheAge)

			if cacheAge <= cacheTTL {

				w.Header().Set("X-Cache", "HIT")
				fmt.Println("CACHE HIT")

				file, err := os.Open(filePath)
				if err != nil {
					http.Error(w, "Failed to read cache", http.StatusInternalServerError)
					return
				}
				defer file.Close()

				io.Copy(w, file)
				return
			}
			fmt.Println("CACHE EXPIRED")

		}

		// CACHE MISS
		resp, err := httpClient.Get(url)
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
		fmt.Println("CACHE MISS")
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
