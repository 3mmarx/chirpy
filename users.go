package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/3mmarx/chirpy/internal/database"
)

func createUserHandler(db *database.Queries) http.HandlerFunc {
	type request struct {
		Email string `json:"email"`
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

		// insert user into db
		dbUser, err := db.CreateUser(r.Context(), req.Email)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't create user"))
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
