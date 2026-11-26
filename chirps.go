package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/3mmarx/chirpy/internal/auth"
	"github.com/3mmarx/chirpy/internal/database"
	"github.com/google/uuid"
)

func getAllChirps(db *database.Queries) http.HandlerFunc {
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// return all chirps for an author
		stringAuthorID := r.URL.Query().Get("author_id")
		if stringAuthorID != "" {
			authorID, err := uuid.Parse(stringAuthorID)
			if err != nil {
				http.Error(w, "Invalid author id", http.StatusBadRequest)
				return
			}

			chirps, err := db.GetAllChirpsForAuthorById(r.Context(), authorID)
			if err != nil {
				http.Error(w, "Couldn't retrieve chirps", http.StatusInternalServerError)
				return
			}

			res := make([]response, 0, len(chirps))
			for _, chirp := range chirps {
				res = append(res, response{
					ID:        chirp.ID,
					CreatedAt: chirp.CreatedAt,
					UpdatedAt: chirp.UpdatedAt,
					Body:      chirp.Body,
					UserID:    chirp.UserID,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(res); err != nil {
				log.Printf("Error encoding chirps response: %v", err)
			}

			return
		}

		chirps, err := db.GetAllChirps(r.Context())
		if err != nil {
			http.Error(w, "Couldn't retrieve chirps", http.StatusInternalServerError)
			return
		}

		res := make([]response, 0, len(chirps))
		for _, chirp := range chirps {
			res = append(res, response{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}

		// send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding chirps response: %v", err)
		}
	}
}

func getChirpById(db *database.Queries) http.HandlerFunc {
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("chirpId")
		chirpId, err := uuid.Parse(path)
		if err != nil {
			http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
			return
		}

		chirp, err := db.GetChirpById(r.Context(), chirpId)
		if err != nil {
			http.Error(w, "Chirp not found", http.StatusNotFound)
			return
		}

		// we don't encode and return chirp directly because it returns keys in the uppercase.
		res := response{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding chirp response: %v", err)
		}
	}
}

func createChirpHandler(apiConfig *apiConfig) http.HandlerFunc {
	type request struct {
		Body string `json:"body"`
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// check if "Authorization" header exists and valid
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// validate the jwt token
		userID, err := auth.ValidateJWT(token, apiConfig.jwtSecret)
		if err != nil {
			http.Error(w, "Invalid jwt token", http.StatusUnauthorized)
			return
		}

		// decode request body
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// validate chirp
		validChirp, err := validateChirp(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// insert into db
		dbChirp, err := apiConfig.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body: validChirp, UserID: userID,
		})
		if err != nil {
			http.Error(w, "Couldn't create chirp", http.StatusInternalServerError)
			return
		}

		res := response{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("Error encoding chirp response: %v", err)
		}
	}
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	// get and validate access token
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// get chirpID, validate it and get it from db
	stringChirpID := r.PathValue("chirpID")
	if stringChirpID == "" {
		http.Error(w, "Supply a chirp id", http.StatusBadRequest)
		return
	}

	chirpID, err := uuid.Parse(stringChirpID)
	if err != nil {
		http.Error(w, "Invalid chirp id", http.StatusBadRequest)
		return
	}

	dbChirp, err := cfg.db.GetChirpById(r.Context(), chirpID)
	if err != nil {
		http.Error(w, "No chirp found", http.StatusNotFound)
		return
	}

	// check if the deleter is the author
	if userID != dbChirp.UserID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	err = cfg.db.DeleteChirpById(r.Context(), dbChirp.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateChirp(chirp string) (string, error) {
	if len(chirp) > 140 {
		return "", errors.New("Chirp is too long")
	}

	forbiddenWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(chirp)
	for i, word := range words {
		if _, ok := forbiddenWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	cleanedBody := strings.Join(words, " ")

	return cleanedBody, nil
}
