package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"webhook-vault/backend/database"
	"webhook-vault/backend/middleware"
	"webhook-vault/backend/models"

	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if len(req.Username) < 3 || len(req.Password) < 6 {
		middleware.JSONError(w, "username min 3 chars, password min 6 chars", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.JSONError(w, "password hashing failure", http.StatusInternalServerError)
		return
	}

	res, err := database.DB.Exec(
		"INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'OPERATOR')",
		req.Username, req.Email, string(hash),
	)
	if err != nil {
		middleware.JSONError(w, "username or email already registered: "+err.Error(), http.StatusConflict)
		return
	}

	uid, _ := res.LastInsertId()
	token, _ := middleware.IssueToken(uid, req.Username, "OPERATOR")

	middleware.JSONSuccess(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User: models.User{
			ID:       uid,
			Username: req.Username,
			Email:    req.Email,
			Role:     "OPERATOR",
		},
	}, "operator account initialized")
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	ident := strings.TrimSpace(req.Username)
	var user models.User
	var hashed string

	err := database.DB.QueryRow(
		"SELECT id, username, email, password, role, created_at FROM users WHERE username = ? OR email = ?",
		ident, ident,
	).Scan(&user.ID, &user.Username, &user.Email, &hashed, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			middleware.JSONError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		middleware.JSONError(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(req.Password)); err != nil {
		middleware.JSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := middleware.IssueToken(user.ID, user.Username, user.Role)

	middleware.JSONSuccess(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	}, "authentication verified")
}

func Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetAuth(r)
	if !ok {
		middleware.JSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var u models.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, role, created_at FROM users WHERE id = ?",
		claims.UserID,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		middleware.JSONError(w, "user not found", http.StatusNotFound)
		return
	}

	middleware.JSONSuccess(w, http.StatusOK, u, "session valid")
}
