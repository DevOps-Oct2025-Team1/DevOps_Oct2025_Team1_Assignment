package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type APIResponse struct {
	Data any `json:"data"`
}

type GetModelsData struct {
	Models []string `json:"models"`
}

var db *sql.DB

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/models", getModelsHandler)

	log.Println("Prompt Manager service running on :8001")
	log.Fatal(http.ListenAndServe(":8001", mux))
}
