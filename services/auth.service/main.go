package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type User_Auth struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordhash"`
	Role         string `json:"role"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Role string `json:"role"`
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	//Authentication
	mux.HandleFunc("/login", loginHandler)
	//Authorization
	mux.HandleFunc("/validate", validateHandler)

	//REMEMBER TO REMOVE
	// mux.HandleFunc("/getAIModels", getAIModelsHandler)
	mux.HandleFunc("GET /users", authMiddleware(getUsersHandler))
	mux.HandleFunc("POST /users", authMiddleware(createUserHandler))
	mux.HandleFunc("PUT /users/", authMiddleware(editUserHandler))
	mux.HandleFunc("DELETE /users/", authMiddleware(deleteUserHandler))

	// Get port from environment or default to 8001
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	log.Println("Auth service running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
