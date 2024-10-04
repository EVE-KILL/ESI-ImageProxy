package endpoints

import (
	"io"
	"net/http"
)

func Root(w http.ResponseWriter, r *http.Request) {
	// Fetch content from the specified URL
	resp, err := http.Get("https://docs.esi.evetech.net/docs/image_server.html")
	if err != nil {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}

	// Copy the content to the response writer
	w.Header().Set("Content-Type", "text/html")
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Failed to write content", http.StatusInternalServerError)
		return
	}
}
