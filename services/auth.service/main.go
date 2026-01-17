package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()
	//Authentication
	mux.HandleFunc("/login", loginHandler)
	//Authorization
	mux.HandleFunc("/validate", validateHandler)

	log.Println("Auth service running on :8080")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
