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

func main() {
	mux := http.NewServeMux()

	mux.Handle("/auth/", http.StripPrefix("/auth", reverseProxy("http://auth:8080")))

	log.Println("API Gateway running on :8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
}

