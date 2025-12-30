// API proxy server that forwards requests with /api prefix. For testing with a
// local instance of Kion.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	// Parse command line flags
	targetPort := flag.String("target", "8081", "Target server port (default: 8081)")
	proxyPort := flag.String("port", "7979", "Proxy listen port (default: 7979)")
	flag.Parse()

	// Set up target URL
	targetURL := fmt.Sprintf("http://localhost:%s", *targetPort)
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom director to strip /api prefix
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Strip /api prefix from the path
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		log.Printf("Proxying: %s %s -> %s%s", req.Method, req.RequestURI, targetURL, req.URL.Path)
	}

	// Set up HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	// Start server
	listenAddr := fmt.Sprintf(":%s", *proxyPort)
	log.Printf("Starting API proxy on http://localhost%s", listenAddr)
	log.Printf("Forwarding requests to %s (stripping /api prefix)", targetURL)
	log.Printf("Example: http://localhost%s/api/v3/accounts -> %s/v3/accounts", *proxyPort, targetURL)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
