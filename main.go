package main

import (
	"fmt"
	"net/http"
)

func main() {
	
     http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Mini CDN is running...")
	})

	// run server on port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
