package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func reverseProxy(target string) http.Handler {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	return proxy
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	//Routes to Authentication service
	mux.Handle("/auth/", http.StripPrefix("/auth", reverseProxy("http://auth:8000")))

	log.Println("API Gateway running on :8000")
	log.Fatal(http.ListenAndServe(":8000", corsMiddleware(mux)))
}
