package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/3mmarx/chirpy/internal/auth"
	"github.com/3mmarx/chirpy/internal/database"
)

func createUserHandler(db *database.Queries) http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// decode the request body
		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// validate the email
		req.Email = strings.TrimSpace(req.Email)
		parsed, err := mail.ParseAddress(req.Email)
		if err != nil || parsed.Address != req.Email {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}

		// generate password hash and insert user into db
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Couldn't create user", http.StatusInternalServerError)
			return
		}

		dbUser, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          req.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Couldn't create user", http.StatusInternalServerError)
			return
		}

		user := response{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Printf("Error encoding user response: %v", err)
		}
	}
}

func loginUser(apiConfig *apiConfig) http.HandlerFunc {
	type request struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"` // optional
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// decode the request body
		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// validate the email
		req.Email = strings.TrimSpace(req.Email)
		parsed, err := mail.ParseAddress(req.Email)
		if err != nil || parsed.Address != req.Email {
			http.Error(w, "Invalid email", http.StatusBadRequest)
			return
		}

		// Checking if user in db
		dbUser, err := apiConfig.db.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			log.Printf("User not found: %v", err)
			http.Error(w, "Couldn't login", http.StatusUnauthorized)
			return
		}

		// validate password
		match, err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword)
		if !match || err != nil {
			log.Printf("Invalid password: %v", err)
			http.Error(w, "Couldn't login", http.StatusUnauthorized)
			return
		}

		expiresInSeconds := 3600
		if req.ExpiresInSeconds != nil {
			if *req.ExpiresInSeconds <= 0 {
				log.Print("Invalid expiration time: ExpiresInSeconds should be more than 0")
				http.Error(w, "expires_in_seconds must be greater than 0", http.StatusBadRequest)
				return
			}

			expiresInSeconds = min(*req.ExpiresInSeconds, 3600)
		}

		token, err := auth.MakeJWT(dbUser.ID, apiConfig.jwtSecret, time.Duration(expiresInSeconds)*time.Second)
		if err != nil {
			log.Printf("failed to generate jwt token: %v", err)
			http.Error(w, "Couldn't login", http.StatusInternalServerError)
			return
		}

		//! I'VE USED THE SAME MESSAGE "Couldn't login" FOR SECURITY REASONS.

		user := response{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
			Token:     token,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Printf("Error encoding user response: %v", err)
		}
	}
}
