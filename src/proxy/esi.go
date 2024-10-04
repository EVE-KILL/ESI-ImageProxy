package proxy

import (
	"bytes"
	"image"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/eve-kill/esi-imageproxy/helpers"
)

// NewProxy creates a new reverse proxy for the given target URL.
func NewProxy(targetURL *url.URL, cache *helpers.Cache) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the Director function to change the Host header
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host

		// Set headers to mimic a browser request
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Connection", "keep-alive")
	}

	// Configure the Transport to use keep-alive settings
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 90 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return proxy
}

// HandleRequest processes incoming requests, handles caching, and serves optimized images.
func HandleRequest(proxy *httputil.ReverseProxy, cache *helpers.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cacheKey := helpers.GenerateCacheKey(r.URL.Path)

		// Determine preferred format based on Accept header
		acceptHeader := r.Header.Get("Accept")
		preferredFormat := "jpeg" // Default format
		if strings.Contains(acceptHeader, "image/webp") {
			preferredFormat = "webp"
		} else if strings.Contains(acceptHeader, "image/png") {
			preferredFormat = "png"
		} else if strings.Contains(acceptHeader, "image/jpeg") || strings.Contains(acceptHeader, "image/jpg") {
			preferredFormat = "jpeg"
		}

		preferredCacheKey := cacheKey + "-" + preferredFormat

		// Check if the optimized image is in the cache
		if _, found := cache.Get(preferredCacheKey); found {
			helpers.ServeOptimizedImage(w, r, cache, cacheKey)
			return
		}

		// Indicate a cache miss
		w.Header().Set("X-Proxy-Cache", "MISS")

		// Capture the response from the upstream server
		recorder := &responseRecorder{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}
		proxy.ServeHTTP(recorder, r)

		// Check if the response status is OK
		if recorder.status != http.StatusOK {
			return
		}

		// Process and cache the image
		originalImage, _, err := image.Decode(recorder.body)
		if err != nil {
			log.Printf("Error decoding image: %v", err)
			http.Error(w, "Failed to decode image", http.StatusInternalServerError)
			return
		}

		// Optimize and cache images in different formats
		helpers.CacheOptimizedImages(cache, cacheKey, originalImage, recorder.Header())

		// Serve the optimized image from cache
		helpers.ServeOptimizedImage(w, r, cache, cacheKey)
	}
}

// responseRecorder is a custom response writer to capture the response for caching.
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

// Header returns the header map that will be sent by WriteHeader.
func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body and delegates to the underlying ResponseWriter.
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
