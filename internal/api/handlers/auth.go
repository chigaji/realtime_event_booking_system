package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/chigaji/realtime_event_booking_system/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validator.Validate(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var storedUser models.User
	err := h.db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", user.Username).Scan(&storedUser.ID, &storedUser.Username, &storedUser.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate and return JWT token here

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// decode the body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//valid the data
	if err := validator.Validate(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// See if user is allready saved in the db
	var existingUser models.User
	err := h.db.QueryRow("SELECT id FROM users WHERE username = $1 OR email = $2", user.Username, user.Email).Scan(&existingUser.ID)
	if err == nil {
		http.Error(w, "Username or Email Already Exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// save user to the database
	_, err = h.db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", user.Username, user.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}
