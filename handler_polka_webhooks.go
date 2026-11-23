package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type PolkaRequest struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}

type Data struct {
	UserID string `json:"user_id"`
}

func (cfg *apiConfig) PolkaWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	var polkaReq PolkaRequest
	err := json.NewDecoder(r.Body).Decode(&polkaReq)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if polkaReq.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(polkaReq.Data.UserID)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	_, err = cfg.db.UpgradeUserToChirpyRedById(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
