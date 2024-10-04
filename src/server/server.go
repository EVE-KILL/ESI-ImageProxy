package server

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/eve-kill/esi-imageproxy/endpoints"
	"github.com/eve-kill/esi-imageproxy/helpers"
	"github.com/eve-kill/esi-imageproxy/proxy"
)

// setupServer initializes the HTTP server with routes and handlers.
func setupServer() *http.ServeMux {
	// Create new router
	mux := http.NewServeMux()

	// Register health and ping endpoints
	mux.HandleFunc("/ping", endpoints.Ping)
	mux.HandleFunc("/healthz", endpoints.Healthz) // Liveness probe
	mux.HandleFunc("/readyz", endpoints.Readyz)   // Readiness probe

	// Handle .well-known requests by dropping them
	mux.HandleFunc("/.well-known/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Handle robots.txt
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("User-agent: *\nDisallow: /"))
		if err != nil {
			log.Printf("Error writing robots.txt: %v", err)
		}
	})

	// Set up proxy for all other routes
	upstreamURL, err := url.Parse("https://images.evetech.net/")
	if err != nil {
		log.Fatalf("Error parsing upstream URL: %v", err)
	}

	// Initialize the Cache
	cache := helpers.NewCache(1*time.Hour, 10*time.Minute)

	proxyHandler := proxy.NewProxy(upstreamURL, cache)

	// Register the proxy handler
	mux.Handle("/", proxy.HandleRequest(proxyHandler, cache))

	return mux
}

// StartServer starts the HTTP server.
func StartServer() {
	// Set up the server components
	mux := setupServer()

	// Start server
	host := helpers.GetEnv("HOST", "0.0.0.0")
	port := helpers.GetEnv("PORT", "8080")

	log.Println("Proxy server started on http://" + host + ":" + port)
	log.Fatal(http.ListenAndServe(host+":"+port, mux))
}
