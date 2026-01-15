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

	mux.Handle("/auth/", http.StripPrefix("/auth", reverseProxy("http://127.0.0.3:8080")))

	log.Println("API Gateway running on :8000")
	log.Fatal(http.ListenAndServe("127.0.0.2:8000", mux))
}

