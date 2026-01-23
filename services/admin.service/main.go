package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Role string `json:"role"`
}

func initDB() {
	connStr := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully")
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	// Admin routes
	mux.HandleFunc("GET /users", authMiddleware(getUsersHandler))
	mux.HandleFunc("POST /users", authMiddleware(createUserHandler))
	mux.HandleFunc("PUT /users/", authMiddleware(editUserHandler))
	mux.HandleFunc("DELETE /users/", authMiddleware(deleteUserHandler))

	log.Println("Admin service running on :8002")
	log.Fatal(http.ListenAndServe(":8002", mux))
}
