package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"username"`
	Password string `json:"password"`
}

type ValidateRequest struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid bool   `json:"valid"`
	Role  string `json:"role"`
}

var db *sql.DB
var jwtSecret []byte

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

func initDB() {
	var err error
	connStr := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	log.Println("Database connected")
}

func getUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func generateJWT(username, role string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (bool, string) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return false, ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if role, exists := claims["role"]; exists {
			if roleStr, ok := role.(string); ok {
				return true, roleStr
			}
		}
	}

	return true, ""
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := getUserByUsername(req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !checkPassword(user.Password, req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.Username, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	valid, role := validateJWT(req.Token)
	resp := ValidateResponse{Valid: valid, Role: role}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}


