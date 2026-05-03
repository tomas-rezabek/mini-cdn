package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	
     http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		// get url from the query
		url := r.URL.Query().Get("url")

		if url == "" {
			http.Error(w, "Missing ?url", http.StatusBadRequest)
			return
		}

		// call origin server
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Failed to fetch origin", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// setup content-type
		w.Header().Set("Contenty-Type", resp.Header.Get("Content-Type"))

		// send data to the client
		io.Copy(w, resp.Body)
	})

	// run server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
