package helpers

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chai2010/webp"
)

// CacheOptimizedImages optimizes and caches images in WebP, PNG, and JPEG formats.
func CacheOptimizedImages(cache *Cache, cacheKey string, img image.Image, headers http.Header) {
	safeHeaders := sanitizeHeaders(headers)

	// Optimize and cache WebP
	webpBuffer := new(bytes.Buffer)
	if err := webp.Encode(webpBuffer, img, &webp.Options{Lossless: false, Quality: 80}); err != nil {
		log.Printf("Error encoding WebP: %v", err)
	} else {
		cache.Set(cacheKey+"-webp", CacheItem{
			Headers: safeHeaders,
			Status:  http.StatusOK,
			Body:    webpBuffer.Bytes(),
		}, 1*time.Hour)
	}

	// Optimize and cache PNG
	pngBuffer := new(bytes.Buffer)
	if err := png.Encode(pngBuffer, img); err != nil {
		log.Printf("Error encoding PNG: %v", err)
	} else {
		cache.Set(cacheKey+"-png", CacheItem{
			Headers: safeHeaders,
			Status:  http.StatusOK,
			Body:    pngBuffer.Bytes(),
		}, 1*time.Hour)
	}

	// Optimize and cache JPEG
	jpegBuffer := new(bytes.Buffer)
	if err := jpeg.Encode(jpegBuffer, img, &jpeg.Options{Quality: 80}); err != nil {
		log.Printf("Error encoding JPEG: %v", err)
	} else {
		cache.Set(cacheKey+"-jpeg", CacheItem{
			Headers: safeHeaders,
			Status:  http.StatusOK,
			Body:    jpegBuffer.Bytes(),
		}, 1*time.Hour)
	}
}

// ServeOptimizedImage serves the optimized image based on the cache and client's Accept header.
func ServeOptimizedImage(w http.ResponseWriter, r *http.Request, cache *Cache, cacheKey string) {
	acceptHeader := r.Header.Get("Accept")

	// Define preferred formats in order of priority
	preferredFormats := []struct {
		Type string
		Key  string
	}{
		{"image/webp", "-webp"},
		{"image/png", "-png"},
		{"image/jpeg", "-jpeg"},
		{"image/jpg", "-jpeg"},
	}

	for _, pf := range preferredFormats {
		if strings.Contains(acceptHeader, pf.Type) {
			if cachedItem, found := cache.Get(cacheKey + pf.Key); found {
				// Pass through cached headers except Content-Type, X-Proxy-Cache, and Content-Length
				for key, values := range cachedItem.Headers {
					if key == "Content-Type" || key == "X-Proxy-Cache" || key == "Content-Length" {
						continue
					}
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}

				// Set the modified headers
				w.Header().Set("Content-Type", pf.Type)
				w.Header().Set("X-Proxy-Cache", "HIT")
				// Let Go set Content-Length automatically

				// Write the image body
				_, err := w.Write(cachedItem.Body)
				if err != nil {
					log.Printf("Error writing optimized image: %v", err)
				}
				return
			}
		}
	}

	// Default to JPEG if no specific format is requested or found
	if cachedItem, found := cache.Get(cacheKey + "-jpeg"); found {
		// Pass through cached headers except Content-Type, X-Proxy-Cache, and Content-Length
		for key, values := range cachedItem.Headers {
			if key == "Content-Type" || key == "X-Proxy-Cache" || key == "Content-Length" {
				continue
			}
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Set the modified headers
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("X-Proxy-Cache", "HIT")
		// Let Go set Content-Length automatically

		// Write the image body
		_, err := w.Write(cachedItem.Body)
		if err != nil {
			log.Printf("Error writing optimized JPEG image: %v", err)
		}
		return
	}

	// If no optimized image is found, return an error
	http.Error(w, "No optimized image found", http.StatusNotFound)
}

// sanitizeHeaders removes headers that should not be cached.
func sanitizeHeaders(headers http.Header) http.Header {
	safeHeaders := headers.Clone()
	safeHeaders.Del("X-Proxy-Cache")
	safeHeaders.Del("Content-Length")
	// Add more headers to remove if necessary
	return safeHeaders
}

// GetContentType maps format to MIME type.
func GetContentType(format string) string {
	switch format {
	case "webp":
		return "image/webp"
	case "png":
		return "image/png"
	case "jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}
