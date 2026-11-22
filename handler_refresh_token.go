package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/3mmarx/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Token string `json:"token"`
	}

	// get refresh token from Authorization header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Provide refresh token", http.StatusUnauthorized)
		return
	}

	// check if its in database
	// If it doesn't exist, or if it's expired or revoked, respond with a 401 status code
	dbUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "token expired or not found", http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// if it exists, respond with 200 and return  {"token":"access_token"}
	json.NewEncoder(w).Encode(request{
		Token: token,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// get refresh token from Authorization header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Provide refresh token", http.StatusUnauthorized)
		return
	}

	// check if its in database
	// If it doesn't exist respond with a 401 status code, else revoke it.
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		http.Error(w, "Refresh token doesn't exist", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
